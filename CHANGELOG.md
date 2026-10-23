# Changelog

## 0.2.0-alpha.1 — sharing controls and faster file selection

1. Server-side, case-insensitive search of the current folder before pagination.
2. File-type filters, retaining navigable matching folders.
3. Name, size and modification-date sorting in both directions, folders first.
4. Configurable page size: UI choices 50/200/500; API accepts 1–500.
5. Multi-file streamed ZIP downloads: up to 64 files and 4 GiB, without recompression.
6. Escaped UTF-8 text previews bounded to 256 KiB.
7. SHA-256 file checksums up to 1 GiB with a 30-second calculation budget.
8. Local connection speed test using 16 MiB of synthetic data; no external service.
9. Dashboard transfer counters and active-session totals, held only in memory.
10. Pause/resume of new authenticated file access from the local dashboard.
11. Revoke all signed-in sessions without restarting the process.
12. Optional automatic sharing expiry with `--share-for` (maximum 24h).
13. Configurable simultaneous transfer/checksum/speed limit with `--max-downloads`.
14. Repeated, case-insensitive filename exclusion rules with `--exclude`.
15. Setup validation without starting listeners using `--check`.

Also: reject special files before opening them; use nonblocking Unix opens against FIFO swaps;
validate query inputs; keep client text out of HTML; update the Go build patch to 1.26.10.
No cloud service, automatic certificate trust, upload endpoint, background service or website deployment added.
See [feature boundaries and validation](docs/FEATURES.md). Builds remain unsigned and unnotarized.

## 0.1.0-alpha.1 — unsigned alpha

- Portable Go app and embedded mobile file browser.
- Windows, macOS, and Linux x86-64/ARM64 build targets.
- Interactive folder and network selection with a local setup dashboard.
- Per-installation constrained certificates and manual device pairing.
- Password protection, session expiry, login throttling, and read-only folder confinement.
- Streamed downloads, byte-range support, Unicode filenames, and directory pagination.
- Privacy scanning, source allowlisting, stripped build paths, checksums, and manual build workflows.
- Search-friendly landing-page content and platform-specific transfer guides.
- Explicit Linux/Ubuntu setup guide and Ubuntu-to-iPhone search content.
- Ubuntu 24.04 WSL 2 runtime verification and native CI for all packaged architectures.
- Privacy audit of staged and older Git objects, archive filenames, and UTF-16 identifiers.
- Six portable downloads, checksum-verifying installer, synthetic UI screenshots, and setup workflows.
- Source-backed keyword research, qualitative keyword backlog, and an organic discovery experiment.
- GitHub owner attribution with a public noreply commit identity and no co-author trailers.

No public release, push, upload, or deployment is performed by the local build tools.
