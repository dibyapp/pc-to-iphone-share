# Product positioning and search content

## Core promise

PC to iPhone Share makes computer-to-phone downloads possible over a local network, with password protection and encrypted transport, while keeping the selected files out of cloud storage. It shares one folder and requires no phone app or account.

Avoid claims such as “100% secure,” “works on every OS,” “fastest file transfer,” “unlimited speed,” “instant certificate setup,” or “audited encryption.” The implementation uses standard TLS, but the app has not received an independent audit. Network and browser limits still apply.

## Pain points to address

| Need | Product response | Honest limit |
| --- | --- | --- |
| Download a PC file to an iPhone without a cable | Browser access over a chosen local interface | Same reachable network and one-time certificate pairing required |
| Avoid sending a private file through cloud storage | Direct computer-to-device HTTPS download | A trusted computer, browser, and first-time trust channel are still needed |
| Move a large file without an email attachment workflow | Streamed file download and byte-range support | Device storage, network speed, and resume support vary |
| Use different desktop and phone ecosystems | Windows/macOS/Linux hosts and iPhone/Android browser clients | Native validation and current trust policies matter |
| Limit temporary sharing | One folder, read-only access, stop control | Deliberately selected folder contents remain the owner's responsibility |

## Search intent and page map

| Search intent | Prepared content |
| --- | --- |
| Secure Wi-Fi file transfer / local file sharing without cloud | `site/index.html`, README |
| Transfer files from Windows PC to iPhone over Wi-Fi | `site/guides/windows-to-iphone.html` |
| Download files from Linux to iPhone without USB | `site/guides/linux-to-iphone.html` |
| Transfer files from Ubuntu to iPhone over Wi-Fi | `site/guides/ubuntu-to-iphone.html`, Linux/Ubuntu setup documentation |
| Download files from Mac to Android over local Wi-Fi | `site/guides/mac-to-android.html` |
| Password-protected local file sharing | `site/guides/private-local-sharing.html` |
| Certificate warning, guest Wi-Fi, download failures | Troubleshooting and certificate documentation |
| Download/install the right Windows, macOS, Ubuntu/Linux build | `site/install.html`, `docs/INSTALL.md` |
| Safari certificate/cross-site and local connection help | `site/troubleshooting.html` |

The static pages include unique titles and descriptions, descriptive headings, FAQ copy, responsive styling, and software/FAQ structured data. They use no analytics, remote fonts, or personal screenshots. The owner approved `dibyapp/pc-to-iphone-share` and its alpha release, so source/download links point there. Canonical and sitemap domains remain unset because the content site is not deployed. The real interface screenshots use synthetic fixture data only. See [SEO-RESEARCH.md](SEO-RESEARCH.md), [KEYWORDS.csv](KEYWORDS.csv), and [GROWTH-PLAN.md](GROWTH-PLAN.md) for evidence, priorities, and the first-month experiment.

## Suggested public copy

**Repository description:** Secure, open-source computer-to-phone file downloads over local Wi-Fi. Windows, macOS, and Linux hosts. iPhone and Android browsers. No cloud uploads.

**Launch headline:** Download your computer's files on your phone — without sending them to the cloud.

**Short launch post:** PC to iPhone Share is an MIT-licensed local file browser for computer-to-phone downloads. Choose one folder, pair your phone with your computer's local certificate, and sign in with a generated password. Transfers use HTTPS over your selected local network. Portable desktop builds and readable source keep setup inspectable. This first version is read-only, IPv4-only, and a pre-release; native platform validation and signing status are documented openly.

**Candidate topics:** local-file-sharing, wifi-file-transfer, iphone, android, windows, macos, linux, ubuntu, privacy, https, self-hosted, golang, open-source.

No launch content has been posted, sent, or published.
