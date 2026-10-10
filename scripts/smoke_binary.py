#!/usr/bin/env python3
"""Run a binary against synthetic data on loopback only. Never uses an existing share."""
import argparse
import hashlib
import http.client
import json
import os
from pathlib import Path
import platform
import socket
import shutil
import ssl
import subprocess
import tempfile
import tarfile
from urllib.parse import urlsplit
import zipfile

def unpack_binary(binary, directory):
    """Copy only a regular native executable; never extract archive paths."""
    name = 'pc-to-iphone-share.exe' if platform.system() == 'Windows' else 'pc-to-iphone-share'
    if binary.suffix == '.zip':
        with zipfile.ZipFile(binary) as archive:
            member = archive.getinfo(name)
            if member.is_dir() or (member.external_attr >> 16) & 0o170000 == 0o120000:
                raise ValueError('archive executable must be a regular file')
            data = archive.read(member)
    elif binary.name.endswith('.tar.gz'):
        with tarfile.open(binary, 'r:gz') as archive:
            member = archive.getmember(name)
            if not member.isfile():
                raise ValueError('archive executable must be a regular file')
            data = archive.extractfile(member).read()
    else:
        return binary.resolve()
    executable = directory / name
    executable.write_bytes(data)
    executable.chmod(0o700)
    return executable

def smoke(binary):
    with tempfile.TemporaryDirectory(prefix='pc-to-iphone-share-smoke-') as temporary:
        temp = Path(temporary)
        executable = unpack_binary(binary, temp)
        folder, state = temp / 'synthetic-files', temp / 'private-state'
        folder.mkdir()
        payload = b'synthetic-download-content-' * 8192
        (folder / 'example.bin').write_bytes(payload)
        with socket.socket() as reservation:
            reservation.bind(('127.0.0.1', 0))
            port = reservation.getsockname()[1]
        process = subprocess.Popen([str(executable), '--folder', str(folder), '--state', str(state),
                                    '--loopback-only', '--port', str(port)],
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                                   creationflags=getattr(subprocess, 'CREATE_NO_WINDOW', 0))
        try:
            password = dashboard = None
            for _ in range(20):
                line = process.stdout.readline()
                if not line:
                    raise RuntimeError('binary stopped before becoming ready (runtime output redacted)')
                if line.startswith('New share password: '):
                    password = line.removeprefix('New share password: ').strip()
                if line.startswith('Local dashboard: '):
                    dashboard = urlsplit(line.removeprefix('Local dashboard: ').strip())
                if line.startswith('Keep this window open.'):
                    break
            if not password or not dashboard:
                raise RuntimeError('setup details missing (runtime output redacted)')
            context = ssl.create_default_context(cafile=str(state / 'ca.pem'))
            cookie = None
            def request(path, method='GET', body=None, extra=None):
                connection = http.client.HTTPSConnection('127.0.0.1', port, context=context, timeout=10)
                headers = {'Cookie': cookie} if cookie else {}
                if body is not None:
                    headers.update({'Origin': f'https://127.0.0.1:{port}', 'Content-Type': 'application/json'})
                    body = json.dumps(body)
                headers.update(extra or {})
                connection.request(method, path, body=body, headers=headers)
                response = connection.getresponse()
                result = response.status, dict(response.getheaders()), response.read()
                connection.close()
                return result
            assert request('/')[0] == 200
            assert request('/api/list')[0] == 401
            status, headers, _ = request('/api/login', 'POST', {'password': password})
            assert status == 200
            cookie = headers['Set-Cookie'].split(';')[0]
            assert request('/api/list')[0] == 200
            assert request('/download?path=example.bin')[2] == payload
            assert request('/download?path=example.bin', extra={'Range': 'bytes=7-101'})[2] == payload[7:102]
            assert request('/', extra={'Sec-Fetch-Site': 'cross-site', 'Sec-Fetch-Mode': 'navigate'})[0] == 200
            assert request('/api/list', extra={'Sec-Fetch-Site': 'cross-site'})[0] == 403
            connection = http.client.HTTPConnection(dashboard.hostname, dashboard.port, timeout=10)
            connection.request('GET', '/api/state')
            response = connection.getresponse(); assert response.status == 401; response.read()
            connection.close()
            connection = http.client.HTTPConnection(dashboard.hostname, dashboard.port, timeout=10)
            connection.request('POST', '/api/stop', headers={'Authorization': 'Bearer ' + dashboard.fragment,
                                'Origin': f'http://127.0.0.1:{dashboard.port}'})
            response = connection.getresponse(); assert response.status == 200; response.read(); connection.close()
            assert process.wait(timeout=10) == 0
            return {'native_smoke': 'passed', 'os': platform.system().lower(),
                    'https_certificate_verified': True, 'login_and_folder_access': 'passed',
                    'range_download': 'passed', 'download_integrity': hashlib.sha256(payload).hexdigest(),
                    'dashboard_isolation': 'passed', 'process_stopped': True,
                    'test_data': 'synthetic only', 'existing_servers_modified': False}
        finally:
            if process.poll() is None:
                process.terminate()
                process.wait(timeout=10)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('binary', type=Path, nargs='?', help='native binary or matching portable zip/tar.gz')
    parser.add_argument('--build', action='store_true', help='build and test the native target in temporary storage')
    parser.add_argument('--go', default=shutil.which('go'))
    args = parser.parse_args()
    if args.build:
        if not args.go:
            raise SystemExit('Go is required for --build.')
        root = Path(__file__).resolve().parents[1]
        go = str(Path(args.go).resolve())
        host_os, host_arch = subprocess.check_output([go, 'env', 'GOHOSTOS', 'GOHOSTARCH'], text=True).split()
        with tempfile.TemporaryDirectory(prefix='pc-to-iphone-native-build-') as temporary:
            binary = Path(temporary) / ('pc-to-iphone-share.exe' if host_os == 'windows' else 'pc-to-iphone-share')
            env = dict(os.environ, GOOS=host_os, GOARCH=host_arch, CGO_ENABLED='0', GOMAXPROCS='2')
            subprocess.run([go, 'build', '-p', '2', '-trimpath', '-buildvcs=false', '-ldflags=-s -w', '-o', str(binary), '.'], cwd=root, env=env, check=True)
            print(json.dumps(smoke(binary)))
    elif args.binary:
        print(json.dumps(smoke(args.binary.resolve())))
    else:
        parser.error('supply a native binary or --build')
