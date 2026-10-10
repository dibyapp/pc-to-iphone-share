"""Check SEO metadata, local links, and JSON-LD without launching a browser or server."""
from html.parser import HTMLParser
import json
from pathlib import Path
import unittest
from urllib.parse import urlsplit

SITE = Path(__file__).resolve().parents[1] / 'site'

class Page(HTMLParser):
    def __init__(self, text):
        super().__init__()
        self.title, self.description, self.links, self.schemas = '', '', [], []
        self.in_title = False
        self.schema = None
        self.h1 = 0
        self.feed(text)
    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if tag == 'title': self.in_title = True
        if tag == 'meta' and attrs.get('name') == 'description': self.description = attrs.get('content', '')
        if tag == 'h1': self.h1 += 1
        if tag in {'a', 'link', 'img'}:
            value = attrs.get('href', attrs.get('src', ''))
            if value: self.links.append(value)
        if tag == 'script' and attrs.get('type') == 'application/ld+json': self.schema = ''
    def handle_data(self, data):
        if self.in_title: self.title += data
        if self.schema is not None: self.schema += data
    def handle_endtag(self, tag):
        if tag == 'title': self.in_title = False
        if tag == 'script' and self.schema is not None:
            self.schemas.append(json.loads(self.schema)); self.schema = None

class SiteTests(unittest.TestCase):
    def test_pages_and_metadata(self):
        files = sorted(SITE.rglob('*.html'))
        self.assertEqual(len(files), 8)
        titles = set()
        for file in files:
            page = Page(file.read_text(encoding='utf-8'))
            self.assertEqual(page.h1, 1, file.name)
            self.assertTrue(20 <= len(page.title) <= 100, file.name)
            self.assertTrue(70 <= len(page.description) <= 200, file.name)
            self.assertNotIn(page.title, titles)
            titles.add(page.title)
            for link in page.links:
                split = urlsplit(link)
                if split.scheme or not split.path: continue
                target = (file.parent / split.path).resolve()
                self.assertTrue(target.is_relative_to(SITE), file.name)
                self.assertTrue(target.is_file(), link)
    def test_structured_content(self):
        page = Page((SITE / 'index.html').read_text(encoding='utf-8'))
        self.assertEqual({item['@type'] for item in page.schemas}, {'SoftwareApplication', 'FAQPage'})
        faq = next(item for item in page.schemas if item['@type'] == 'FAQPage')
        self.assertEqual(len(faq['mainEntity']), 6)
        text = (SITE / 'index.html').read_text(encoding='utf-8')
        for item in faq['mainEntity']:
            self.assertIn(item['name'], text)

if __name__ == '__main__':
    unittest.main()
