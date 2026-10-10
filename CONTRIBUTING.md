# Contributing to PC to iPhone Share

Thank you for helping make local file sharing easier to use and easier to inspect.

## Development

Use Go 1.26 or newer. Application code and certificates use the standard library only. Browser assets are embedded during compilation. Python 3 is required only for release/privacy tooling, and Node is optional for JavaScript syntax checks.

```sh
go fmt ./...
go test ./...
go vet ./...
python scripts/privacy_check.py
```

Use a temporary folder containing synthetic files for manual testing. Store private test state outside the repository. Never test against your personal documents, production certificate directory, or an existing sharing process. Prefer an isolated loopback listener on a separate port.

## Pull requests

Describe the problem, resulting behavior, relevant limitations, and how you verified the change. Keep changes focused. Security changes should include a regression test that demonstrates the protected boundary, rather than just matching implementation details. Accessibility and layout changes should include a small viewport check and keyboard check.

Run `python scripts/privacy_check.py` before opening a pull request. Do not commit passwords, certificate material, screenshots containing real filenames or addresses, access notes, runtime configuration, logs, or machine-specific paths. Use synthetic examples. Build with `-trimpath -buildvcs=false` so executables do not embed local source paths or VCS metadata.

## Useful contributions

- Native checks on macOS Intel/Apple Silicon and Linux x86-64/ARM64.
- Browser download behavior on current Safari and Android browsers.
- Documentation that makes certificate pairing understandable without reducing HTTPS protections.
- Accessibility, empty/error states, and navigation through large directories.
- Security reviews of `os.Root` containment, state-directory permissions, and session boundaries.

Changes that add internet exposure, automatic certificate trust, write access, automatic telemetry, shell execution through a web endpoint, or account systems need a design discussion first. A change should not unexpectedly broaden the folder or network being shared.

## Community

Be respectful, concrete, and open to correction. Critique changes, not people. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
