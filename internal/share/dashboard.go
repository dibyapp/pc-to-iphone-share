package share

import (
	"crypto/subtle"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
)

type Dashboard struct {
	App                            *App
	Password, Token, Host, Version string
	Stop                           chan struct{}
	once                           sync.Once
}

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	securityHeaders(w)
	address, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip := net.ParseIP(address); ip == nil || !ip.IsLoopback() || r.Host != d.Host {
		deny(w, 403, "Local dashboard only.")
		return
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		deny(w, 403, "Open the dashboard directly from this computer.")
		return
	}
	if r.Method == "GET" && r.URL.Path == "/" {
		asset(w, r, "dashboard.html")
		return
	}
	if r.Method == "GET" && r.URL.Path == "/dashboard.js" {
		asset(w, r, "dashboard.js")
		return
	}
	if r.Method == "GET" && r.URL.Path == "/style.css" {
		asset(w, r, "style.css")
		return
	}
	token := r.Header.Get("Authorization")
	if subtle.ConstantTimeCompare([]byte(token), []byte("Bearer "+d.Token)) != 1 {
		deny(w, 401, "Open the dashboard URL printed in the app console.")
		return
	}
	if r.Method == "POST" && r.Header.Get("Origin") != "http://"+d.Host {
		deny(w, 403, "Invalid dashboard origin.")
		return
	}
	switch {
	case r.Method == "GET" && r.URL.Path == "/api/state":
		reply(w, 200, map[string]any{"url": d.App.Origin(), "folder": filepath.Base(d.App.config.Folder), "password": d.Password, "fingerprint": d.App.certificates.Fingerprint, "version": d.Version, "profileUrl": d.App.Origin() + "/install.mobileconfig", "port": strconv.Itoa(d.App.config.Port)})
	case r.Method == "GET" && r.URL.Path == "/api/stats":
		reply(w, 200, d.App.statistics())
	case r.Method == "POST" && (r.URL.Path == "/api/pause" || r.URL.Path == "/api/resume"):
		d.App.mu.Lock()
		d.App.paused = r.URL.Path == "/api/pause"
		d.App.mu.Unlock()
		reply(w, 200, map[string]bool{"ok": true})
	case r.Method == "POST" && r.URL.Path == "/api/revoke":
		d.App.mu.Lock()
		clear(d.App.sessions)
		d.App.mu.Unlock()
		reply(w, 200, map[string]bool{"ok": true})
	case r.Method == "GET" && r.URL.Path == "/api/profile":
		w.Header().Set("Content-Type", "application/x-apple-aspen-config")
		w.Header().Set("Content-Disposition", "attachment; filename=PC-to-iPhone-Share.mobileconfig")
		w.Write(d.App.certificates.Profile)
	case r.Method == "GET" && r.URL.Path == "/api/certificate":
		w.Header().Set("Content-Type", "application/pkix-cert")
		w.Header().Set("Content-Disposition", "attachment; filename=PC-to-iPhone-Share.cer")
		w.Write(d.App.certificates.DER)
	case r.Method == "POST" && r.URL.Path == "/api/stop":
		reply(w, 200, map[string]bool{"ok": true})
		d.once.Do(func() { close(d.Stop) })
	default:
		deny(w, 404, "Not found.")
	}
}
