# Alpha preparation and publication record

Repository: https://github.com/dibyapp/pc-to-iphone-share
Initial release: https://github.com/dibyapp/pc-to-iphone-share/releases/tag/v0.1.0-alpha.1
Author and maintainer: dibyapp, using a public GitHub noreply identity. No co-author trailers are added.

## Verification before the first push

- MIT application license and Go runtime notice; no third-party Go modules.
- Go test suites, vet, and isolated executable checks on Windows x64 and Ubuntu 24.04 x64 under WSL 2.
- Six portable CPU/OS archives, allowlisted source archive, normalized metadata, SHA-256 checksums, and stripped local build paths.
- Privacy scans of source filenames/content, staged blobs, all local Git objects, archive members, and binaries. The explicitly approved public GitHub handle is allowed; personal paths and private machine identity remain blocked.
- Original private app credentials/certificate contents compared locally against public deliverables without printing or packaging them.
- Installer tests for CPU selection, exact checksums, path containment, linked-member rejection, and refusal to overwrite an existing file.
- Synthetic screenshot fixtures only; real HTML/CSS/JS at phone and desktop widths, with demo labels. No production share was captured or restarted.
- Eight static content pages, descriptive metadata, useful OS notes, installation links, safe troubleshooting, and screenshot alt text.
- Source-backed search research, 66 qualitative keyword candidates including 7 excluded intents, and a 30-day organic discovery experiment. No volume/difficulty values or results are invented.
- Native CI configured for all six packaged architectures plus Ubuntu 22.04/24.04. Current hosted results are visible in the repository's Actions page; do not treat a prepared workflow as a passed run.

## Scope and remaining limitations

This is an unsigned alpha, not a stable audited product. Native hosted checks validate application runtime behavior; physical phone/LAN pairing, sleep/wake, firewall policy, and distro-specific checks remain separate. The original private sharing application and its downloads are outside this repository and are not modified for publication.

The owner authorized creating/pushing the public GitHub repository and publishing the unsigned prerelease. That authorization does not deploy the static website, submit promotional posts, send messages, purchase stars, or enable recurring monitoring. Those actions have not been performed. Private GitHub vulnerability reporting is configured during repository publication; no private owner email is placed in source or commit metadata.

Automated scans found no sensitive material in reviewed public deliverables. They do not guarantee detection of every credential format. All packages are portable console applications; iPhone/Android are clients, and not every historical OS/architecture is supported. See PLATFORMS.md, SECURITY.md, and INSTALL.md for boundaries.
