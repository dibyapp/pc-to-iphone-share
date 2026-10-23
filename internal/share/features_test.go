package share

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func decodeResponse(t *testing.T, r *http.Response, target any) {
	t.Helper()
	defer r.Body.Close()
	expectStatus(t, r, 200)
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}

type listResult struct {
	Entries           []Entry
	Total, NextOffset int
	HasMore           bool
}

func TestSearchFilterSortAndPageAcrossFullFolder(t *testing.T) {
	f := setupFixture(t)
	for i := 0; i < 205; i++ {
		writeTestFile(t, filepath.Join(f.root, "item-"+strings.Repeat("x", i+1)+".txt"), []byte("text"))
	}
	writeTestFile(t, filepath.Join(f.root, "z-final.JPG"), []byte("image"))
	writeTestFile(t, filepath.Join(f.root, "A-note.txt"), []byte("small"))
	writeTestFile(t, filepath.Join(f.root, "b-note.txt"), []byte("much larger content"))
	if err := os.Mkdir(filepath.Join(f.root, "z-folder"), 0700); err != nil {
		t.Fatal(err)
	}
	f.login(t)
	var result listResult
	decodeResponse(t, f.request(t, "GET", "/api/list?q=FINAL&type=images&limit=1", nil, nil), &result)
	if len(result.Entries) != 1 || result.Entries[0].Name != "z-final.JPG" || result.Total != 1 {
		t.Fatal("search did not inspect entries beyond the first page")
	}
	decodeResponse(t, f.request(t, "GET", "/api/list?q=note&sort=size&direction=desc&limit=1", nil, nil), &result)
	if result.Entries[0].Name != "b-note.txt" || result.Total != 2 || !result.HasMore || result.NextOffset != 1 {
		t.Fatal("size sorting or page size failed")
	}
	decodeResponse(t, f.request(t, "GET", "/api/list?q=note&sort=size&direction=desc&limit=1&offset=1", nil, nil), &result)
	if result.Entries[0].Name != "A-note.txt" || result.HasMore {
		t.Fatal("second filtered page incorrect")
	}
	old := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Chtimes(filepath.Join(f.root, "A-note.txt"), old, old)
	decodeResponse(t, f.request(t, "GET", "/api/list?q=note&sort=modified", nil, nil), &result)
	if result.Entries[0].Name != "A-note.txt" {
		t.Fatal("modified sorting failed")
	}
	decodeResponse(t, f.request(t, "GET", "/api/list?type=folders", nil, nil), &result)
	if result.Total != 1 || !result.Entries[0].Directory {
		t.Fatal("folder filter failed")
	}
	decodeResponse(t, f.request(t, "GET", "/api/list?sort=name&direction=desc&limit=500", nil, nil), &result)
	if !result.Entries[0].Directory || len(result.Entries) != 209 {
		t.Fatal("folders must stay first in both sort directions")
	}
	for _, query := range []string{"offset=no", "offset=-1", "limit=0", "limit=501", "limit=no", "sort=unknown", "direction=no", "type=unknown", "q=" + strings.Repeat("a", 201)} {
		expectStatus(t, f.request(t, "GET", "/api/list?"+query, nil, nil), 400)
	}
}

