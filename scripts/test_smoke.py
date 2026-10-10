"""Verify portable-archive handling without starting any server."""
import io
from pathlib import Path
from unittest.mock import patch
import tarfile
import tempfile
import unittest
import zipfile

from smoke_binary import unpack_binary


class PortableArchiveTests(unittest.TestCase):
    def test_zip_selects_native_filename(self):
        for system, name in [('Windows', 'pc-to-iphone-share.exe'), ('Darwin', 'pc-to-iphone-share')]:
            with self.subTest(system=system), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                archive = root / 'portable.zip'
                with zipfile.ZipFile(archive, 'w') as bundle:
                    bundle.writestr(name, b'synthetic executable')
                    bundle.writestr('../outside', b'never extracted')
                with patch('smoke_binary.platform.system', return_value=system):
                    binary = unpack_binary(archive, root)
                self.assertEqual(binary.name, name)
                self.assertEqual(binary.read_bytes(), b'synthetic executable')
                self.assertEqual(sorted(p.name for p in root.iterdir()), sorted([name, 'portable.zip']))

    def test_linux_tar(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            archive = root / 'portable.tar.gz'
            with tarfile.open(archive, 'w:gz') as bundle:
                entry = tarfile.TarInfo('pc-to-iphone-share')
                entry.size = 9
                bundle.addfile(entry, io.BytesIO(b'synthetic'))
            with patch('smoke_binary.platform.system', return_value='Linux'):
                binary = unpack_binary(archive, root)
            self.assertEqual(binary.read_bytes(), b'synthetic')

    def test_linked_tar_executable_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            archive = root / 'portable.tar.gz'
            with tarfile.open(archive, 'w:gz') as bundle:
                entry = tarfile.TarInfo('pc-to-iphone-share')
                entry.type, entry.linkname = tarfile.SYMTYPE, '/outside'
                bundle.addfile(entry)
            with patch('smoke_binary.platform.system', return_value='Linux'):
                with self.assertRaises(ValueError):
                    unpack_binary(archive, root)


if __name__ == '__main__':
    unittest.main()
