#!/usr/bin/env python3
"""Scan an explicit public-source allowlist and release members, without uploading data."""
import argparse
import getpass
import json
import os
from pathlib import Path, PurePosixPath
import re
import socket
import subprocess
import tarfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
PUBLIC_FILES = ('main.go', 'go.mod', 'go.sum', 'LICENSE', 'README.md', 'CONTRIBUTING.md',
                'SECURITY.md', 'CODE_OF_CONDUCT.md', 'CHANGELOG.md', 'THIRD_PARTY_NOTICES.md',
                '.gitignore', '.gitattributes')
PUBLIC_DIRS = ('internal', 'docs', 'site', 'scripts', '.github')
BLOCKED_SUFFIXES = ('.pem', '.key', '.p12', '.pfx', '.cer', '.crt', '.mobileconfig', '.log', '.pid')
BLOCKED_NAMES = {'private', 'state', 'data', '.env', 'access.txt', 'config.json', '.git',
                 'node_modules', '__pycache__', '.tooling', 'dist', 'release', 'artifacts'}
PATTERNS = {
    'private key': rb'-----BEGIN (?:[A-Z]+ )?PRIVATE KEY-----',
    'cloud/API credential': rb'(?:AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9]{30,}|sk-[A-Za-z0-9_-]{32,})',
    'Windows user directory': rb'(?i)[a-z]:[\\/]Users[\\/][^\\/\s<>"\x00]{2,}',
    'Unix home directory': rb'/(?:Users|home)/[a-zA-Z0-9_.-]{2,}/',
}

def source_files():
    known = set(PUBLIC_FILES) | set(PUBLIC_DIRS) | {'.git', 'dist', 'release', 'artifacts', '__pycache__'}
    for item in ROOT.iterdir():
        if item.name not in known:
            raise ValueError(f'unreviewed project root entry: {item.name}')
    for name in PUBLIC_FILES:
        if (ROOT / name).is_symlink():
            raise ValueError(f'linked source path: {name}')
    files = [ROOT / name for name in PUBLIC_FILES if (ROOT / name).is_file()]
    for directory in PUBLIC_DIRS:
        base = ROOT / directory
        if not base.exists():
            continue
        if base.is_symlink():
            raise ValueError(f'linked source directory: {directory}')
        for item in sorted(base.rglob('*')):
            relative = item.relative_to(ROOT)
            if any(part in BLOCKED_NAMES for part in relative.parts):
                if '__pycache__' in relative.parts:
                    continue
                raise ValueError(f'blocked source path: {relative.as_posix()}')
            if item.is_symlink():
                raise ValueError(f'linked source path: {relative.as_posix()}')
            if item.is_file():
                if item.suffix.lower() in BLOCKED_SUFFIXES:
                    raise ValueError(f'certificate/secret artifact in source: {relative.as_posix()}')
                files.append(item)
    files = sorted(set(files))
    if (ROOT / '.git').exists():
        tracked = subprocess.run(['git', 'ls-files', '--cached', '-z'], cwd=ROOT, capture_output=True, check=True).stdout
        allowed = {file.relative_to(ROOT).as_posix() for file in files}
        for name in tracked.decode().split('\0'):
            if name and name not in allowed:
                raise ValueError(f'tracked file outside public allowlist: {name}')
    return files

def private_identifiers(extra=()):
    # Evaluate local identity at scan time; never write its values into a report or archive.
    account = getpass.getuser()
    generic_accounts = {'root', 'runner', 'runneradmin', 'admin', 'administrator', 'system', 'user', 'ubuntu', 'share'}
    values = [str(ROOT), str(Path.home()), socket.gethostname(), *extra]
    if account.lower() not in generic_accounts:
        values.append(account)
    return [value.encode().lower() for value in values if value and len(value) >= 4]

def inspect(data, label, forbidden):
    issues = []
    lower = data.lower()
    # The owner explicitly approved public GitHub attribution. Its handle
    # contains a local account-name prefix; allow the exact public handle,
    # while user-directory patterns and the actual local account stay blocked.
    lower = re.sub(rb'(?<![a-z0-9_-])dibyapp(?![a-z0-9_-])', b'public-github-owner', lower)
    for value in forbidden:
        # Windows resources or text exports can encode identifiers as UTF-16.
        text = value.decode('utf-8')
        if any(encoded in lower for encoded in (value, text.encode('utf-16le'), text.encode('utf-16be'))):
            issues.append(f'{label}: local identifier detected (value redacted)')
            break
    for name, pattern in PATTERNS.items():
        if re.search(pattern, data):
            issues.append(f'{label}: {name} detected (value redacted)')
    return issues

