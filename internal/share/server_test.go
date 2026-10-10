package share

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type fixture struct {
	app         *App
	server      *httptest.Server
	client      *http.Client
	password    string
	root, state string
}

func setupFixture(t *testing.T) fixture {
	t.Helper()
	temp := t.TempDir()
	root := filepath.Join(temp, "share")
	state := filepath.Join(temp, "state")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(state, 0700); err != nil {
		t.Fatal(err)
	}
	c := Config{Folder: root, Bind: "127.0.0.1", Prefix: 32, Port: 19443}
	password, err := newPassword(&c)
	if err != nil {
		t.Fatal(err)
	}
	certs, err := ensureCertificates(state, c.Bind)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateConfig(c, state, true); err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(c, certs)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(app)
	_, port, _ := net.SplitHostPort(server.Listener.Addr().String())
	app.config.Port, _ = strconv.Atoi(port)
	server.TLS = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{certs.TLS}}
	server.StartTLS()
	pool := x509.NewCertPool()
	ca, err := x509.ParseCertificate(certs.DER)
	if err != nil {
		t.Fatal(err)
	}
	pool.AddCert(ca)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}}, Timeout: 5 * time.Second}
	t.Cleanup(func() { client.CloseIdleConnections(); server.Close(); app.Close() })
	return fixture{app, server, client, password, root, state}
}

func (f fixture) request(t *testing.T, method, route string, body any, headers map[string]string) *http.Response {
	t.Helper()
	var data io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		data = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, f.app.Origin()+route, data)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", f.app.Origin())
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res, err := f.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}
func expectStatus(t *testing.T, r *http.Response, status int) {
	t.Helper()
	if r.StatusCode != status {
		data, _ := io.ReadAll(r.Body)
		t.Fatalf("status %d, want %d: %s", r.StatusCode, status, data)
	}
}
func (f fixture) login(t *testing.T) {
	t.Helper()
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, nil), 200)
}
func writeTestFile(t *testing.T, name string, content []byte) {
	t.Helper()
	if err := os.WriteFile(name, content, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPSAuthenticationAndHeaders(t *testing.T) {
	f := setupFixture(t)
	page := f.request(t, "GET", "/", nil, nil)
	expectStatus(t, page, 200)
	if page.TLS == nil || len(page.TLS.VerifiedChains) == 0 {
		t.Fatal("certificate was not verified")
	}
	if page.Header.Get("X-Content-Type-Options") != "nosniff" || page.Header.Get("X-Frame-Options") != "DENY" || page.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("security headers missing")
	}
	for _, route := range []string{"/api/list", "/download?path=file.txt"} {
		expectStatus(t, f.request(t, "GET", route, nil, nil), 401)
	}
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": "wrong"}, nil), 401)
	login := f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, nil)
	expectStatus(t, login, 200)
	cookie := login.Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
		t.Fatal("cookie not protected")
	}
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 200)
	expectStatus(t, f.request(t, "POST", "/api/logout", map[string]string{}, nil), 200)
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 401)
}

func TestSafariCrossSiteEntryAndCSRF(t *testing.T) {
	f := setupFixture(t)
	headers := map[string]string{"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "navigate", "Sec-Fetch-Dest": "document"}
	for _, route := range []string{"/", "/app.js", "/style.css", "/install.mobileconfig", "/certificate.cer"} {
		expectStatus(t, f.request(t, "GET", route, nil, headers), 200)
	}
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, headers), 403)
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, map[string]string{"Origin": "https://example.invalid"}), 403)
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, map[string]string{"Origin": ""}), 403)
	f.login(t)
	for _, route := range []string{"/api/list", "/download?path=file.txt"} {
		expectStatus(t, f.request(t, "GET", route, nil, headers), 403)
	}
	expectStatus(t, f.request(t, "POST", "/api/logout", map[string]string{}, headers), 403)
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 200)
}

