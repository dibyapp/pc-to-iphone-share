package share

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Options apply to one run only; no filenames or transfer history are persisted.
type Options struct {
	MaxDownloads int
	ShareFor     time.Duration
	Exclude      []string
}

func (o Options) validate() error {
	if o.MaxDownloads < 1 || o.MaxDownloads > 16 {
		return errors.New("max-downloads must be between 1 and 16")
	}
	if o.ShareFor < 0 || o.ShareFor > 24*time.Hour {
		return errors.New("share-for must be between zero and 24h")
	}
	if len(o.Exclude) > 32 {
		return errors.New("at most 32 exclusion patterns are allowed")
	}
	for _, pattern := range o.Exclude {
		if pattern == "" || len(pattern) > 128 || strings.ContainsAny(pattern, "/\\:") || strings.IndexFunc(pattern, unicode.IsControl) >= 0 {
			return errors.New("exclusions must be filename patterns without path separators")
		}
		if _, err := path.Match(pattern, "example"); err != nil {
			return errors.New("invalid exclusion pattern")
		}
	}
	return nil
}

func (a *App) excluded(name string) bool {
	for _, component := range strings.Split(name, "/") {
		for _, pattern := range a.options.Exclude {
			if match, _ := path.Match(strings.ToLower(pattern), strings.ToLower(component)); match {
				return true
			}
		}
	}
	return false
}

func listOptions(r *http.Request) (offset, limit int, query, kind, order string, descending bool, err error) {
	q := r.URL.Query()
	limit = 200
	if s := q.Get("offset"); s != "" {
		offset, err = strconv.Atoi(s)
		if err != nil {
			return
		}
	}
	if s := q.Get("limit"); s != "" {
		limit, err = strconv.Atoi(s)
		if err != nil {
			return
		}
	}
	query = strings.TrimSpace(q.Get("q"))
	kind = q.Get("type")
	order = q.Get("sort")
	descending = q.Get("direction") == "desc"
	if offset < 0 || offset > 10000 || limit < 1 || limit > 500 || len(query) > 200 || strings.IndexFunc(query, unicode.IsControl) >= 0 {
		err = errors.New("invalid list options")
		return
	}
	if !oneOf(kind, "", "all", "folders", "images", "videos", "audio", "documents", "archives", "other") || !oneOf(order, "", "name", "size", "modified") || !oneOf(q.Get("direction"), "", "asc", "desc") {
		err = errors.New("invalid filter or sort")
	}
	return
}
func oneOf(value string, allowed ...string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	return false
}
func fileKind(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".heic", ".avif", ".bmp", ".tiff":
		return "images"
	case ".mp4", ".mov", ".mkv", ".webm", ".avi", ".m4v":
		return "videos"
	case ".mp3", ".wav", ".flac", ".m4a", ".aac", ".ogg":
		return "audio"
	case ".txt", ".md", ".csv", ".json", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".rtf", ".log", ".yaml", ".yml", ".xml":
		return "documents"
	case ".zip", ".gz", ".tar", ".7z", ".rar", ".xz", ".bz2":
		return "archives"
	default:
		return "other"
	}
}
func filterEntries(entries []Entry, query, kind string) []Entry {
	result := make([]Entry, 0, len(entries))
	query = strings.ToLower(query)
	for _, e := range entries {
		if !strings.Contains(strings.ToLower(e.Name), query) {
			continue
		}
		if kind == "folders" && !e.Directory {
			continue
		}
		// Keep matching folders navigable when filtering file types.
		if kind != "" && kind != "all" && kind != "folders" && !e.Directory && fileKind(e.Name) != kind {
			continue
		}
		result = append(result, e)
	}
	return result
}
func sortEntries(entries []Entry, order string, desc bool) {
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Directory != b.Directory {
			return a.Directory
		}
		cmp := 0
		if order == "size" && a.Size != b.Size {
			if a.Size < b.Size {
				cmp = -1
			} else {
				cmp = 1
			}
		}
		if order == "modified" && !a.Modified.Equal(b.Modified) {
			if a.Modified.Before(b.Modified) {
				cmp = -1
			} else {
				cmp = 1
			}
		}
		if cmp == 0 {
			cmp = strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			if cmp == 0 {
				cmp = strings.Compare(a.Name, b.Name)
			}
		}
		if desc {
			return cmp > 0
		}
		return cmp < 0
	})
}

type transferWriter struct {
	http.ResponseWriter
	bytes  uint64
	status int
	failed bool
}

