# Screenshot provenance

All JPEGs in this directory are browser screenshots of PC to iPhone Share's actual HTML/CSS/JavaScript, rendered with the loopback-only `scripts/demo_preview.py` fixture. Files, dates, address, password text, and fingerprint text are synthetic. Every image has an on-screen demo label. No real share, personal path, certificate, login session, or production credential was captured.

- `folder-phone.jpg`: responsive file browser at a phone-width viewport.
- `login-phone.jpg`: responsive password sign-in page.
- `folder-desktop.jpg`: file browser at the browser's default desktop viewport.
- `dashboard-desktop.jpg`: setup dashboard with explicitly fake values.
- `installation-page.jpg`: local, undeployed installation content with the approved public GitHub release URLs.

These are browser captures, not physical iPhone/Android operating-system screenshots. Recreate them by running the demo script with `--screen browser`, `login`, or `dashboard`, opening its printed loopback URL, and using browser screenshot tooling. For the dashboard fixture use the fragment `#demo-only`; it is not an authentication secret. Do not capture the original private sharing application. The demo server must never be exposed to the LAN.
