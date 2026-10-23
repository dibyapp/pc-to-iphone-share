#!/usr/bin/env python3
"""Loopback-only screenshot fixture. Synthetic data; never opens a shared folder."""
import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
from urllib.parse import urlsplit

WEB = Path(__file__).resolve().parents[1] / 'internal/share/web'
ENTRIES = [
    {'name': 'Reading list.pdf', 'directory': False, 'size': 2516582, 'modified': '2026-01-01T12:00:00Z'},
    {'name': 'Weekend photos.zip', 'directory': False, 'size': 88080384, 'modified': '2026-01-01T12:00:00Z'},
    {'name': 'Workshop notes.txt', 'directory': False, 'size': 12288, 'modified': '2026-01-01T12:00:00Z'},
    {'name': 'Reference', 'directory': True, 'size': 0, 'modified': '2026-01-01T12:00:00Z'},
]
ENTRIES.sort(key=lambda entry: (not entry['directory'], entry['name'].lower()))


class Demo(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def send(self, status, data, kind='application/json'):
        self.send_response(status)
        self.send_header('Content-Type', kind)
        self.send_header('Content-Length', str(len(data)))
        self.send_header('Cache-Control', 'no-store')
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        path = urlsplit(self.path).path
        if path == '/api/list':
            if self.server.screen == 'login' and 'demo-unlocked' not in self.headers.get('Cookie', ''):
                return self.send(401, b'{"error":"Please sign in."}')
            return self.send(200, json.dumps({'path':'','name':'Example shared folder','entries':ENTRIES,
                'nextOffset':len(ENTRIES),'total':len(ENTRIES),'hasMore':False,'truncated':False}).encode())
        if path == '/api/state':
            return self.send(200, json.dumps({'folder':'Example shared folder','url':'https://your-computer:9443',
                'password':'DEMO ONLY - not a real password','fingerprint':'DEMO - verify your real fingerprint locally',
                'version':'0.2.0-alpha.1'}).encode())
        if path == '/api/stats':
            return self.send(200, json.dumps({'paused':False,'active':0,'completed':3,'bytes':12582912,
                'sessions':1,'maxDownloads':4,'expiresAt':'0001-01-01T00:00:00Z'}).encode())
        names = {'/':'dashboard.html' if self.server.screen == 'dashboard' else 'index.html',
                 '/style.css':'style.css','/app.js':'app.js','/dashboard.js':'dashboard.js'}
        if path not in names:
            return self.send(404, b'{"error":"Screenshot demo only."}')
        file = WEB / names[path]
        data = file.read_bytes()
        if file.suffix == '.html':
            banner = '<p style="text-align:center;color:#b8d9a7;font:11px system-ui;letter-spacing:1px;margin:12px">SCREENSHOT DEMO · SYNTHETIC DATA · NO LIVE SHARE</p>'
            data = data.replace(b'<body>', ('<body>'+banner).encode())
        self.send(200, data, {'.html':'text/html; charset=utf-8','.js':'text/javascript; charset=utf-8',
                              '.css':'text/css; charset=utf-8'}[file.suffix])

    def do_POST(self):
        self.rfile.read(min(int(self.headers.get('Content-Length', '0')), 4096))
        if self.path == '/api/login':
            self.send_response(200)
            self.send_header('Set-Cookie', 'demo-unlocked=1; SameSite=Strict; HttpOnly')
            self.end_headers()
            self.wfile.write(b'{}')
        elif self.path == '/api/logout':
            self.send(200, b'{}')
        else:
            self.send(405, b'{"error":"Screenshot demo has no management actions."}')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--screen', choices=['login','browser','dashboard'], default='browser')
    parser.add_argument('--port', type=int, default=0)
    args = parser.parse_args()
    server = ThreadingHTTPServer(('127.0.0.1', args.port), Demo)
    server.screen = args.screen
    print(f'Synthetic preview: http://127.0.0.1:{server.server_port}/', flush=True)
    server.serve_forever()
