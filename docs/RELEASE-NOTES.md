# PC to iPhone Share v0.2.0-alpha.1

Download one computer folder on an iPhone or Android browser over local Wi-Fi. Password-protected HTTPS, read-only access, and no cloud upload or phone app.

## Fifteen additions in this release

1. Search the current folder before pagination, including filenames beyond the loaded page.
2. Filter photos, videos, audio, documents, archives, folders and other files.
3. Sort by name, size or modification date, ascending or descending.
4. Choose 50, 200 or 500 items per browser page.
5. Download a selection as a streamed ZIP (up to 64 files/4 GiB, no recompression).
6. Preview UTF-8 text safely, up to 256 KiB.
7. Calculate SHA-256 for files up to 1 GiB with a 30-second budget.
8. Measure the local connection with a 16 MiB synthetic speed test.
9. See in-memory transfer and session totals on the computer dashboard.
10. Pause and resume new file access without restarting.
11. Sign out all devices from the dashboard.
12. Stop automatically with `--share-for`, up to 24h.
13. Limit heavy simultaneous operations with `--max-downloads` (1–16; default 4).
14. Hide filename patterns at every depth with repeated `--exclude` options.
15. Validate local setup without starting listeners using `--check`.

Search is not recursive and the folder scan remains capped at 10,000 entries. Pause and session revocation
allow already authorized requests to finish; revocation does not change the password. Exclusions and runtime
limits must be passed on each launch. Individual downloads retain byte-range support and have no application
file-size cap. ZIP downloads do not support resume. [Detailed boundaries and tests](https://github.com/dibyapp/pc-to-iphone-share/blob/main/docs/FEATURES.md).

Also rejects special files before opening and uses nonblocking Unix opens to avoid a concurrently swapped
FIFO blocking a handler. Built with Go 1.26.10. No new third-party Go dependency.

**This is an unsigned, unnotarized alpha.** Windows x64 and Ubuntu 24.04 x64 under WSL 2 passed local runtime checks. Hosted native CI covers every packaged CPU/OS target; current results are available in [Actions](https://github.com/dibyapp/pc-to-iphone-share/actions/workflows/ci.yml). Physical phone/LAN pairing and distro-specific tests remain separate, and no independent security audit is claimed.

## Download and start

Choose the Windows/macOS/Linux archive for x64 (`amd64`) or ARM64. Verify its exact filename in `SHA256SUMS`, extract into an app folder, and run the executable or included launcher. No Go, Node, Python, administrator installation, or package manager is required for a portable binary.

The release contains six desktop archives, one allowlisted source ZIP, `SHA256SUMS`, and `BUILD-INFO.json`. macOS Intel and Apple Silicon are separate builds; Ubuntu and other Linux distributions use the corresponding Linux archive.

See [INSTALL.md](https://github.com/dibyapp/pc-to-iphone-share/blob/main/docs/INSTALL.md) for download choices, extraction, optional checksum-verifying installer, and source builds.

## First connection

Choose one folder and your private network interface. Keep the local management dashboard URL private; use the separately displayed phone URL on the receiving device. First-time [certificate pairing](https://github.com/dibyapp/pc-to-iphone-share/blob/main/docs/CERTIFICATES.md) is manual: verify the fingerprint and install only your own computer's public CA. Sign in with the generated password, download, then stop sharing when finished.

## Included controls

- Selected-interface/subnet binding and TLS 1.2+ with per-installation constrained certificates.
- Salted password hashing, expiring sessions, login throttling, and protected same-origin APIs.
- Confined read-only file access, linked-path rejection, streamed attachment downloads, byte ranges, and pagination.
- Loopback management dashboard; no telemetry, request logging, automatic updates, uploads, or remote-sharing service.
- Synthetic UI screenshots, platform guides, troubleshooting, contributor workflows, and MIT-licensed source.

IPv6-only networks, 32-bit hosts, native mobile server apps, phone-to-computer uploads, automatic device discovery, and remote internet sharing are outside this alpha's scope. Do not disable OS/browser security protections to run it. [Threat model and private vulnerability reporting](https://github.com/dibyapp/pc-to-iphone-share/blob/main/SECURITY.md).

Author and maintainer: [dibyapp](https://github.com/dibyapp). No co-author attribution is added. No private files, passwords, certificates, runtime settings, logs, or machine-specific launch configuration are included in the release. The content site remains undeployed.
