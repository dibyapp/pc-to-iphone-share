"""Exercise target selection, checksums, and archive containment with synthetic data."""
import hashlib
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch
import zipfile
from install import checksum, safe_name, target, unpack, main


class InstallerTests(unittest.TestCase):
    def test_platform_selection(self):
        for system, cpu, expected in [('Windows','AMD64',('windows','amd64')),('Windows','ARM64',('windows','arm64')),
            ('Darwin','arm64',('macos','arm64')),('Darwin','x86_64',('macos','amd64')),
            ('Linux','x86_64',('linux','amd64')),('Linux','aarch64',('linux','arm64'))]:
            self.assertEqual(target(system,cpu),expected)
        with self.assertRaises(ValueError): target('Linux','i686')

    def test_checksum_requires_exact_asset(self):
        digest = hashlib.sha256(b'synthetic').hexdigest()
        self.assertEqual(checksum((digest+'  example.zip\n').encode(),'example.zip'),digest)
        with self.assertRaises(ValueError): checksum((digest+'  other.zip\n').encode(),'example.zip')
        with self.assertRaises(ValueError): checksum((digest+'  example.zip\n').encode()*2,'example.zip')

    def test_regular_zip_installs_without_overwrite(self):
        buffer = io.BytesIO()
        with zipfile.ZipFile(buffer,'w') as archive: archive.writestr('pc-to-iphone-share.exe',b'synthetic')
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            unpack(buffer.getvalue(),'example.zip',root)
            self.assertEqual((root/'pc-to-iphone-share.exe').read_bytes(),b'synthetic')
            with self.assertRaises(FileExistsError): unpack(buffer.getvalue(),'example.zip',root)

    def test_traversal_and_links_are_rejected(self):
        for name in ['../outside','/absolute','C:outside','..\\outside']:
            self.assertFalse(safe_name(name))
        buffer=io.BytesIO()
        with tarfile.open(fileobj=buffer,mode='w:gz') as archive:
            member=tarfile.TarInfo('pc-to-iphone-share'); member.type=tarfile.SYMTYPE; member.linkname='/outside'
            archive.addfile(member)
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(ValueError): unpack(buffer.getvalue(),'example.tar.gz',Path(directory))
            self.assertEqual(list(Path(directory).iterdir()),[])

    def test_corrupt_download_creates_no_installation(self):
        with tempfile.TemporaryDirectory() as directory:
            destination=Path(directory)/'new-installation'
            digest=hashlib.sha256(b'expected archive').hexdigest()
            manifest=(digest+'  pc-to-iphone-share-0.1.0-alpha.1-linux-amd64.tar.gz\n').encode()
            with patch('install.target',return_value=('linux','amd64')), \
                 patch('install.fetch',side_effect=[manifest,b'corrupt archive']), \
                 patch('sys.argv',['install.py','--directory',str(destination)]):
                with self.assertRaisesRegex(ValueError,'Checksum mismatch'): main()
            self.assertFalse(destination.exists())


if __name__ == '__main__': unittest.main()
