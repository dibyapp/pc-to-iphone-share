#!/usr/bin/env python3
"""Generate accurate static SEO content locally. No hosting, analytics, or deployment."""
from pathlib import Path
import html
import json

ROOT = Path(__file__).resolve().parents[1]
SITE = ROOT / 'site'

GUIDE_NOTES = {
    'windows-to-iphone': 'Most Windows PCs use x64; ARM PCs need the ARM64 build. Extract the ZIP first and use the included CMD launcher. The unsigned alpha may be blocked by Windows protections; inspect and build source instead if needed. Safari downloads appear in Downloads or Files and do not automatically import media into Photos or Music. Do not use your Windows account password: the app generates a separate share password.',
    'linux-to-iphone': 'Use uname -m to choose x86_64 (amd64) or aarch64 (arm64). The static Go binary needs no glibc, Python, Node, or package installation. Extract outside the shared folder and run as your ordinary user. A noexec mount or lost execute permission can prevent launching. UFW/firewalld rules and guest-network isolation can block phone access; limit any allowance to the intended interface, subnet, and port.',
    'ubuntu-to-iphone': 'Ubuntu uses the Linux archive, including Debian-based desktop variants. Local executable checks passed on Ubuntu 24.04 under WSL 2; this validates Linux runtime behavior, not phone reachability through WSL NAT. For Windows-hosted files, the Windows app avoids that extra networking layer. Physical Ubuntu desktop/Wi-Fi and mobile pairing checks are recorded separately. Use a supported Ubuntu release, keep the terminal open, and prevent sleep during downloads.',
    'mac-to-android': 'Choose ARM64 for Apple Silicon or amd64 for Intel Macs. The alpha is unsigned and unnotarized: it does not bypass Gatekeeper. Terminal source builds are an alternative. Android user-CA installation menus and browser trust behavior vary, especially on managed devices. Install only your own public CA and verify its fingerprint; if device policy prevents that trust workflow, use a compatible alternative. Downloads go through the browser, not an automatic Photos import.',
    'private-local-sharing': 'Review the folder contents before sharing. Read-only endpoints prevent remote writes but do not make an accidental secret file safe to share. Avoid selecting a home directory, private settings folder, or source checkout. Subnet checks do not prove physical Wi-Fi membership, and administrators or same-user processes remain trusted. Hard links and mounts have different boundaries from symlinks. Stop the app when the temporary session is over.',
}

FAQ = [
    ('Does PC to iPhone Share upload my files to the cloud?', 'No. File downloads travel directly from your computer to your receiving device over the selected local network. The app includes no cloud relay or file-upload service.'),
    ('Can I download Windows files on an iPhone?', 'Yes. Run the Windows host, choose a folder, pair Safari with its local certificate, and sign in using the generated password.'),
    ('Do I need to install an app on my phone?', 'No. The receiving device uses its browser. A one-time local certificate pairing step is required to verify HTTPS.'),
    ('Is PC to iPhone Share available for every operating system?', 'Portable host builds target Windows, macOS, and Linux on x86-64 and ARM64. Ubuntu, Debian, Fedora, Linux Mint, Arch, openSUSE, and Alpine use the matching Linux build. iPhone, iPad, and Android devices are browser clients. Other operating systems and 32-bit hosts are not in the binary support matrix.'),
    ('Is there a maximum file size?', 'The app imposes no file-size cap. Available device storage, filesystem limits, browser behavior, and network conditions still matter. Downloads stream and support HTTP byte ranges.'),
    ('Is this a stable release?', 'The first release is an unsigned alpha. Source and six portable builds are available from the official GitHub repository. Native validation, physical device checks, and signing status are documented openly.'),
]

