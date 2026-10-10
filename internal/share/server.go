package share

import (
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

//go:embed web/*
var web embed.FS

const idleDuration = 30 * time.Minute
const absoluteDuration = 8 * time.Hour
const attemptWindow = 15 * time.Minute

type session struct{ idle, until time.Time }
type attempt struct {
	count int
	until time.Time
}
type App struct {
	config       Config
	certificates Certificates
	root         *os.Root
	network      *net.IPNet
	mu           sync.Mutex
	sessions     map[[32]byte]session
	attempts     map[string]attempt
	logins       chan struct{}
}

func NewApp(c Config, certificates Certificates) (*App, error) {
	root, err := os.OpenRoot(c.Folder)
	if err != nil {
		return nil, errors.New("cannot open the shared folder")
	}
	ip := net.ParseIP(c.Bind).To4()
	mask := net.CIDRMask(c.Prefix, 32)
	return &App{config: c, certificates: certificates, root: root, network: &net.IPNet{IP: ip.Mask(mask), Mask: mask}, sessions: make(map[[32]byte]session), attempts: make(map[string]attempt), logins: make(chan struct{}, 2)}, nil
}

func (a *App) Close() error    { return a.root.Close() }
func (a *App) Address() string { return net.JoinHostPort(a.config.Bind, strconv.Itoa(a.config.Port)) }
func (a *App) Origin() string  { return "https://" + a.Address() }

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
}

func reply(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func deny(w http.ResponseWriter, status int, message string) {
	reply(w, status, map[string]string{"error": message})
}

func asset(w http.ResponseWriter, r *http.Request, name string) {
	data, err := web.ReadFile("web/" + name)
	if err != nil {
		deny(w, 500, "Asset unavailable.")
		return
	}
	contentType := map[string]string{"index.html": "text/html", "app.js": "text/javascript", "style.css": "text/css", "dashboard.html": "text/html", "dashboard.js": "text/javascript"}[name]
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)
	ipText, _, err := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(ipText)
	if err != nil || ip == nil || !a.network.Contains(ip) {
		deny(w, 403, "Only the selected local network can access this app.")
		return
	}
	if r.Host != a.Address() {
		deny(w, 403, "Invalid host.")
		return
	}
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "POST" {
		deny(w, 405, "Method not allowed.")
		return
	}
	read := r.Method == "GET" || r.Method == "HEAD"
	route := r.URL.Path
	public := read && (route == "/" || route == "/app.js" || route == "/style.css" || route == "/install.mobileconfig" || route == "/certificate.cer")
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" && !public {
		deny(w, 403, "Cross-site requests are blocked.")
		return
	}
	if r.Method == "POST" && r.Header.Get("Origin") != a.Origin() {
		deny(w, 403, "Invalid request origin.")
		return
	}
	if public {
		switch route {
		case "/":
			asset(w, r, "index.html")
		case "/app.js":
			asset(w, r, "app.js")
		case "/style.css":
			asset(w, r, "style.css")
		case "/install.mobileconfig":
			w.Header().Set("Content-Type", "application/x-apple-aspen-config")
			w.Header().Set("Content-Disposition", "attachment; filename=PC-to-iPhone-Share.mobileconfig")
			if r.Method != "HEAD" {
				_, _ = w.Write(a.certificates.Profile)
			}
		case "/certificate.cer":
			w.Header().Set("Content-Type", "application/pkix-cert")
			w.Header().Set("Content-Disposition", "attachment; filename=PC-to-iPhone-Share.cer")
			if r.Method != "HEAD" {
				_, _ = w.Write(a.certificates.DER)
			}
		}
		return
	}
	if route == "/api/login" && r.Method == "POST" {
		a.login(w, r, ipText)
		return
	}
	key, ok := a.authenticate(r)
	if !ok {
		deny(w, 401, "Please sign in.")
		return
	}
	switch {
	case route == "/api/logout" && r.Method == "POST":
		a.mu.Lock()
		delete(a.sessions, key)
		a.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "__Host-PCToIPhoneShare", Value: "", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		reply(w, 200, map[string]bool{"ok": true})
	case route == "/api/list" && r.Method == "GET":
		a.list(w, r)
	case route == "/download" && read:
		a.download(w, r)
	default:
		deny(w, 404, "Not found.")
	}
}

func (a *App) allowAttempt(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for key, value := range a.attempts {
		if !now.Before(value.until) {
			delete(a.attempts, key)
		}
	}
	for key, value := range a.sessions {
		if !now.Before(value.idle) || !now.Before(value.until) {
			delete(a.sessions, key)
		}
	}
	if len(a.attempts) >= 1024 {
		if _, exists := a.attempts[ip]; !exists {
			return false
		}
	}
	for _, key := range []string{ip, "all-peers"} {
		value := a.attempts[key]
		limit := 5
		if key == "all-peers" {
			limit = 30
		}
		if now.Before(value.until) && value.count >= limit {
			return false
		}
	}
	for _, key := range []string{ip, "all-peers"} {
		value := a.attempts[key]
		if !now.Before(value.until) {
			value = attempt{until: now.Add(attemptWindow)}
		}
		value.count++
		a.attempts[key] = value
	}
	return true
}

