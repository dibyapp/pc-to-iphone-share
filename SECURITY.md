# Security policy and threat model

PC to iPhone Share is an unsigned alpha, with no independent security audit or supported stable release yet. Please evaluate it according to the boundaries below rather than assuming it provides complete endpoint isolation.

## Intended use

A trusted user runs the app on a trusted computer, chooses one folder, and makes its files downloadable by password-authenticated devices on one selected private IPv4 subnet. The computer's OS account and the receiving device are trusted. The app provides no write access and no internet-sharing mode.

## Controls

- TLS 1.2+ and per-installation certificates. Certificate pairing is manual. The app never disables certificate verification or silently adds system trust.
- A constrained local CA with private-IP name constraints and a reserved DNS namespace. The CA private key stays in local private state. Client profiles contain only public certificates.
- Random passwords, salted PBKDF2-HMAC-SHA256 hashing (600,000 iterations), constant-time comparisons, and limits of five attempts per peer and thirty globally per fifteen minutes. Two password computations run concurrently.
- Server-held sessions with inactivity and absolute expiry. Logout and process restart revoke sessions. There is no remember-password feature.
- Exact Host and Origin checks, fetch metadata checks for private routes, restrictive CSP, frame blocking, attachment downloads, no remote assets, and no CORS permission.
- `os.Root` confines opened files even when directories are changed during traversal. Additional checks reject symlinks/junctions and invalid paths. Only regular files are downloaded.
- A loopback-only management dashboard with a random bearer token delivered in the URL fragment. It provides no unauthenticated access to the password or stop control.
- Unix private state uses a directory mode of 0700 and files of 0600. Windows removes inherited permissions and grants the current account and SYSTEM full access. Existing explicit ACL entries on a user-selected state directory should be reviewed; choose a newly created dedicated directory.
- A bounded number of connections, password computations, active sessions, failed-attempt records, and listed directory entries. Downloads stream instead of being loaded into memory.
- Configurable shared operation slots for downloads, ZIPs, checksums and speed tests. Exclusion patterns apply at every path component to all file endpoints. Unix nonblocking opens and special-file checks avoid serving FIFOs/devices.
- UTF-8 text previews use JSON and browser text nodes, never an HTML execution context. ZIP selections are confined regular files; checksums and synthetic tests have separate size/time/rate limits. See [feature boundaries](docs/FEATURES.md).
- Local authenticated dashboard controls pause new access and revoke sessions. Optional expiry denies new private requests at the deadline and shuts down with up to 30 seconds for active streams. None of these controls retroactively deletes a downloaded copy.

## Boundaries and residual risks

**First-time trust:** A hostile network can interfere with the first certificate download. Verify the certificate fingerprint using the local dashboard through a trusted channel. Downloading a root certificate and blindly trusting it is not identity verification. Remove the profile when the app is no longer used.

**Subnet membership:** A private subnet is not proof that a device is physically on the same Wi-Fi. VPNs, router forwarding, and deliberately bridged networks can extend reachability. No firewall rule is installed automatically; firewall policy is an additional boundary.

**Local access:** Administrators and processes running as the same user can read private state or alter files. This app cannot protect secrets from a compromised computer. A manually broadened state-directory ACL is outside the default setup guarantee.

**Folder contents:** Hard links and mounted filesystems are not the same as symlinks. A regular file reachable within the selected folder is shareable even if another name or mount also refers to its data. Choose a dedicated folder populated with files you intentionally want to share. Non-dot hidden files may appear; use a clean share folder. Rejecting links is not a substitute for reviewing folder contents.

**Availability:** Peers can exhaust login limits or consume bandwidth and disk I/O. Limits reduce resource abuse; they do not provide denial-of-service immunity. Individual downloads have no application file-size cap; ZIP selections are limited to 64 files/4 GiB, browser checksums to 1 GiB/30 seconds, text previews to 256 KiB, and synthetic speed tests to 16 MiB/one per 30 seconds for the share. A directory scan stops at 10,000 entries. Pause/revocation affect new requests; already authorized requests can finish. Revoking sessions leaves the password unchanged. Exclusions apply only when passed at launch and do not detect secrets automatically.

**Certificates and signing:** Local CA certificates expire after five years; server certificates are regenerated at each start and expire after 90 days. Restart a long-running instance before its server certificate expires. Portable builds are currently unsigned and unnotarized. SHA-256 checksums detect corruption but do not prove publisher identity.

**IPv6 and unsupported OSes:** Network serving is IPv4 only in this release. The local dashboard is IPv4 loopback. Unsupported architectures and operating systems require separate porting and native validation.

## Reporting a vulnerability

Do not post passwords, private keys, real file contents, local configuration, or personal network details in a public issue. Use synthetic reproduction steps.

Use [GitHub private vulnerability reporting](https://github.com/dibyapp/pc-to-iphone-share/security/advisories/new). If that private channel is unavailable, hold sensitive details and request a reporting channel with only a redacted, non-sensitive description. Never attach real configuration, passwords, keys, certificate material, addresses, or personal files. The current version is an alpha; fixes should be validated before assuming protection.

## Review priorities

The highest-priority review areas are confined file access across platforms, Windows state ACLs, first-time trust, dashboard isolation, rate-limit concurrency, and browser-specific download behavior. The automated suite covers authentication, Safari-style navigation, traversal, junctions, session expiry, range downloads, pagination, and dashboard access.