GUIDES = [
    ('windows-to-iphone', 'Transfer files from Windows PC to iPhone over Wi-Fi',
     'Download files from a Windows computer to an iPhone using local Wi-Fi, Safari, a generated password, and verified HTTPS. No cable or cloud upload.',
     'A PC file, ready on your iPhone.', 'Windows', 'iPhone Safari',
     'Extract the Windows amd64 or ARM64 archive. Run Start-PC-to-iPhone-Share.cmd or pc-to-iphone-share.exe. Keep the portable app outside the folder you share.'),
    ('linux-to-iphone', 'Download files from Linux to iPhone without a USB cable',
     'Browse a Linux folder and download files on your iPhone over a local network with PC to iPhone Share. Read-only access, password protection, and HTTPS.',
     'Linux files. An iPhone browser.', 'Linux', 'iPhone Safari',
     'Extract the Linux amd64 or ARM64 archive. In a terminal, run ./pc-to-iphone-share. Select one folder and the network interface reachable by your phone.'),
    ('ubuntu-to-iphone', 'Transfer files from Ubuntu to iPhone over Wi-Fi',
     'Download files from Ubuntu to iPhone with PC to iPhone Share: a portable Linux app, Safari, password protection, and local HTTPS. No cable or cloud uploads.',
     'Ubuntu files, ready on your iPhone.', 'Ubuntu Linux', 'iPhone Safari',
     'Use the Linux amd64 archive for an x86-64 Ubuntu computer, or Linux ARM64 for an aarch64 system. Extract into an app folder, run ./pc-to-iphone-share, and choose your shared folder and private network. No package-manager installation or root access is required.'),
    ('mac-to-android', 'Transfer files from Mac to Android over local Wi-Fi',
     'Download Mac files on Android over local Wi-Fi with PC to iPhone Share. Choose one folder, pair the local certificate, and use a password-protected browser.',
     'Your Mac files, on Android.', 'macOS', 'an Android browser',
     'Choose the macOS ARM64 build for Apple Silicon or amd64 for Intel. Run the portable binary from Terminal, choose a folder, and select your local network. The pre-release binary is unsigned and unnotarized.'),
    ('private-local-sharing', 'Password-protected local file sharing without cloud uploads',
     'Share one computer folder with your phone over a local network. PC to iPhone Share provides read-only HTTPS downloads without accounts or cloud uploads.',
     'Share a folder. Keep it local.', 'Windows, macOS, or Linux', 'a phone browser',
     'Run the portable build for your desktop operating system and CPU. Choose a dedicated folder containing only files you intend to share. Keep private settings and the executable outside that folder.'),
]

def head(title, description, prefix=''):
    return f'''<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{html.escape(title)}</title><meta name="description" content="{html.escape(description)}"><meta name="theme-color" content="#edf2e7"><meta property="og:type" content="website"><meta property="og:title" content="{html.escape(title)}"><meta property="og:description" content="{html.escape(description)}"><meta name="twitter:card" content="summary"><link rel="icon" href="{prefix}mark.svg" type="image/svg+xml"><link rel="stylesheet" href="{prefix}style.css"></head><body>'''

def header(prefix=''):
    return f'<header class="wrap nav"><a class="brand" href="{prefix}index.html"><img src="{prefix}mark.svg" width="34" height="38" alt=""><span>PC to iPhone Share</span></a><nav><a href="{prefix}install.html">Install</a><a href="{prefix}troubleshooting.html">Connection help</a><a href="https://github.com/dibyapp/pc-to-iphone-share">GitHub</a></nav><span class="pill">MIT · Alpha</span></header>'

def footer(prefix=''):
    return f'<footer class="wrap"><a class="brand" href="{prefix}index.html">PC to iPhone Share</a><p>One folder. Your local network. Open source.</p><p>Unsigned alpha · Portable builds are unsigned · No independent security audit</p></footer></body></html>'

