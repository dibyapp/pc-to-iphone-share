# Pair a device with the local HTTPS certificate

Every installation creates its own local certificate authority. Pairing lets a browser authenticate the computer's encrypted connection. A publicly trusted certificate for an arbitrary private LAN address is not included.

Verify the phone address and certificate SHA-256 fingerprint against the computer's loopback dashboard. Use a trusted channel for the first transfer of the certificate. The certificate profile contains one public certificate; it contains no private key, VPN configuration, or device-management payload.

## iPhone and iPad

1. On your own trusted network, open the phone URL in Safari. The first visit can show a certificate warning. You must handle that warning yourself after verifying the address; the app does not bypass it.
2. Use **First time on this iPhone?** on the login page to download the profile. Alternatively, save the profile from the computer's local dashboard and transfer that public file to the phone through a trusted channel.
3. Install **PC to iPhone Share Local Certificate** in Settings → General → VPN & Device Management, or through **Profile Downloaded**.
4. Open Settings → General → About → Certificate Trust Settings and enable full trust for that certificate.
5. Reopen Safari and sign in.

See [Apple's certificate trust instructions](https://support.apple.com/en-us/102390). Labels and profile handling may vary by OS release. Remove the profile when you stop using this installation.

## Android

Download the public `.cer` certificate from the app or local dashboard, then use the device's security settings to install a CA certificate. Manufacturers use different labels. Some browsers use their own trust store, and some managed devices disallow user-installed CAs. See [Google's certificate management guidance for Pixel](https://support.google.com/pixelphone/answer/2844832/add-amp-remove-certificates).

Verify the fingerprint before trusting the certificate. If your device does not allow this certificate to be trusted, do not disable TLS checks; use a device/browser policy that supports local certificate trust.

## Windows computer browser

Save the public certificate from the loopback dashboard. Import it into the **current user's Trusted Root Certification Authorities** store only after verifying its fingerprint. Windows asks you to confirm trust. Approve that prompt yourself. The app never silently imports certificates.

Close and reopen a browser that continues using stale trust information. Embedded browsers may have an independent trust store or lifecycle. The local dashboard remains available without changing system trust because it is bound only to loopback and requires a token.

## macOS and Linux browser

Use your OS or browser's certificate settings to import the verified public certificate. Some Linux browsers maintain their own certificate database. Corporate policies can prohibit local trust. Consult the browser's official certificate documentation and keep verification enabled.

## Expiry and reset

The root CA lasts five years. Server certificates are regenerated on each start and last 90 days. An ordinary address change reuses the same root CA. Creating a new private state directory creates a new identity and requires re-pairing. Never publish or send `ca.key`; only the public `.cer` or profile belongs on the receiving device.
