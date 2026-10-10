# Platform and validation matrix

The application supports modern desktop hosts and phone/tablet browsers. “Cross-platform” does not mean every historical OS, CPU, or mobile OS can host the server.

| Host target | Executable/archive | Runtime baseline | Current validation |
| --- | --- | --- | --- |
| Windows amd64 | `.exe` / `.zip` | Windows 10+ / Server 2016+ | Automated suite and isolated executable smoke check on Windows amd64 |
| Windows arm64 | `.exe` / `.zip` | Compatible Windows ARM64 | Cross-compiled; native smoke check pending |
| macOS amd64 | console binary / `.zip` | macOS 12+ with Go 1.26 builds | Cross-compiled; native smoke check pending |
| macOS arm64 | console binary / `.zip` | Apple Silicon, macOS 12+ | Cross-compiled; native smoke check pending |
| Linux amd64 | console binary / `.tar.gz` | Supported Linux distribution; Go runtime kernel baseline 3.2+ | Automated suite and executable HTTPS/download smoke check on Ubuntu 24.04 under WSL 2; physical LAN check pending |
| Linux arm64 | console binary / `.tar.gz` | Modern ARM64 Linux | Cross-compiled; native smoke check pending |

The Go 1.26 baseline preserves macOS 12 support. Building with a newer Go version can raise OS requirements. See [Go's minimum requirements](https://go.dev/wiki/MinimumRequirements). Linux builds disable cgo, so they do not require a dynamically linked C runtime.

Ubuntu, Debian, Linux Mint, Fedora, Arch, openSUSE, and Alpine use the corresponding Linux executable. See [the Linux/Ubuntu guide](LINUX.md) for setup, distro compatibility, firewall behavior, and WSL boundaries. Only Ubuntu 24.04 under WSL 2 has been executed locally so far; the other distributions are not individually validated.

Native CI covers Windows x64/ARM64, Intel/Apple Silicon macOS, Ubuntu 22.04/24.04 x64, and Ubuntu 24.04 ARM64. Runner labels follow [GitHub's hosted runner matrix](https://docs.github.com/en/actions/reference/runners/github-hosted-runners). The table records local validation before publication; see [current native CI results](https://github.com/dibyapp/pc-to-iphone-share/actions/workflows/ci.yml) for hosted checks. Physical phone/LAN testing remains separate.

## Receiving devices

| Client | Support scope |
| --- | --- |
| iPhone / iPad | Safari file browser, manual local-certificate trust, Safari/Files downloads |
| Android phones / tablets | Browser file downloads; manual user-CA trust must be supported by the device/browser |
| Desktop browsers | File browser works after local-certificate pairing |

There is no native iOS/Android server or App Store package in this release. Mobile devices are clients. IPv6-only networks, 32-bit hosts, and unsupported operating systems are outside the binary support matrix. Additional Go ports may be investigated, but compilation alone is not a claim of support.

## Before a stable release

Run the automated suite and a real browser/download check on every advertised host architecture. Check certificate pairing, Unicode names, a zero-byte file, a range download, pagination, sleep/wake, network changes, and a firewall-restricted connection. Record results without personal filenames or network details. macOS signing/notarization and Windows code-signing are still unconfigured.