def scan_source(forbidden):
    issues, count = [], 0
    for file in source_files():
        issues += inspect(file.relative_to(ROOT).as_posix().encode(), 'source filename', forbidden)
        issues += inspect(file.read_bytes(), file.relative_to(ROOT).as_posix(), forbidden)
        count += 1
    # Review staged blobs too: a cleaned working copy must not hide an earlier
    # staged version containing a secret in the same allowed filename.
    if (ROOT / '.git').exists():
        names = subprocess.run(['git', 'ls-files', '--cached', '-z'], cwd=ROOT, capture_output=True, check=True).stdout.decode().split('\0')
        names = [name for name in names if name]
        if names:
            if any('\n' in name or '\r' in name for name in names):
                raise ValueError('control characters in a staged filename')
            output = subprocess.run(['git', 'cat-file', '--batch'], cwd=ROOT,
                                    input=''.join(':' + name + '\n' for name in names).encode(),
                                    capture_output=True, check=True).stdout
            import io
            stream = io.BytesIO(output)
            for name in names:
                header = stream.readline().split()
                if len(header) != 3 or header[1] != b'blob':
                    raise ValueError('cannot inspect staged source')
                data = stream.read(int(header[2])); stream.read(1)
                issues += inspect(data, 'staged:' + name, forbidden)
        # Audit every local object, including older staged blobs no longer in
        # the index. History must be reviewed before a repository is published.
        objects = subprocess.run(['git', 'cat-file', '--batch-all-objects', '--batch'], cwd=ROOT,
                                 capture_output=True, check=True).stdout
        import io
        stream = io.BytesIO(objects)
        while header := stream.readline():
            fields = header.split()
            if len(fields) != 3:
                raise ValueError('cannot inspect Git objects')
            data = stream.read(int(fields[2])); stream.read(1)
            if fields[1] in (b'blob', b'commit', b'tag'):
                issues += inspect(data, 'Git object ' + fields[0].decode(), forbidden)
    return issues, count

def safe_member(name):
    if '\\' in name or ':' in name:
        return False
    p = PurePosixPath(name)
    return not p.is_absolute() and '..' not in p.parts and not any(part in BLOCKED_NAMES for part in p.parts) and p.suffix.lower() not in BLOCKED_SUFFIXES

def scan_artifacts(directory, forbidden):
    issues, count = [], 0
    for file in sorted(directory.iterdir()):
        if file.is_symlink() or not file.is_file():
            raise ValueError('release directory must contain archives and manifests only')
        issues += inspect(file.name.encode(), 'artifact filename', forbidden)
        if file.suffix == '.zip':
            with zipfile.ZipFile(file) as archive:
                for member in archive.infolist():
                    issues += inspect(member.filename.encode(), file.name + ':member filename', forbidden)
                    if member.is_dir():
                        continue
                    if not safe_member(member.filename):
                        issues.append(f'{file.name}: blocked archive member')
                    if (member.external_attr >> 16) & 0o170000 == 0o120000:
                        issues.append(f'{file.name}: linked archive member')
                    issues += inspect(archive.read(member), file.name + ':' + member.filename, forbidden)
                    count += 1
        elif file.name.endswith('.tar.gz'):
            with tarfile.open(file, 'r:gz') as archive:
                for member in archive.getmembers():
                    issues += inspect(member.name.encode(), file.name + ':member filename', forbidden)
                    if member.isdir():
                        continue
                    if not member.isfile() or not safe_member(member.name):
                        issues.append(f'{file.name}: unsafe archive member')
                        continue
                    issues += inspect(archive.extractfile(member).read(), file.name + ':' + member.name, forbidden)
                    count += 1
        else:
            issues += inspect(file.read_bytes(), file.name, forbidden)
            count += 1
    return issues, count

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--artifacts', type=Path)
    parser.add_argument('--deny-text', action='append', default=[], help='additional local identifiers; values are never printed')
    args = parser.parse_args()
    forbidden = private_identifiers(args.deny_text)
    issues, count = scan_source(forbidden)
    members = 0
    if args.artifacts:
        more, members = scan_artifacts(args.artifacts.resolve(), forbidden)
        issues += more
    if issues:
        for issue in issues:
            print(issue)
        raise SystemExit('Privacy check FAILED. Nothing was uploaded.')
    print(json.dumps({'privacy_check': 'passed', 'source_files': count, 'release_members': members,
                      'staged_and_local_git_objects_checked': (ROOT / '.git').exists(),
                      'local_identifiers': 'not detected', 'private_keys': 'not detected', 'upload': False}))

if __name__ == '__main__':
    try:
        main()
    except ValueError as error:
        raise SystemExit(str(error))
