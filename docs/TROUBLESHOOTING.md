# Connection and download troubleshooting

## The site cannot be reached

Use the printed phone URL exactly, including `https://` and the port. Plain HTTP cannot talk to the HTTPS listener. The phone and computer must be on a network that permits device-to-device traffic. Guest Wi-Fi often isolates clients.

Check that the app is still running, the computer is awake, and you selected the intended Wi-Fi/LAN interface. If its address changed, restart the app and use the newly printed URL. No other sharing server is stopped by this app.

## The certificate is not trusted

Complete [certificate pairing](CERTIFICATES.md) on each receiving device. Trusting a certificate on a computer does not trust it on a phone. Restart a browser that caches old trust. Do not turn off certificate verification.

## Cross-site requests are blocked

The public login page and certificate download accept entry from another app. The private file API, login/logout requests, and downloads still reject cross-site use. Open the phone URL directly in Safari or your regular browser, then sign in from that page. Downloads linked from an unrelated website are intentionally blocked. In-app webviews with opaque origins may not support this security model.

## Windows firewall blocks the phone

Ask for a narrow inbound TCP rule for the selected executable, selected private interface, chosen port, and local subnet. Do not disable the firewall or allow all networks. The app never creates a rule automatically. A VPN or virtual interface may route traffic differently; choose the interface shared with your phone.

## The folder is empty or some files are missing

Refresh after adding files. Dot-prefixed paths, links/junctions, invalid Windows names, and non-regular files are excluded. Non-dot hidden files may still be listed. Use a clean, dedicated share folder. Click **Load more files** for additional pages. A scan stops at 10,000 directory entries; split larger collections into subfolders.

## A large file is slow or fails

The app streams files without a fixed app-level size cap. Actual speed depends on Wi-Fi signal, disk performance, concurrent downloads, the browser, and free storage. Keep the computer awake and close competing transfers if appropriate. HTTP ranges are supported, but browser resume behavior varies. Stopping the app gives active transfers at most 30 seconds to finish.

## The password is lost or a login is throttled

Restart with `--reset-password` to generate a new password and revoke sessions. New passwords are shown only in the local console/dashboard. There is no plaintext password file. Login limits are per peer and global; wait fifteen minutes if throttled. Password reset should be performed only from the computer by its owner.

## The binary is unsigned

This is an unpublished pre-release. Code signing and notarization are not yet configured. Check SHA-256 integrity against the provided manifest and evaluate provenance separately. Do not infer publisher authenticity from a checksum supplied by the same untrusted source. Build from reviewed source if required by your security policy.