def build_support_pages():
    base = 'https://github.com/dibyapp/pc-to-iphone-share/releases/download/v0.1.0-alpha.1/'
    rows = []
    for label, os_name in [('Windows','windows'),('macOS','macos'),('Ubuntu / Linux','linux')]:
        suffix = 'tar.gz' if os_name == 'linux' else 'zip'
        links = [f'<a href="{base}pc-to-iphone-share-0.1.0-alpha.1-{os_name}-{arch}.{suffix}">{text}</a>'
                 for arch,text in [('amd64','Intel Mac' if os_name == 'macos' else 'Intel/AMD x64'),('arm64','Apple Silicon' if os_name == 'macos' else 'ARM64')]]
        rows.append(f'<tr><th>{label}</th><td>{links[0]}</td><td>{links[1]}</td></tr>')
    body = '<main class="wrap article"><p class="eyebrow">DOWNLOAD ON YOUR COMPUTER</p><h1>Install once.<br>Download on your phone.</h1><p class="lede">Choose your desktop OS and CPU. The phone uses its browser; no mobile app is installed. This is an unsigned alpha with manual certificate pairing.</p>'
    body += '<h2>1. Choose a portable download</h2><div class="table-scroll"><table><thead><tr><th>Computer</th><th>x64</th><th>ARM64</th></tr></thead><tbody>'+''.join(rows)+'</tbody></table></div>'
    body += f'<p><a href="{base}SHA256SUMS">SHA-256 checksums</a> · <a href="https://github.com/dibyapp/pc-to-iphone-share/blob/main/docs/INSTALL.md">Complete installation and source-build guide</a></p>'
    body += '<h2>2. Extract and start</h2><p>Keep the portable app outside the folder you will share. Windows: run Start-PC-to-iPhone-Share.cmd. macOS: run ./pc-to-iphone-share in Terminal. Ubuntu/Linux: extract the TAR.GZ and run ./pc-to-iphone-share in a terminal. No Go, Node, Python, admin installation, or package-manager command is needed to run a binary. Binaries are unsigned and unnotarized; the app does not bypass OS protections.</p><h2>3. Choose your folder and network</h2><p>Select one dedicated folder and the actual private interface reachable by your phone. Keep the computer awake. The app leaves your firewall, router, certificate trust, and services under your control.</p><h2>4. Pair, sign in, download</h2><p>Open the computer dashboard from the console, keep that management URL private, and use the separate phone URL. Verify the local CA fingerprint through a trusted channel, install only your own public certificate, and explicitly enable trust on iPhone/iPad. Enter the generated password and tap a filename.</p><p><a href="https://github.com/dibyapp/pc-to-iphone-share/blob/main/docs/CERTIFICATES.md">Device certificate pairing</a> · <a href="troubleshooting.html">Connection help</a></p><h2>Prefer an inspected source build?</h2><p>Clone the repository, install Go 1.26, run go test ./... and go vet ./..., then build with -trimpath -buildvcs=false. The optional Python source-repository installer verifies the pinned release checksum and never launches or overwrites an existing app.</p><p class="note">Windows x64 and Ubuntu 24.04 under WSL 2 passed local runtime checks. Other architecture and physical-device results are documented openly. iOS and Android are receiving clients; there is no mobile server package or automatic discovery.</p></main>'
    page = head('Install PC to iPhone Share on Windows, macOS and Ubuntu', 'Choose a portable PC to iPhone Share download for Windows, Mac or Ubuntu/Linux. Check SHA-256, extract, pair your phone, and download files over local HTTPS.')+header()+body+footer()
    (SITE/'install.html').write_text(page,encoding='utf-8',newline='\n')
    body = '<main class="wrap article"><p class="eyebrow">GET TO YOUR FIRST DOWNLOAD</p><h1>Connection help,<br>without guesswork.</h1><p class="lede">Check the exact phone address, selected interface, certificate trust, and receiving-device storage before changing settings.</p>'
    sections = [
        ('The page cannot be reached','Use the separately displayed phone URL including https:// and the app port (9443 by default). The dashboard is computer-only and cannot be used from the phone. Both devices need a reachable private network; guest Wi-Fi client isolation can block them. Check that the terminal is still open, the computer is awake, and the chosen interface is the intended LAN interface.'),
        ('The certificate could not be verified','Trust is per receiving device. Verify the fingerprint using the local computer dashboard through a trusted channel, install only that computer\'s public CA/profile, and enable full trust on iPhone/iPad. Trusting the certificate on Windows does not trust it on an iPhone. The app does not turn off certificate verification. See the certificate guide for supported manual steps.'),
        ('Safari says cross-site requests are blocked','Open the exact phone URL directly in a fresh Safari tab and sign in there. Use one scheme, hostname/address, and port for the session. Do not embed the page in another site or reuse a dashboard URL. Protected APIs intentionally reject cross-site requests; disabling Safari privacy settings is not the documented fix.'),
        ('The executable will not start','Choose the correct CPU/OS archive and extract it before running. On Linux, check execute permission and noexec mounts. Unsigned Windows/macOS binaries may be blocked by OS protections; use an inspected source build until signing is available. Old or unsupported operating systems and 32-bit hosts are outside the packaged targets.'),
        ('A download stops or a file is missing','Keep the computer awake and the app running. Check phone storage and the browser download manager. Resume depends on the client despite HTTP range support. Search filters loaded entries: use Load more files for additional pages. Dot paths, symlinks/junctions, and invalid names are intentionally excluded; directories over 10,000 entries should be split.'),
        ('Still stuck?','Report the app version, OS/CPU, sanitized error message, and synthetic reproduction steps. Never post private addresses, usernames, paths, filenames, passwords, dashboard URLs/tokens, config, keys, or certificate fingerprints in a public issue.'),
    ]
    for title, explanation in sections:
        body += f'<h2>{html.escape(title)}</h2><p>{html.escape(explanation)}</p>'
    body += '<p><a href="https://github.com/dibyapp/pc-to-iphone-share/blob/main/docs/CERTIFICATES.md">Certificate guide</a> · <a href="https://github.com/dibyapp/pc-to-iphone-share/issues/new/choose">Sanitized support report</a> · <a href="install.html">Installation</a></p></main>'
    page=head('PC to iPhone Share connection, certificate and Safari help', 'Fix local file-sharing setup: unreachable phone URLs, certificate trust, Safari cross-site errors, Linux permissions and interrupted downloads. Safe checks first.')+header()+body+footer()
    (SITE/'troubleshooting.html').write_text(page,encoding='utf-8',newline='\n')

