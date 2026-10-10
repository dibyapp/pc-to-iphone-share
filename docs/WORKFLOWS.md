# Everyday workflows and contributor workflows

## Computer folder → iPhone

1. Install the portable app on Windows, macOS, or Linux, keeping it outside your shared folder.
2. Start it, choose one folder, and select the private interface connected to the same reachable network as the iPhone.
3. Open the computer's private local dashboard. Keep that URL private; the separate phone URL is the one to use on the phone.
4. Follow [certificate pairing](CERTIFICATES.md), verifying the fingerprint through a trusted channel. iPhone users install the profile and explicitly enable full trust for their own local CA. [Apple's instructions](https://support.apple.com/en-us/102390).
5. Open the phone URL in Safari, enter the password, and tap a file. Find it in Safari Downloads or Files.
6. Keep the computer awake until the transfer completes. Stop sharing when finished.

![Phone file browser with synthetic example files](screenshots/folder-phone.jpg)

The screenshot is a browser capture of the shipped interface with synthetic fixture responses; it is not a live iPhone OS screenshot or proof of a physical-device test.

## Windows/Mac/Linux → Android

Use the same host steps, then install the public certificate through the Android device's supported security settings. Manufacturer menus and browser trust behavior differ. This project does not bypass device policy. Open the phone URL, sign in, and use the browser's download manager. If the browser does not accept a user CA, choose a compatible managed trust workflow or another transfer tool.

## Transfer a large file

Choose a folder containing the intended file; ensure the receiving device has enough space. Downloads stream and support HTTP byte ranges. Resume behavior depends on the browser/download manager. A network change, computer sleep, or shutdown can interrupt a transfer. The app does not impose a file-size cap, but it does not promise unlimited speed or resumability on every client.

## Temporarily share reference documents

Copy only the intended files into a dedicated folder; do not select your home, Downloads root, private configuration, or source checkout. The app has no upload/delete/rename endpoints. Share only with devices you trust on the intended network; stop it afterwards. See [SECURITY.md](../SECURITY.md) for filesystem and local-administrator boundaries.

## Develop without touching an existing share

```sh
go test ./...
go vet ./...
python scripts/test_privacy.py
python scripts/test_install.py
python scripts/test_smoke.py
python scripts/test_site.py
python scripts/smoke_binary.py --build
```

The smoke check uses synthetic files and private state in temporary storage, listens on loopback, verifies TLS/authentication/downloads, and stops only its own process. Screenshot fixtures are in `scripts/demo_preview.py`; they are not embedded in the production executable and never open a shared folder.

## CI and release

CI validates source on pushes to `main`, pull requests, and manual dispatch. Native runners cover Windows x64/ARM64, Intel/Apple Silicon macOS, and Ubuntu x64/ARM64. Artifact builds are manual. Public releases require a deliberate maintainer action; neither CI nor the build workflow deploys the website or restarts a sharing app. See [RELEASING.md](RELEASING.md).