func TestZIPStreamsExactBytesAndRejectsUnsafeSelections(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	os.Mkdir(filepath.Join(f.root, "nested"), 0700)
	files := map[string][]byte{"zero.txt": {}, "nested/notes.txt": []byte("<script>literal only</script>"), "image.bin": bytes.Repeat([]byte{0, 1, 255}, 40000)}
	params := url.Values{}
	for name, data := range files {
		writeTestFile(t, filepath.Join(f.root, filepath.FromSlash(name)), data)
		params.Add("path", name)
	}
	r := f.request(t, "GET", "/download.zip?"+params.Encode(), nil, nil)
	expectStatus(t, r, 200)
	data, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.File) != len(files) {
		t.Fatal("wrong ZIP count")
	}
	for _, member := range archive.File {
		stream, err := member.Open()
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(stream)
		stream.Close()
		if err != nil || !bytes.Equal(got, files[member.Name]) {
			t.Fatal("ZIP bytes or CRC mismatch")
		}
		if member.Method != zip.Store {
			t.Fatal("ZIP should avoid recompressing media")
		}
	}
	for _, query := range []string{"", "path=../outside", "path=zero.txt&path=zero.txt", "path=nested", "path=.hidden", "path=missing.txt"} {
		expectStatus(t, f.request(t, "GET", "/download.zip?"+query, nil, nil), 400)
	}
	params = url.Values{}
	for i := 0; i < 65; i++ {
		params.Add("path", "zero.txt")
	}
	expectStatus(t, f.request(t, "GET", "/download.zip?"+params.Encode(), nil, nil), 400)
	// HEAD must not create a download or send an archive body.
	before := f.app.statistics()["completed"]
	r = f.request(t, "HEAD", "/download.zip?path=zero.txt", nil, nil)
	expectStatus(t, r, 200)
	data, _ = io.ReadAll(r.Body)
	if len(data) != 0 || f.app.statistics()["completed"] != before {
		t.Fatal("HEAD counted as download")
	}
}

func TestPreviewBoundariesAndChecksumIntegrity(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	body := []byte("<img src=x onerror=alert(1)>\nUnicode: 日本語")
	writeTestFile(t, filepath.Join(f.root, "notes.txt"), body)
	writeTestFile(t, filepath.Join(f.root, "binary.txt"), []byte{0, 1, 255})
	writeTestFile(t, filepath.Join(f.root, "page.html"), body)
	writeTestFile(t, filepath.Join(f.root, "large.txt"), []byte(strings.Repeat("€", 100000)))
	var preview struct {
		Text      string
		Truncated bool
	}
	decodeResponse(t, f.request(t, "GET", "/api/preview?path=notes.txt", nil, nil), &preview)
	if preview.Text != string(body) || preview.Truncated {
		t.Fatal("preview changed text")
	}
	decodeResponse(t, f.request(t, "GET", "/api/preview?path=large.txt", nil, nil), &preview)
	if len(preview.Text) > 256*1024 || !preview.Truncated || !strings.HasSuffix(preview.Text, "€") {
		t.Fatal("preview UTF-8 limit broken")
	}
	for _, name := range []string{"binary.txt", "page.html"} {
		expectStatus(t, f.request(t, "GET", "/api/preview?path="+name, nil, nil), 415)
	}
	for _, endpoint := range []string{"preview", "checksum"} {
		expectStatus(t, f.request(t, "GET", "/api/"+endpoint+"?path=../outside.txt", nil, nil), 400)
	}
	var checksum struct {
		SHA256 string
		Size   int64
	}
	decodeResponse(t, f.request(t, "GET", "/api/checksum?path=notes.txt", nil, nil), &checksum)
	expected := sha256.Sum256(body)
	if checksum.SHA256 != hex.EncodeToString(expected[:]) || checksum.Size != int64(len(body)) {
		t.Fatal("checksum mismatch")
	}
}

func TestNewEndpointsRequireSessionAndSameSite(t *testing.T) {
	f := setupFixture(t)
	for _, route := range []string{"/download.zip?path=x.txt", "/api/preview?path=x.txt", "/api/checksum?path=x.txt", "/api/speed-test"} {
		expectStatus(t, f.request(t, "GET", route, nil, nil), 401)
		expectStatus(t, f.request(t, "GET", route, nil, map[string]string{"Sec-Fetch-Site": "cross-site"}), 403)
	}
}