func TestFolderIsolationAndDownloadIntegrity(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	payload := bytes.Repeat([]byte("download-test-data-"), 65536)
	writeTestFile(t, filepath.Join(f.root, "file.bin"), payload)
	writeTestFile(t, filepath.Join(f.root, "empty.txt"), nil)
	writeTestFile(t, filepath.Join(f.root, "emoji-🙂.txt"), []byte("Unicode"))
	writeTestFile(t, filepath.Join(f.root, ".hidden.txt"), []byte("hidden"))
	for _, name := range []string{"../state/ca.key", "..\\state\\ca.key", "file.bin:stream", ".hidden.txt", "file.bin.", "CON", "nested/../file.bin"} {
		expectStatus(t, f.request(t, "GET", "/download?path="+url.QueryEscape(name), nil, nil), 400)
	}
	listing := f.request(t, "GET", "/api/list", nil, nil)
	data, _ := io.ReadAll(listing.Body)
	if bytes.Contains(data, []byte(".hidden.txt")) {
		t.Fatal("dot file listed")
	}
	full := f.request(t, "GET", "/download?path=file.bin", nil, nil)
	expectStatus(t, full, 200)
	actual, _ := io.ReadAll(full.Body)
	if !bytes.Equal(actual, payload) {
		t.Fatal("file changed during streaming")
	}
	if !strings.HasPrefix(full.Header.Get("Content-Disposition"), "attachment") || full.Header.Get("Content-Type") != "application/octet-stream" {
		t.Fatal("unsafe download headers")
	}
	partial := f.request(t, "GET", "/download?path=file.bin", nil, map[string]string{"Range": "bytes=10-99"})
	expectStatus(t, partial, 206)
	data, _ = io.ReadAll(partial.Body)
	if !bytes.Equal(data, payload[10:100]) {
		t.Fatal("invalid range content")
	}
	suffix := f.request(t, "GET", "/download?path=file.bin", nil, map[string]string{"Range": "bytes=-19"})
	expectStatus(t, suffix, 206)
	data, _ = io.ReadAll(suffix.Body)
	if !bytes.Equal(data, payload[len(payload)-19:]) {
		t.Fatal("invalid suffix range")
	}
	expectStatus(t, f.request(t, "GET", "/download?path=file.bin", nil, map[string]string{"Range": "bytes=20-1"}), 416)
	head := f.request(t, "HEAD", "/download?path=file.bin", nil, nil)
	expectStatus(t, head, 200)
	if head.ContentLength != int64(len(payload)) {
		t.Fatal("HEAD size mismatch")
	}
	expectStatus(t, f.request(t, "GET", "/download?path=empty.txt", nil, nil), 200)
	unicode := f.request(t, "GET", "/download?path="+url.QueryEscape("emoji-🙂.txt"), nil, nil)
	expectStatus(t, unicode, 200)
	if !strings.Contains(unicode.Header.Get("Content-Disposition"), "filename*") {
		t.Fatal("Unicode attachment filename missing")
	}
}

func TestExternalLinkAndRootContainment(t *testing.T) {
	f := setupFixture(t)
	outside := filepath.Join(t.TempDir(), "outside")
	os.Mkdir(outside, 0700)
	writeTestFile(t, filepath.Join(outside, "secret.txt"), []byte("outside"))
	createDirectoryLink(t, filepath.Join(f.root, "escape"), outside)
	f.login(t)
	expectStatus(t, f.request(t, "GET", "/download?path=escape/secret.txt", nil, nil), 400)
	if _, err := f.app.root.Open(filepath.Join("escape", "secret.txt")); err == nil {
		t.Fatal("os.Root allowed escape")
	}
}

func TestSessionExpiryAndThrottling(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	f.app.mu.Lock()
	for key, value := range f.app.sessions {
		value.idle = time.Now().Add(-time.Second)
		f.app.sessions[key] = value
	}
	f.app.mu.Unlock()
	expectStatus(t, f.request(t, "GET", "/api/list", nil, nil), 401)
	for i := 0; i < 4; i++ {
		expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": "wrong"}, nil), 401)
	}
	last := f.request(t, "POST", "/api/login", map[string]string{"password": f.password}, nil)
	expectStatus(t, last, 429)
	if last.Header.Get("Retry-After") != "900" {
		t.Fatal("retry header missing")
	}
}

