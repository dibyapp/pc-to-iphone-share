import io
from pathlib import Path
import tempfile
import subprocess
import unittest
from unittest.mock import patch
import zipfile
from privacy_check import inspect, safe_member, scan_artifacts, scan_source

class PrivacyTests(unittest.TestCase):
    def test_sensitive_material_is_detected(self):
        marker = b'-----BEGIN ' + b'PRIVATE KEY-----'
        self.assertTrue(inspect(marker, 'example', []))
        self.assertTrue(inspect(b'prefix machine-identifier suffix', 'example', [b'machine-identifier']))
        self.assertTrue(inspect('machine-identifier'.encode('utf-16le'), 'example', [b'machine-identifier']))
        self.assertFalse(inspect(b'https://github.com/dibyapp/pc-to-iphone-share', 'public attribution', [b'diby' + b'a']))
        self.assertTrue(inspect(b'account: diby' + b'a', 'local identity', [b'diby' + b'a']))
        self.assertTrue(inspect(b'C:' + b'/Users/' + b'example-person/project', 'example', []))

    def test_unsafe_members_are_rejected(self):
        for name in ['../outside.txt', '..\\outside.txt', 'C:outside.txt', '/absolute.txt', 'private/config.json', 'ca.key', 'iphone.mobileconfig']:
            self.assertFalse(safe_member(name), name)
        self.assertTrue(safe_member('docs/QUICKSTART.md'))

    def test_archive_contents_are_inspected(self):
        with tempfile.TemporaryDirectory(prefix='pc-to-iphone-privacy-test-') as temporary:
            folder = Path(temporary)
            with zipfile.ZipFile(folder / 'bad.zip', 'w') as archive:
                archive.writestr('private/config.json', '{}')
                archive.writestr('README.md', 'machine-identifier')
            issues, members = scan_artifacts(folder, [b'machine-identifier'])
            self.assertGreaterEqual(len(issues), 2)
            self.assertEqual(members, 2)

    def test_old_staged_git_blob_is_detected_after_cleanup(self):
        with tempfile.TemporaryDirectory(prefix='pc-to-iphone-git-privacy-test-') as temporary:
            root = Path(temporary)
            subprocess.run(['git', 'init', '-q', str(root)], check=True)
            readme = root / 'README.md'
            readme.write_text('synthetic-sensitive-identifier', encoding='utf-8')
            subprocess.run(['git', '-C', str(root), 'add', 'README.md'], check=True)
            readme.write_text('clean public documentation', encoding='utf-8')
            subprocess.run(['git', '-C', str(root), 'add', 'README.md'], check=True)
            with patch('privacy_check.ROOT', root):
                issues, count = scan_source([b'synthetic-sensitive-identifier'])
            self.assertEqual(count, 1)
            self.assertTrue(any('Git object' in issue for issue in issues))

if __name__ == '__main__':
    unittest.main()