func TestExclusionsApplyToAllFileEndpointsAtEveryDepth(t *testing.T) {
	f := setupFixture(t)
	f.app.options.Exclude = []string{"*.key", "private"}
	f.login(t)
	os.Mkdir(filepath.Join(f.root, "PRIVATE"), 0700)
	writeTestFile(t, filepath.Join(f.root, "PRIVATE", "notes.txt"), []byte("synthetic private text"))
	writeTestFile(t, filepath.Join(f.root, "sample.KEY"), []byte("synthetic fixture"))
	writeTestFile(t, filepath.Join(f.root, "public.txt"), []byte("public"))
	var result listResult
	decodeResponse(t, f.request(t, "GET", "/api/list", nil, nil), &result)
	if result.Total != 1 || result.Entries[0].Name != "public.txt" {
		t.Fatal("excluded entries leaked in listing")
	}
	for _, route := range []string{"/download?path=sample.KEY", "/download?path=PRIVATE/notes.txt", "/api/list?path=PRIVATE", "/download.zip?path=PRIVATE/notes.txt", "/api/preview?path=PRIVATE/notes.txt", "/api/checksum?path=PRIVATE/notes.txt"} {
		expectStatus(t, f.request(t, "GET", route, nil, nil), 400)
	}
}

func dashboardAction(d *Dashboard, route, method string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://"+d.Host+route, nil)
	r.Host = d.Host
	r.RemoteAddr = "127.0.0.1:1"
	r.Header.Set("Authorization", "Bearer "+d.Token)
	if method == "POST" {
		r.Header.Set("Origin", "http://"+d.Host)
	}
	w := httptest.NewRecorder()
	d.ServeHTTP(w, r)
	return w
}
func TestDashboardPauseResumeRevokeAndPrivateStatistics(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	writeTestFile(t, filepath.Join(f.root, "hello.txt"), []byte("hello"))
	d := &Dashboard{App: f.app, Token: randomToken(32), Host: "127.0.0.1:19444", Stop: make(chan struct{})}
	if dashboardAction(d, "/api/pause", "POST").Code != 200 {
		t.Fatal("pause failed")
	}
	expectStatus(t, f.request(t, "GET", "/download?path=hello.txt", nil, nil), 503)
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 503)
	if dashboardAction(d, "/api/resume", "POST").Code != 200 {
		t.Fatal("resume failed")
	}
	r := f.request(t, "GET", "/download?path=hello.txt", nil, nil)
	expectStatus(t, r, 200)
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	stats := dashboardAction(d, "/api/stats", "GET")
	if stats.Code != 200 {
		t.Fatal("statistics failed")
	}
	var s struct{ Completed, Bytes, Sessions, Active int }
	json.Unmarshal(stats.Body.Bytes(), &s)
	if s.Completed != 1 || s.Bytes != 5 || s.Sessions != 1 || s.Active != 0 {
		t.Fatalf("incorrect counters: %+v", s)
	}
	for _, private := range []string{f.root, f.password, "hello.txt", d.Token} {
		if bytes.Contains(stats.Body.Bytes(), []byte(private)) {
			t.Fatal("statistics include private details")
		}
	}
	if dashboardAction(d, "/api/revoke", "POST").Code != 200 {
		t.Fatal("revoke failed")
	}
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 401)
	f.login(t)
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 200)
	for _, route := range []string{"/api/pause", "/api/resume", "/api/revoke"} {
		r := httptest.NewRequest("POST", "http://"+d.Host+route, nil)
		r.Host = d.Host
		r.RemoteAddr = "127.0.0.1:1"
		r.Header.Set("Authorization", "Bearer "+d.Token)
		r.Header.Set("Origin", "http://example.invalid")
		w := httptest.NewRecorder()
		d.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal("dashboard mutation accepted cross-site request")
		}
	}
}

func TestExpiryAndOperationLimits(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	writeTestFile(t, filepath.Join(f.root, "notes.txt"), []byte("hello"))
	// Occupy every operation slot without slow network timing assumptions.
	for i := 0; i < cap(f.app.transfers); i++ {
		f.app.transfers <- struct{}{}
	}
	for _, route := range []string{"/download?path=notes.txt", "/download.zip?path=notes.txt", "/api/checksum?path=notes.txt", "/api/speed-test"} {
		expectStatus(t, f.request(t, "GET", route, nil, nil), 429)
	}
	for len(f.app.transfers) > 0 {
		<-f.app.transfers
	}
	f.app.expires = time.Now().Add(-time.Second)
	for _, route := range []string{"/api/list", "/download?path=notes.txt", "/api/speed-test"} {
		expectStatus(t, f.request(t, "GET", route, nil, nil), 410)
	}
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, nil), 410)
}

