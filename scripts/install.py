#!/usr/bin/env python3
"""Install a pinned portable release, verify its checksum, and never launch it."""
import argparse
import hashlib
import io
import os
from pathlib import Path, PurePosixPath
import platform
import re
import tarfile
import tempfile
import urllib.request
import zipfile

REPOSITORY = 'dibyapp/pc-to-iphone-share'
VERSION = '0.1.0-alpha.1'


def target(system=None, machine=None):
    systems = {'Windows':'windows','Darwin':'macos','Linux':'linux'}
    arches = {'AMD64':'amd64','x86_64':'amd64','arm64':'arm64','aarch64':'arm64','ARM64':'arm64'}
    try:
        return systems[system or platform.system()], arches[machine or platform.machine()]
    except KeyError:
        raise ValueError('Portable releases support Windows/macOS/Linux on x64 or ARM64.')


def safe_name(name):
    path = PurePosixPath(name)
    return bool(name) and not path.is_absolute() and '..' not in path.parts and '\\' not in name and ':' not in name


def unpack(data, filename, destination):
    """Validate all members before writing regular files; links are rejected."""
    files = []
    if filename.endswith('.zip'):
        with zipfile.ZipFile(io.BytesIO(data)) as archive:
            for item in archive.infolist():
                if item.file_size > 64*1024*1024:
                    raise ValueError('Unexpected unpacked release size.')
                if not safe_name(item.filename) or (item.external_attr >> 16) & 0o170000 == 0o120000:
                    raise ValueError('Unsafe archive member; installation stopped.')
                if not item.is_dir():
                    files.append((item.filename, archive.read(item), (item.external_attr >> 16) & 0o777))
    else:
        with tarfile.open(fileobj=io.BytesIO(data), mode='r:gz') as archive:
            for item in archive.getmembers():
                if item.size > 64*1024*1024:
                    raise ValueError('Unexpected unpacked release size.')
                if not safe_name(item.name) or not (item.isdir() or item.isfile()):
                    raise ValueError('Unsafe archive member; installation stopped.')
                if item.isfile():
                    files.append((item.name, archive.extractfile(item).read(), item.mode))
    names = [name for name, _, _ in files]
    if len(files) > 500 or sum(len(content) for _, content, _ in files) > 128*1024*1024:
        raise ValueError('Unexpected unpacked release size.')
    if len(names) != len(set(names)):
        raise ValueError('Duplicate archive member; installation stopped.')
    for name, content, mode in files:
        path = destination / name
        path.parent.mkdir(parents=True, exist_ok=True)
        with path.open('xb') as output:
            output.write(content)
        path.chmod(0o755 if mode & 0o111 else 0o644)


def fetch(url, maximum):
    request = urllib.request.Request(url, headers={'User-Agent':'PC-to-iPhone-Share-installer'})
    with urllib.request.urlopen(request, timeout=60) as response:
        if not response.url.startswith('https://'):
            raise ValueError('Release download must use HTTPS.')
        data = response.read(maximum + 1)
    if len(data) > maximum:
        raise ValueError('Unexpected release download size.')
    return data


def checksum(manifest, filename):
    rows = [line.split('  ', 1) for line in manifest.decode('ascii').splitlines()]
    matches = [row[0] for row in rows if len(row) == 2 and row[1] == filename]
    if len(matches) != 1 or not re.fullmatch(r'[0-9a-f]{64}', matches[0]):
        raise ValueError('Release checksum missing or invalid.')
    return matches[0]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', default=VERSION, help='explicit release version, without v')
    parser.add_argument('--directory', type=Path, help='new installation directory; existing directories are never overwritten')
    parser.add_argument('--dry-run', action='store_true', help='show target and downloads without fetching or installing')
    args = parser.parse_args()
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?', args.version):
        parser.error('Use a semantic version without shell or path characters.')
    system, arch = target()
    suffix = 'tar.gz' if system == 'linux' else 'zip'
    filename = f'pc-to-iphone-share-{args.version}-{system}-{arch}.{suffix}'
    base = f'https://github.com/{REPOSITORY}/releases/download/v{args.version}/'
    print(f'Target: {system}/{arch}; release v{args.version}')
    if args.dry_run:
        print(base + filename)
        print('No files downloaded, installed, or launched.')
        return
    if args.directory:
        destination = args.directory.expanduser().absolute()
    else:
        if system == 'windows':
            app_dir = Path(os.environ.get('LOCALAPPDATA', Path.home()/'AppData/Local'))/'Programs/PCToIPhoneShare'
        elif system == 'macos':
            app_dir = Path.home()/'Applications/PCToIPhoneShare'
        else:
            app_dir = Path.home()/'.local/lib/pc-to-iphone-share'
        destination = app_dir / args.version
    if destination.exists() or destination.is_symlink():
        raise ValueError('Installation directory already exists. Choose a new --directory; no files were changed.')
    manifest = fetch(base+'SHA256SUMS', 1024*1024)
    data = fetch(base+filename, 64*1024*1024)
    if hashlib.sha256(data).hexdigest() != checksum(manifest, filename):
        raise ValueError('Checksum mismatch; no application was installed.')
    destination.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='pc-to-iphone-install-', dir=destination.parent) as scratch:
        unpack(data, filename, Path(scratch))
        binary = Path(scratch)/('pc-to-iphone-share.exe' if system == 'windows' else 'pc-to-iphone-share')
        if not binary.is_file():
            raise ValueError('Release has no expected executable.')
        destination.mkdir()
        for item in Path(scratch).iterdir():
            item.rename(destination/item.name)
    print(f'Checksum verified. Installed portable files in: {destination}')
    print('Start the executable or included launcher yourself. Nothing was launched.')
    print('No PATH, firewall, certificate trust, service, or private-state changes were made.')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError) as error:
        raise SystemExit(str(error))