func (a *App) login(w http.ResponseWriter, r *http.Request, ip string) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		deny(w, 415, "Expected JSON.")
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2048))
	if err != nil {
		deny(w, 413, "Request too large.")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if json.Unmarshal(data, &body) != nil || len(body.Password) > 512 {
		deny(w, 400, "Enter your password.")
		return
	}
	select {
	case a.logins <- struct{}{}:
		defer func() { <-a.logins }()
	default:
		w.Header().Set("Retry-After", "15")
		deny(w, 429, "Login is busy. Try again shortly.")
		return
	}
	if !a.allowAttempt(ip) {
		w.Header().Set("Retry-After", "900")
		deny(w, 429, "Too many login attempts. Try again in 15 minutes.")
		return
	}
	salt, _ := hex.DecodeString(a.config.Salt)
	expected, _ := hex.DecodeString(a.config.PasswordHash)
	actual, err := derive(body.Password, salt, a.config.Iterations)
	if err != nil || subtle.ConstantTimeCompare(actual, expected) != 1 {
		deny(w, 401, "Incorrect password.")
		return
	}
	token := randomToken(32)
	key := sha256.Sum256([]byte(token))
	a.mu.Lock()
	if len(a.sessions) >= 256 {
		a.mu.Unlock()
		deny(w, 429, "Too many active sessions.")
		return
	}
	now := time.Now()
	a.sessions[key] = session{idle: now.Add(idleDuration), until: now.Add(absoluteDuration)}
	a.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "__Host-PCToIPhoneShare", Value: token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: int(absoluteDuration.Seconds())})
	reply(w, 200, map[string]bool{"ok": true})
}

func (a *App) authenticate(r *http.Request) ([32]byte, bool) {
	var key [32]byte
	cookie, err := r.Cookie("__Host-PCToIPhoneShare")
	if err != nil || len(cookie.Value) != 43 {
		return key, false
	}
	key = sha256.Sum256([]byte(cookie.Value))
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[key]
	now := time.Now()
	if !ok || !now.Before(s.idle) || !now.Before(s.until) {
		delete(a.sessions, key)
		return key, false
	}
	s.idle = now.Add(idleDuration)
	a.sessions[key] = s
	return key, true
}

func safeRelative(name string) (string, error) {
	if len(name) > 2048 || strings.ContainsAny(name, "\\:") {
		return "", errors.New("invalid path")
	}
	if name == "" {
		return ".", nil
	}
	for _, c := range name {
		if unicode.IsControl(c) {
			return "", errors.New("invalid path")
		}
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return "", errors.New("invalid path")
		}
		base := strings.ToUpper(strings.Split(part, ".")[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return "", errors.New("invalid path")
		}
	}
	if !fs.ValidPath(name) {
		return "", errors.New("invalid path")
	}
	return filepath.FromSlash(name), nil
}

func (a *App) open(name string) (*os.File, error) {
	relative, err := safeRelative(name)
	if err != nil {
		return nil, err
	}
	// Reject links for clarity; os.Root also enforces containment when directories
	// are renamed or swapped between this check and the open operation.
	current := ""
	if relative != "." {
		for _, component := range strings.Split(relative, string(filepath.Separator)) {
			current = filepath.Join(current, component)
			info, err := a.root.Lstat(current)
			if err != nil {
				return nil, err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return nil, errors.New("linked files are not shared")
			}
		}
	}
	return a.root.Open(relative)
}

type Entry struct {
	Name      string    `json:"name"`
	Directory bool      `json:"directory"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
}

func (a *App) list(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("path")
	file, err := a.open(name)
	if err != nil {
		deny(w, 400, "Folder is unavailable or its path is blocked.")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.IsDir() {
		deny(w, 400, "This is not a folder.")
		return
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 || offset > 10000 {
		deny(w, 400, "Invalid page.")
		return
	}
	items, err := file.ReadDir(10001)
	if err != nil && err != io.EOF {
		deny(w, 403, "Cannot read this folder.")
		return
	}
	truncated := len(items) > 10000
	if truncated {
		items = items[:10000]
	}
	entries := make([]Entry, 0, len(items))
	for _, item := range items {
		child, err := a.open(path.Join(name, item.Name()))
		if err != nil {
			continue
		}
		stat, err := child.Stat()
		child.Close()
		if err != nil || (!stat.Mode().IsRegular() && !stat.IsDir()) {
			continue
		}
		entries = append(entries, Entry{item.Name(), stat.IsDir(), stat.Size(), stat.ModTime()})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Directory != entries[j].Directory {
			return entries[i].Directory
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	total := len(entries)
	if offset > total {
		offset = total
	}
	end := offset + 200
	if end > total {
		end = total
	}
	reply(w, 200, map[string]any{"name": filepath.Base(a.config.Folder), "path": name, "entries": entries[offset:end], "total": total, "nextOffset": end, "hasMore": end < total, "truncated": truncated})
}

func (a *App) download(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("path")
	file, err := a.open(name)
	if err != nil {
		deny(w, 400, "File is unavailable or its path is blocked.")
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		deny(w, 400, "Only regular files can be downloaded.")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(name)}))
	http.ServeContent(w, r, path.Base(name), stat.ModTime(), file)
}

// A connection cap covers TLS handshakes and slow clients without buffering file downloads.
type cappedListener struct {
	net.Listener
	slots chan struct{}
}
type cappedConn struct {
	net.Conn
	slots chan struct{}
	once  sync.Once
}

func (l *cappedListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	select {
	case l.slots <- struct{}{}:
		return &cappedConn{Conn: c, slots: l.slots}, nil
	default:
		c.Close()
		return nil, temporaryBusy{}
	}
}
func (c *cappedConn) Close() error {
	err := c.Conn.Close()
	c.once.Do(func() { <-c.slots })
	return err
}

type temporaryBusy struct{}

func (temporaryBusy) Error() string   { return "connection limit reached" }
func (temporaryBusy) Temporary() bool { return true }
func (temporaryBusy) Timeout() bool   { return false }

func httpServer(handler http.Handler) *http.Server {
	return &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192, ErrorLog: log.New(io.Discard, "", 0)}
}
