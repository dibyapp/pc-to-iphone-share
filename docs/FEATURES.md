# File selection and temporary sharing controls

Version **0.2.0-alpha.1** adds these 15 features to the original read-only HTTPS workflow. The application still needs one reachable local network and manual certificate pairing. It does not automatically create a hotspot or promise a transfer rate.

| Addition | Why it helps | Boundary and automated coverage |
| --- | --- | --- |
| Folder-wide search | Finds filenames beyond the currently displayed page | Case-insensitive substring search, current folder only, before pagination; integration test uses more than 200 entries |
| File-type filters | Separates photos, videos, audio, documents, archives and other files | Extension-based, not content recognition; matching folders stay navigable; filter tests |
| Sorting | Finds the newest or largest files | Name, size and modified, ascending/descending; folders first; sort and tie tests |
| Page-size controls | Balances small screens and large collections | UI: 50/200/500; API: 1–500; bad-input and pagination tests |
| Multi-file ZIP | Downloads a selection in one browser action | 1–64 regular files, up to 4 GiB total; streamed, uncompressed ZIP; no recursive folders or ZIP resume; CRC/byte, empty-file, unsafe-path and duplicate tests |
| Text previews | Checks a document before downloading | UTF-8 txt/md/csv/json/log/yaml/yml/xml only; 256 KiB, rendered as text; binary rejection and Unicode-boundary tests |
| SHA-256 | Compares the computer's file hash with a downloaded copy | Up to 1 GiB, 30 seconds, shared operation cap; detects size/mtime changes; integrity and containment tests; not malware scanning |
| Local speed test | Helps distinguish connection throughput from file behavior | 16 MiB synthetic data, one test per 30 seconds for the whole share; displayed MB/s and Mbps; bounded-size/rate/counter tests |
| Transfer statistics | Shows activity without keeping a file history | Loopback dashboard only; bytes written, successful download responses, active operations and live sessions; privacy/count tests |
| Pause/resume | Temporarily stops new file access | Existing requests may finish; signed-in browsing/preview/hash/download/test requests get 503 while paused; access-control tests |
| Sign out all devices | Revokes current browser sessions | Active requests may finish. Password is unchanged and devices may sign in again; revoke/relogin/CSRF tests |
| Automatic expiry | Limits an unattended sharing window | `--share-for 30m`; maximum 24h, zero disables; new private requests denied after deadline; process allows existing streams up to 30 seconds during shutdown; deadline and executable tests |
| Transfer limits | Avoids too many simultaneous heavy operations | `--max-downloads 4`, range 1–16; ordinary downloads, ZIPs, checksums and speed tests share the cap; 429/Retry-After and recovery tests |
| Exclusions | Keeps named files or directories out of the selected share | Repeat `--exclude`; filename globs, case-insensitive, every path component/depth; not a secret detector; all file endpoints tested |
| Setup check | Catches local configuration mistakes before sharing | `--check` validates folder readability, private state separation and interface/options without starting listeners or saving a password/certificate; it may create/protect the private settings directory; CLI tests |

The folder scan inspects at most the first 10,000 directory entries. Searches are not recursive, filters operate on filenames, and directory changes during pagination can change results. Refresh after changing files. ZIP selections are cleared on a fresh listing or folder change. Individual streamed downloads keep byte-range support and have no application file-size cap.

## Practical examples

Run the included executable in a terminal. Use `pc-to-iphone-share.exe` on Windows.

```sh
./pc-to-iphone-share --check
./pc-to-iphone-share --share-for 30m --max-downloads 2 --exclude '*.key' --exclude 'private'
```

Folder and interface selection remain interactive unless supplied explicitly. Runtime options apply only to that invocation; they are not saved. Apply your exclusions every time you start sharing. Patterns match individual names (`*.key`, `private`, `backup-*`), not slash-separated paths. Quoting prevents your shell from expanding the pattern first. Excluding a directory also excludes its descendants.

In the phone browser, adjust search/type/sort/page size, select files, then use **Download ZIP**. **Preview** displays text without interpreting markup. **SHA-256** reads the selected file on the computer; compare the returned digest using your device's trusted checksum tool. A changing file cannot be guaranteed to produce a consistent snapshot; keep files stable during downloads and hashing.

Open **Check local transfer speed** and run the synthetic test. This measures the browser/TLS/local-network path, not internet speed or storage performance. The test consumes local network capacity, so avoid running it during important downloads.

On the computer dashboard, **Pause new access** blocks new file operations while leaving active requests alone. **Sign out all devices** clears sessions; to keep devices out, pause or stop sharing. For a lost password, stop and restart with `--reset-password`. Session revocation alone does not rotate the password.

## Validation and limitations

Go integration tests use isolated temporary files and verified loopback TLS. Native executable smoke tests cover download/range integrity, the new file operations and host controls. A local automated browser check exercises the phone-width UI and dashboard with synthetic data. Hosted CI runs on Windows x64/ARM64, macOS Intel/Apple Silicon, and Linux x64/ARM64; see the [exact commit's checks](https://github.com/dibyapp/pc-to-iphone-share/actions/workflows/ci.yml) for results. A configured CI matrix does not by itself prove success.

Physical iPhone/Android certificate pairing, native Wi-Fi throughput, every Linux distro, signing and notarization are separate checks. No independent security audit is claimed. ZIP errors after streaming starts terminate the response rather than silently returning a valid-looking partial archive. Transfer counters represent HTTP responses, not proof that a receiving device saved a file; a resumed range counts as a separate successful response. Checksums and synthetic tests use operation slots but do not inflate file-transfer totals.