def build():
    SITE.mkdir(exist_ok=True); (SITE / 'guides').mkdir(exist_ok=True)
    schema = {'@context': 'https://schema.org', '@type': 'SoftwareApplication', 'name': 'PC to iPhone Share',
              'applicationCategory': 'UtilitiesApplication', 'operatingSystem': 'Windows 10+; macOS 12+; Linux',
              'softwareVersion': '0.1.0-alpha.1', 'url': 'https://github.com/dibyapp/pc-to-iphone-share',
              'downloadUrl': 'https://github.com/dibyapp/pc-to-iphone-share/releases/tag/v0.1.0-alpha.1',
              'author': {'@type':'Person','name':'dibyapp','url':'https://github.com/dibyapp'}, 'license': 'https://opensource.org/license/mit',
              'description': 'Password-protected local computer-to-phone file downloads over HTTPS. No cloud uploads, accounts, or mobile app required.'}
    faq_schema = {'@context': 'https://schema.org', '@type': 'FAQPage',
                  'mainEntity': [{'@type': 'Question', 'name': question, 'acceptedAnswer': {'@type': 'Answer', 'text': answer}} for question, answer in FAQ]}
    page = head('PC to iPhone Share | Secure Wi-Fi File Transfer to Your Phone', 'Download files from Windows, Mac, or Linux to iPhone and Android over local Wi-Fi. Password-protected HTTPS. No cloud uploads, cables, or phone app.') + header()
    page += '''<main><section class="wrap hero"><div class="hero-copy"><p class="eyebrow">LOCAL FILE SHARING, WITH LESS FRICTION</p><h1>Your files.<br>One Wi-Fi away.</h1><p class="lede">Download a computer folder on your phone, without sending it through the cloud. Choose what to share. Protect it with a password. Keep the connection encrypted.</p><a class="button" href="install.html">Download for your computer <span>↗</span></a><p class="micro">Windows · macOS · Linux hosts<br>iPhone · iPad · Android browser clients</p></div><figure class="actual-screenshot"><img src="assets/folder-phone.jpg" width="375" height="912" alt="PC to iPhone Share phone-width file browser with synthetic example files"><figcaption>Actual interface · Synthetic screenshot demo</figcaption></figure></section>
<section class="wrap value-grid"><article><span class="number">01</span><h2>Skip the upload.</h2><p>Your computer sends files directly to your receiving device over the selected local network.</p></article><article><span class="number">02</span><h2>Use the browser.</h2><p>No account, phone app, or cable. Pair the local certificate once, then sign in and download.</p></article><article><span class="number">03</span><h2>Share one folder.</h2><p>Read-only access, a generated password, expiring sessions, and a stop control put the scope in your hands.</p></article></section>
<section class="band" id="how"><div class="wrap"><p class="eyebrow">A CLEAR START</p><h2>From desktop to phone,<br>in a few deliberate steps.</h2><ol class="steps"><li><strong>Run the portable app.</strong><p>Choose the build for your desktop OS and processor. No Node, Python, or Go runtime is required.</p></li><li><strong>Choose a folder and network.</strong><p>The app listens on one private IPv4 interface. No automatic firewall or router changes.</p></li><li><strong>Pair your device.</strong><p>Use the local dashboard to verify the address and certificate fingerprint. Install only your own computer’s public certificate.</p></li><li><strong>Sign in. Download. Stop when done.</strong><p>Use the phone URL and generated password. Keep the computer awake while files transfer.</p></li></ol><p class="note">First-time certificate trust is a manual setup step. Portable builds are unsigned pre-releases; platform validation and signing are documented honestly.</p></div></section>
<section class="wrap guides"><p class="eyebrow">FOR YOUR DEVICES</p><h2>A familiar browser,<br>across different ecosystems.</h2><div class="guide-grid">'''
    for slug, title, *_ in GUIDES:
        page += f'<a class="guide" href="guides/{slug}.html"><h3>{html.escape(title)}</h3><span>Read the guide →</span></a>'
    page += '</div></section><section class="wrap faq" id="faq"><p class="eyebrow">GOOD QUESTIONS</p><h2>Know what you’re sharing.</h2>'
    for question, answer in FAQ:
        page += f'<details><summary>{html.escape(question)}</summary><p>{html.escape(answer)}</p></details>'
    page += '</section><section class="wrap final-card"><h2>Less cloud.<br>More control.</h2><p>PC to iPhone Share is MIT licensed, readable, and built around one job: downloading a chosen computer folder on your own devices.</p><a class="button" href="https://github.com/dibyapp/pc-to-iphone-share">Explore source and downloads ↗</a></section></main>'
    page += '<script type="application/ld+json">' + json.dumps(schema) + '</script><script type="application/ld+json">' + json.dumps(faq_schema) + '</script>' + footer()
    (SITE / 'index.html').write_text(page, encoding='utf-8', newline='\n')
    for slug, title, description, headline, host, client, first in GUIDES:
        page = head(title + ' | PC to iPhone Share', description, '../') + header('../')
        note = GUIDE_NOTES[slug]
        page += f'<main class="wrap article"><p class="eyebrow">A PRACTICAL LOCAL TRANSFER GUIDE</p><h1>{html.escape(headline)}</h1><p class="lede">{html.escape(description)}</p><h2>Before you start</h2><p>Use a trusted {html.escape(host)} computer and {html.escape(client)} on the same reachable private network. The computer must remain awake. Use one folder containing only files you intend to share.</p><h2>Setup and download</h2><ol class="article-steps"><li>{html.escape(first)}</li><li>Open the local dashboard URL printed in the app console. Keep that dashboard URL private; its token authorizes local management. Use the separately displayed phone URL on your receiving device.</li><li>Verify the public certificate fingerprint against the dashboard through a trusted channel. Install the local certificate using your device’s trust settings. iPhone/iPad users also enable it under Certificate Trust Settings.</li><li>Open the phone URL, enter the generated password, and tap a file. The browser’s download manager shows the result. Search loaded entries or choose Load more files to browse additional pages.</li><li>Stop the app when finished. Existing transfers have up to 30 seconds to finish during shutdown.</li></ol><h2>Details for this workflow</h2><p>{note}</p><p><a href="../install.html">Choose the executable for your computer</a> · <a href="../troubleshooting.html">Connection and download help</a></p><h2>Why this avoids a cloud-upload workflow</h2><p>The app streams from your computer directly over the selected local interface. It includes no cloud relay, analytics, or account system. There is no upload or write endpoint on the receiving device’s file browser.</p><h2>If the connection fails</h2><p>Use the exact HTTPS address and port. Check client isolation on guest Wi-Fi, the selected interface, sleep settings, and a narrowly scoped firewall rule. If the browser rejects the certificate, complete pairing on that receiving device; trust on a computer does not automatically transfer to a phone.</p><h2>What to expect from this pre-release</h2><p>Portable builds target Windows, macOS, and Linux x86-64/ARM64. Windows x64 and Ubuntu 24.04 x64 under WSL 2 passed local runtime checks; macOS, other architectures, and physical Linux/Wi-Fi use still require device validation. Binaries are unsigned and unnotarized. IPv6-only networks, mobile server hosting, remote internet sharing, and phone-to-computer uploads are outside this version’s scope.</p><p class="note">The app has no file-size cap, but free storage, browser behavior, filesystem limits, and network conditions apply. A directory scan stops at 10,000 entries. Certificate pairing and local-folder selection are security decisions you make yourself.</p><a class="button" href="../index.html#how">Back to the setup overview ↗</a></main>' + footer('../')
        (SITE / 'guides' / (slug + '.html')).write_text(page, encoding='utf-8', newline='\n')
    build_support_pages()
    print(f'Generated {len(GUIDES) + 3} static content pages. Nothing deployed.')

if __name__ == '__main__':
    build()