func TestSpeedTestIsBoundedRateLimitedAndExcludedFromFileStats(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	r := f.request(t, "GET", "/api/speed-test", nil, nil)
	expectStatus(t, r, 200)
	n, err := io.Copy(io.Discard, r.Body)
	r.Body.Close()
	if err != nil || n != 16<<20 {
		t.Fatal("incorrect synthetic test size")
	}
	expectStatus(t, f.request(t, "GET", "/api/speed-test", nil, nil), 429)
	stats := f.app.statistics()
	if stats["completed"].(uint64) != 0 || stats["bytes"].(uint64) != 0 || stats["active"].(int) != 0 {
		t.Fatal("synthetic bytes counted as file transfers")
	}
}

func TestRuntimeOptionValidationAndSetupCheck(t *testing.T) {
	for _, options := range []Options{{MaxDownloads: 0}, {MaxDownloads: 17}, {MaxDownloads: 1, ShareFor: -time.Second}, {MaxDownloads: 1, ShareFor: 25 * time.Hour}, {MaxDownloads: 1, Exclude: []string{"../private"}}, {MaxDownloads: 1, Exclude: []string{"["}}} {
		if options.validate() == nil {
			t.Fatal("invalid runtime option accepted")
		}
	}
	temp := t.TempDir()
	folder := filepath.Join(temp, "files")
	state := filepath.Join(temp, "settings")
	os.Mkdir(folder, 0700)
	var output bytes.Buffer
	if err := Run([]string{"--folder", folder, "--state", state, "--loopback-only", "--check", "--exclude", "*.key", "--share-for", "5m", "--max-downloads", "2"}, "test", strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "No servers started") || strings.Contains(output.String(), "password:") || strings.Contains(output.String(), "Local dashboard:") {
		t.Fatal("setup check exposed runtime credentials or launched servers")
	}
	if _, err := os.Stat(filepath.Join(state, "config.json")); !os.IsNotExist(err) {
		t.Fatal("setup check saved new password")
	}
	if _, err := os.Stat(filepath.Join(state, "ca-key.pem")); !os.IsNotExist(err) {
		t.Fatal("setup check generated certificate material")
	}
}

func TestCanceledZIPReleasesSlotWithoutCountingSuccess(t *testing.T) {
	f := setupFixture(t)
	writeTestFile(t, filepath.Join(f.root, "example.txt"), bytes.Repeat([]byte("synthetic"), 8192))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest("GET", f.app.Origin()+"/download.zip?path=example.txt", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	func() {
		defer func() {
			if recovered := recover(); recovered != http.ErrAbortHandler {
				t.Fatal("canceled ZIP must abort its HTTP stream")
			}
		}()
		f.app.downloadZIP(w, r)
	}()
	stats := f.app.statistics()
	if stats["active"].(int) != 0 || stats["completed"].(uint64) != 0 || len(f.app.transfers) != 0 {
		t.Fatal("canceled ZIP leaked an operation slot or counted success")
	}
}

func TestSortTiesAreDeterministicAndFoldersStayFirst(t *testing.T) {
	entries := []Entry{{Name: "b.txt", Size: 10}, {Name: "a.txt", Size: 10}, {Name: "folder", Directory: true}}
	sortEntries(entries, "size", false)
	if entries[0].Name != "folder" || entries[1].Name != "a.txt" {
		t.Fatal("size tie did not fall back to name")
	}
	sortEntries(entries, "size", true)
	if entries[0].Name != "folder" || entries[1].Name != "b.txt" {
		t.Fatal("descending tie or directory ordering changed")
	}
}
