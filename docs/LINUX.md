# Linux and Ubuntu: download computer files on your phone

PC to iPhone Share's Linux executable is a portable, pure-Go binary. Ubuntu uses the same Linux archive as Debian, Fedora, Linux Mint, Arch, openSUSE, and Alpine for the matching processor. No Go, Python, Node, glibc, or package-manager installation is needed to run it. Linux distributions are compatibility targets; the validation status below distinguishes tested environments from untested ones.

| Distribution/environment | Archive | Validation |
| --- | --- | --- |
| Ubuntu 24.04, x86-64, WSL 2 | `linux-amd64.tar.gz` | Automated suite and isolated HTTPS/download smoke check executed locally |
| Ubuntu 22.04 / 24.04 desktop, x86-64 | `linux-amd64.tar.gz` | Native CI prepared; physical desktop/Wi-Fi check pending |
| Ubuntu ARM64 | `linux-arm64.tar.gz` | Native ARM64 CI prepared; local hardware check pending |
| Debian / Linux Mint / Fedora / Arch / openSUSE, x86-64 or ARM64 | Matching Linux archive | Same runtime requirements; individual distribution checks pending |
| Alpine, x86-64 or ARM64 | Matching Linux archive | Static binary does not need glibc; runtime check pending |

Use a supported distribution release and an ordinary desktop filesystem. The Go 1.26 runtime requires Linux kernel 3.2+; that runtime minimum is not a recommendation to use an obsolete distribution. See [Go minimum requirements](https://go.dev/wiki/MinimumRequirements).

## Run the portable app

Check the CPU with `uname -m`: `x86_64` means amd64; `aarch64` means arm64. Download the matching archive after a public release is approved, or build the local artifacts using [RELEASING.md](RELEASING.md).

For the current development archive:

```sh
tar -xzf pc-to-iphone-share-0.1.0-alpha.1-linux-amd64.tar.gz
./pc-to-iphone-share --version
./pc-to-iphone-share
```

Extract into a dedicated app directory outside the folder you plan to share. Keep the terminal open. Choose the folder and the actual private Wi-Fi/LAN interface reachable by your phone. Open the printed local dashboard on the computer, pair the phone's certificate using [CERTIFICATES.md](CERTIFICATES.md), then open the separate phone URL and sign in.

If execution is denied, check file permissions and whether the destination is mounted `noexec`. The packaged executable has execute permission; if your archive tool loses it, `chmod u+x pc-to-iphone-share` restores it. Run as your ordinary user. The default port 9443 requires no root privileges.

## Network, firewall, and desktop behavior

The app does not change UFW, firewalld, nftables, router settings, or certificate stores. If your firewall blocks downloads, allow inbound TCP on the selected app port only from your intended local subnet and interface. Do not disable the firewall or expose the port through the router. Guest-network client isolation can prevent access even when both devices use the same Wi-Fi name.

Keep the computer awake during downloads. A terminal launcher works on desktop and headless Linux; the local dashboard needs a browser on the host. Private state uses `$XDG_CONFIG_HOME/PCToIPhoneShare`, or the OS configuration-directory fallback, with user-only directory/file permissions. Keep it outside the shared folder and source repository.

WSL 2 was used for local Linux runtime validation with synthetic data on loopback. WSL networking and Windows firewall behavior differ from a physical Ubuntu desktop: this test does not establish phone reachability through WSL's NAT. For sharing Windows files directly, the Windows executable avoids that extra networking layer. WSL 1 is outside the Go runtime support baseline.

## Validate a release on Linux

Contributors with Python 3 can test the actual matching Linux archive:

```sh
python3 scripts/test_smoke.py
python3 scripts/smoke_binary.py dist/pc-to-iphone-share-0.1.0-alpha.1-linux-amd64.tar.gz
```

This creates temporary synthetic files and certificate state, listens only on loopback, verifies TLS/login/download/range behavior, and stops only its own process. It does not use an existing share or production password. Physical-device certificate pairing, LAN access, and sleep/wake checks remain separate checks in [PLATFORMS.md](PLATFORMS.md).