func TestPaginationAndConcurrentDownloads(t *testing.T) {
	f := setupFixture(t)
	f.login(t)
	for i := 0; i < 205; i++ {
		writeTestFile(t, filepath.Join(f.root, "file-"+strconv.Itoa(i)+".txt"), []byte("page"))
	}
	var first, second struct {
		Entries    []Entry
		Total      int
		HasMore    bool
		NextOffset int
	}
	r := f.request(t, "GET", "/api/list", nil, nil)
	json.NewDecoder(r.Body).Decode(&first)
	r = f.request(t, "GET", "/api/list?offset=200", nil, nil)
	json.NewDecoder(r.Body).Decode(&second)
	if len(first.Entries) != 200 || !first.HasMore || first.Total != 205 || len(second.Entries) != 5 || second.HasMore {
		t.Fatal("pagination failed")
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := f.request(t, "GET", "/download?path=file-0.txt", nil, nil)
			expectStatus(t, r, 200)
			io.Copy(io.Discard, r.Body)
		}()
	}
	wg.Wait()
}

func TestNetworkAndHostBoundaries(t *testing.T) {
	f := setupFixture(t)
	request := httptest.NewRequest("GET", f.app.Origin()+"/", nil)
	request.Host = f.app.Address()
	request.RemoteAddr = "203.0.113.8:12345"
	response := httptest.NewRecorder()
	f.app.ServeHTTP(response, request)
	if response.Code != 403 {
		t.Fatal("outside subnet allowed")
	}
	request.RemoteAddr = "127.0.0.1:12345"
	request.Host = "example.invalid"
	response = httptest.NewRecorder()
	f.app.ServeHTTP(response, request)
	if response.Code != 403 {
		t.Fatal("host spoofing allowed")
	}
	expectStatus(t, f.request(t, "DELETE", "/api/list", nil, nil), 405)
	expectStatus(t, f.request(t, "POST", "/api/login", map[string]string{"password": strings.Repeat("x", 3000)}, nil), 413)
}

func TestCertificatePersistenceAndPrivateConfig(t *testing.T) {
	f := setupFixture(t)
	next, err := ensureCertificates(f.state, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(next.DER, f.app.certificates.DER) {
		t.Fatal("root CA changed across restart")
	}
	if bytes.Contains(next.Profile, []byte("PRIVATE KEY")) {
		t.Fatal("private key in public profile")
	}
	if err := saveConfig(f.state, f.app.config); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(f.state, "config.json"))
	if bytes.Contains(data, []byte(f.password)) {
		t.Fatal("plaintext password persisted")
	}
	c := f.app.config
	c.Folder = filepath.Dir(f.state)
	if validateConfig(c, f.state, true) == nil {
		t.Fatal("state directory could be shared")
	}
	c = f.app.config
	c.Bind = "0.0.0.0"
	if validateConfig(c, f.state, true) == nil {
		t.Fatal("all-interface binding accepted")
	}
}

func TestDashboardLocalAuthentication(t *testing.T) {
	f := setupFixture(t)
	d := &Dashboard{App: f.app, Token: randomToken(32), Password: f.password, Host: "127.0.0.1:19444", Stop: make(chan struct{})}
	ask := func(route, remote, token, origin, method string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://"+d.Host+route, nil)
		r.Host = d.Host
		r.RemoteAddr = remote
		r.Header.Set("Authorization", token)
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		d.ServeHTTP(w, r)
		return w
	}
	if ask("/api/state", "127.0.0.1:1", "", "", "GET").Code != 401 {
		t.Fatal("unauthenticated dashboard exposed")
	}
	if ask("/api/state", "203.0.113.8:1", "Bearer "+d.Token, "", "GET").Code != 403 {
		t.Fatal("remote dashboard exposed")
	}
	if ask("/api/state", "127.0.0.1:1", "Bearer "+d.Token, "", "GET").Code != 200 {
		t.Fatal("local dashboard rejected")
	}
	if ask("/api/stop", "127.0.0.1:1", "Bearer "+d.Token, "http://example.invalid", "POST").Code != 403 {
		t.Fatal("dashboard CSRF accepted")
	}
	if ask("/api/stop", "127.0.0.1:1", "Bearer "+d.Token, "http://"+d.Host, "POST").Code != 200 {
		t.Fatal("stop failed")
	}
	select {
	case <-d.Stop:
	default:
		t.Fatal("stop channel still open")
	}
}
