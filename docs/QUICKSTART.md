# Secure local file sharing: quickstart

## Choose the right executable

Use the archive for your operating system and CPU. Most PCs use amd64 (x86-64); Apple Silicon and many modern ARM PCs use arm64. See [PLATFORMS.md](PLATFORMS.md). Extract the archive before running the app. Keep the executable outside the folder you share.

On Windows, run `Start-PC-to-iPhone-Share.cmd` or `pc-to-iphone-share.exe`. On macOS, run the binary from Terminal; the archive also includes `Start-PC-to-iPhone-Share.command`. On Linux, run `./pc-to-iphone-share` from a terminal. These are portable console executables, not background services or mobile apps.

Ubuntu, Debian, Linux Mint, Fedora, Arch, openSUSE, and Alpine use the matching Linux archive; see [LINUX.md](LINUX.md). Intel Macs use amd64; Apple Silicon Macs use arm64. No platform needs an extra application runtime to execute a portable build.

The builds are unsigned pre-release binaries. Use source builds if you need to establish your own provenance. Signing/notarization remains a release gate; the app does not bypass operating-system protections.

## Choose a folder and network

Paste or drag a folder path into the app console. If more than one network is listed, choose the Wi-Fi or LAN interface shared with your phone. Avoid VPN, virtual-machine, and container interfaces unless you intend to share with them.

The app binds only that interface, on port 9443 by default. If another program uses that port, choose a different unprivileged port with `--port`. Existing applications are never stopped or modified.

## Pair and download

Open the printed local dashboard URL on the computer. Its fragment contains a private local access token; do not share the dashboard URL. Share only the separately displayed **phone URL** with your own device. No passwords or dashboard tokens are embedded in the phone URL.

The dashboard displays the newly generated password once. Save it in your own password manager if desired. It is not stored in plaintext by the app. On later runs the password stays unchanged; if you lose it, restart with `--reset-password`. This creates a new password and revokes sessions.

Join the same network on your phone. Follow [CERTIFICATES.md](CERTIFICATES.md) to pair the device, then enter the password. Tap a filename to download. Look in Safari Downloads or your Android browser's download manager. The app's search filters loaded entries; use **Load more files** to search additional pages.

## Stop and restart

Ctrl+C or **Stop sharing** stops the app. Active downloads have up to 30 seconds to finish before connections close. A process restart invalidates browser sessions.

Keep your computer awake during transfers. If its address changes, restart and use the newly printed phone URL. The same local CA is reused, so a paired device does not need a new profile for an ordinary address change. Server certificates renew on each app start.

Private state lives in your OS user configuration directory under `PCToIPhoneShare`, unless `--state` specifies a dedicated private directory. Never place state in the selected share or in the repository. Uninstalling means removing the portable executable and, if desired, its local state and the certificate profile from paired devices; no system service is installed.
