#!/usr/bin/env python3
"""Build portable desktop archives locally. No push, upload, release, or deploy operation."""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

from privacy_check import ROOT, source_files, private_identifiers, scan_source, scan_artifacts

TARGETS = [('windows', 'amd64'), ('windows', 'arm64'), ('darwin', 'amd64'),
           ('darwin', 'arm64'), ('linux', 'amd64'), ('linux', 'arm64')]
EPOCH = 1704067200

def zip_archive(destination, entries):
    with zipfile.ZipFile(destination, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        for name, data, mode in entries:
            info = zipfile.ZipInfo(name, (2024, 1, 1, 0, 0, 0))
            info.create_system = 3
            info.external_attr = (0o100000 | mode) << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            archive.writestr(info, data)

def tar_archive(destination, entries):
    with destination.open('wb') as output:
        with gzip.GzipFile(filename='', mode='wb', fileobj=output, mtime=0) as compressed:
            with tarfile.open(fileobj=compressed, mode='w') as archive:
                for name, data, mode in entries:
                    info = tarfile.TarInfo(name)
                    info.size, info.mode, info.mtime = len(data), mode, EPOCH
                    info.uid, info.gid, info.uname, info.gname = 0, 0, '', ''
                    archive.addfile(info, io.BytesIO(data))

def check_issues(issues):
    if issues:
        raise SystemExit('\n'.join(issues) + '\nRelease preparation stopped; nothing uploaded.')

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--version', default='0.1.0-alpha.1')
    parser.add_argument('--go', default=shutil.which('go'))
    parser.add_argument('--output', default='dist', help='empty output directory under dist/, release/, or artifacts/')
    args = parser.parse_args()
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?', args.version):
        raise SystemExit('Use a semantic version without shell or path characters.')
    if not args.go:
        raise SystemExit('Install Go 1.26 or supply --go /path/to/go.')
    go = str(Path(args.go).resolve())
    forbidden = private_identifiers()
    issues, file_count = scan_source(forbidden)
    check_issues(issues)
    base_env = dict(os.environ, CGO_ENABLED='0', GOMAXPROCS='2')
    for command in ([go, 'test', '-count=1', './...'], [go, 'vet', './...']):
        subprocess.run(command, cwd=ROOT, env=base_env, check=True)
    toolchain = subprocess.check_output([go, 'version'], env=base_env, text=True).split()[2]
    if not re.match(r'go1\.(?:2[6-9]|[3-9]\d)\.', toolchain):
        raise SystemExit('Go 1.26 or newer is required.')
    out = (ROOT / args.output).resolve()
    try:
        relative_output = out.relative_to(ROOT)
    except ValueError:
        raise SystemExit('Output must stay inside this project.')
    if not relative_output.parts or relative_output.parts[0] not in {'dist', 'release', 'artifacts'}:
        raise SystemExit('Use an output directory under dist/, release/, or artifacts/.')
    out.mkdir(parents=True, exist_ok=True)
    # Never recursively remove an output directory: unexpected files require review.
    if any(out.iterdir()):
        raise SystemExit('dist must be empty. Preserve previous builds elsewhere before rebuilding.')
    documentation = ['README.md', 'LICENSE', 'THIRD_PARTY_NOTICES.md', 'SECURITY.md',
                     'CONTRIBUTING.md', 'CODE_OF_CONDUCT.md', 'CHANGELOG.md']
    documentation += [file.relative_to(ROOT).as_posix() for file in sorted((ROOT / 'docs').glob('*.md'))]
    documentation += [file.relative_to(ROOT).as_posix() for file in sorted((ROOT / 'docs').glob('*.csv'))]
    documentation += [file.relative_to(ROOT).as_posix() for file in sorted((ROOT / 'docs/screenshots').glob('*')) if file.is_file()]
    common = [(name, (ROOT / name).read_bytes(), 0o644) for name in documentation]
    with tempfile.TemporaryDirectory(prefix='pc-to-iphone-share-build-') as scratch:
        for platform, arch in TARGETS:
            label = 'macos' if platform == 'darwin' else platform
            name = 'pc-to-iphone-share.exe' if platform == 'windows' else 'pc-to-iphone-share'
            binary = Path(scratch) / name
            env = dict(base_env, GOOS=platform, GOARCH=arch)
            print(f'Building {label}/{arch}', flush=True)
            subprocess.run([go, 'build', '-p', '2', '-trimpath', '-buildvcs=false',
                            '-ldflags', '-s -w -X main.version=' + args.version, '-o', str(binary), '.'],
                           cwd=ROOT, env=env, check=True)
            entries = [(name, binary.read_bytes(), 0o755), *common]
            if platform == 'windows':
                entries.append(('Start-PC-to-iPhone-Share.cmd', b'@echo off\r\n"%~dp0pc-to-iphone-share.exe" %*\r\nif errorlevel 1 pause\r\n', 0o644))
            else:
                launcher = b'#!/bin/sh\nset -eu\ncd -- "$(dirname -- "$0")"\nexec ./pc-to-iphone-share "$@"\n'
                entries.append(('Start-PC-to-iPhone-Share.command' if platform == 'darwin' else 'start-pc-to-iphone-share.sh', launcher, 0o755))
            filename = f'pc-to-iphone-share-{args.version}-{label}-{arch}'
            if platform == 'linux':
                tar_archive(out / (filename + '.tar.gz'), entries)
            else:
                zip_archive(out / (filename + '.zip'), entries)
    source = [(f'pc-to-iphone-share-{args.version}/' + file.relative_to(ROOT).as_posix(), file.read_bytes(), 0o644)
              for file in source_files()]
    zip_archive(out / f'pc-to-iphone-share-{args.version}-source.zip', source)
    info = {'version': args.version, 'toolchain': toolchain,
            'targets': [f'{os_name}/{arch}' for os_name, arch in TARGETS],
            'flags': ['CGO_ENABLED=0', '-trimpath', '-buildvcs=false', '-s', '-w'],
            'source_files': file_count, 'signing': 'unsigned, not notarized',
            'native_validation': 'Cross-compilation only in this step; native target validation is a separate release gate',
            'publishing': 'none'}
    (out / 'BUILD-INFO.json').write_text(json.dumps(info, indent=2) + '\n', encoding='utf-8')
    manifest = []
    for file in sorted(out.iterdir()):
        manifest.append(hashlib.sha256(file.read_bytes()).hexdigest() + '  ' + file.name)
    (out / 'SHA256SUMS').write_text('\n'.join(manifest) + '\n', encoding='ascii')
    check_issues(scan_artifacts(out, forbidden)[0])
    print('Built and privacy-scanned six desktop archives and one source archive. Nothing uploaded.')

if __name__ == '__main__':
    main()