func (w *transferWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
		w.ResponseWriter.WriteHeader(code)
	}
}
func (w *transferWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += uint64(n)
	if err != nil || n != len(p) {
		w.failed = true
	}
	return n, err
}
func (w *transferWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (a *App) beginTransfer(w http.ResponseWriter, r *http.Request) (func(uint64, bool), bool) {
	a.mu.Lock()
	if a.paused || (!a.expires.IsZero() && !time.Now().Before(a.expires)) {
		a.mu.Unlock()
		deny(w, 503, "Sharing is paused or expired.")
		return nil, false
	}
	select {
	case a.transfers <- struct{}{}:
	default:
		a.mu.Unlock()
		w.Header().Set("Retry-After", "3")
		deny(w, 429, "Transfer limit reached. Try again shortly.")
		return nil, false
	}
	a.active++
	a.mu.Unlock()
	return func(bytes uint64, ok bool) {
		<-a.transfers
		a.mu.Lock()
		defer a.mu.Unlock()
		a.active--
		if r.Method != "HEAD" {
			a.sent += bytes
			if ok && r.Context().Err() == nil {
				a.completed++
			}
		}
	}, true
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func (a *App) downloadZIP(w http.ResponseWriter, r *http.Request) {
	names := r.URL.Query()["path"]
	if len(names) == 0 || len(names) > 64 || len(r.URL.RawQuery) > 16384 {
		deny(w, 400, "Choose 1 to 64 regular files for a ZIP.")
		return
	}
	type member struct {
		file     *os.File
		name     string
		size     int64
		modified time.Time
	}
	members := make([]member, 0, len(names))
	defer func() {
		for _, m := range members {
			m.file.Close()
		}
	}()
	seen := map[string]bool{}
	var total int64
	for _, name := range names {
		if seen[strings.ToLower(name)] {
			deny(w, 400, "Duplicate ZIP paths are not allowed.")
			return
		}
		seen[strings.ToLower(name)] = true
		f, err := a.open(name)
		if err != nil {
			deny(w, 400, "A selected file is unavailable or blocked.")
			return
		}
		info, err := f.Stat()
		if err != nil || !info.Mode().IsRegular() {
			f.Close()
			deny(w, 400, "ZIP selection must contain regular files only.")
			return
		}
		members = append(members, member{f, name, info.Size(), info.ModTime()})
		if info.Size() > (4<<30)-total {
			deny(w, 413, "ZIP selection exceeds 4 GiB. Download large files separately.")
			return
		}
		total += info.Size()
	}
	finish, ok := a.beginTransfer(w, r)
	if !ok {
		return
	}
	counter := &transferWriter{ResponseWriter: w}
	success := false
	defer func() { finish(counter.bytes, success) }()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=shared-files.zip")
	if r.Method == "HEAD" {
		return
	}
	// Stored entries avoid spending CPU recompressing photos and videos. Stream directly to the browser.
	archive := zip.NewWriter(counter)
	for _, m := range members {
		h := &zip.FileHeader{Name: m.name, Method: zip.Store}
		h.SetMode(0600)
		entry, err := archive.CreateHeader(h)
		if err != nil {
			panic(http.ErrAbortHandler)
		}
		n, err := io.CopyN(entry, contextReader{r.Context(), m.file}, m.size)
		info, statErr := m.file.Stat()
		if err != nil || n != m.size || statErr != nil || info.Size() != m.size || !info.ModTime().Equal(m.modified) {
			panic(http.ErrAbortHandler)
		}
	}
	if err := archive.Close(); err != nil {
		panic(http.ErrAbortHandler)
	}
	success = true
}

func (a *App) readRegular(w http.ResponseWriter, r *http.Request) (*os.File, os.FileInfo, bool) {
	f, err := a.open(r.URL.Query().Get("path"))
	if err != nil {
		deny(w, 400, "File is unavailable or blocked.")
		return nil, nil, false
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		deny(w, 400, "Choose a regular file.")
		return nil, nil, false
	}
	return f, info, true
}
func (a *App) preview(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("path")
	if !oneOf(strings.ToLower(path.Ext(name)), ".txt", ".md", ".csv", ".json", ".log", ".yaml", ".yml", ".xml") {
		deny(w, 415, "Preview supports UTF-8 text documents only.")
		return
	}
	f, _, ok := a.readRegular(w, r)
	if !ok {
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(contextReader{r.Context(), f}, 256*1024+1))
	if err != nil {
		deny(w, 409, "Cannot read this file.")
		return
	}
	truncated := len(data) > 256*1024
	if truncated {
		data = data[:256*1024]
		for i := 0; i < 3 && len(data) > 0 && !utf8.Valid(data); i++ {
			data = data[:len(data)-1]
		}
	}
	if !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
		deny(w, 415, "This file is not UTF-8 text.")
		return
	}
	reply(w, 200, map[string]any{"text": string(data), "truncated": truncated})
}
func (a *App) checksum(w http.ResponseWriter, r *http.Request) {
	f, info, ok := a.readRegular(w, r)
	if !ok {
		return
	}
	defer f.Close()
	if info.Size() > 1<<30 {
		deny(w, 413, "Browser checksums are limited to 1 GiB. Use your computer for larger files.")
		return
	}
	finish, ok := a.beginTransfer(w, r)
	if !ok {
		return
	}
	defer finish(0, false)
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	hash := sha256.New()
	n, err := io.Copy(hash, contextReader{ctx, io.LimitReader(f, info.Size()+1)})
	after, statErr := f.Stat()
	if err != nil || statErr != nil || n != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		deny(w, 409, "File changed or checksum timed out. Retry when the file is stable.")
		return
	}
	reply(w, 200, map[string]any{"sha256": hex.EncodeToString(hash.Sum(nil)), "size": n})
}
func (a *App) speedTest(w http.ResponseWriter, r *http.Request) {
	finish, ok := a.beginTransfer(w, r)
	if !ok {
		return
	}
	defer finish(0, false)
	a.mu.Lock()
	limited := time.Now().Before(a.speedUntil)
	if !limited && r.Method != "HEAD" {
		a.speedUntil = time.Now().Add(30 * time.Second)
	}
	a.mu.Unlock()
	if limited {
		w.Header().Set("Retry-After", "30")
		deny(w, 429, "Wait 30 seconds between speed tests.")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(16<<20))
	if r.Method == "HEAD" {
		return
	}
	chunk := randomBytes(64 << 10)
	for i := 0; i < 256; i++ {
		if r.Context().Err() != nil {
			return
		}
		if _, err := w.Write(chunk); err != nil {
			return
		}
	}
}

func (a *App) statistics() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	sessions := 0
	for _, s := range a.sessions {
		if now.Before(s.idle) && now.Before(s.until) {
			sessions++
		}
	}
	return map[string]any{"paused": a.paused, "active": a.active, "completed": a.completed, "bytes": a.sent, "sessions": sessions, "startedAt": a.started, "expiresAt": a.expires, "maxDownloads": a.options.MaxDownloads}
}
