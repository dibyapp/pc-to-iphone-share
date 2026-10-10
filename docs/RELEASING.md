# Local release preparation — no publishing

Version 0.1.0-alpha.1 is prepared locally. Nothing in this procedure pushes a repository, creates a hosted release, uploads files, or deploys the content site.

## Reproducible inputs and private-data boundaries

Application builds require Go 1.26 and use only the standard library. Use a pinned patch version for a release and record it in `BUILD-INFO.json`. Python 3 generates archives/checksums from an explicit source allowlist. `-trimpath -buildvcs=false` strips local source paths and repository metadata; `-ldflags=-s -w` strips debugging data.

Public source includes only application code, embedded public web assets, developer tooling, documentation, explicitly synthetic screenshots, license files, and checked-in workflow definitions. Runtime configuration, passwords, private keys, profiles, logs, personal screenshots, user files, local toolchains, `.git`, and build output are never copied into the source archive.

```sh
go test ./...
go vet ./...
python scripts/privacy_check.py
python scripts/test_privacy.py
python scripts/test_install.py
python scripts/test_smoke.py
python scripts/test_site.py
python scripts/build_release.py --version 0.1.0-alpha.1
python scripts/privacy_check.py --artifacts dist
python scripts/smoke_binary.py --build
```

Use `--go /path/to/go` if the toolchain is not on PATH. The builder performs tests, vet, source privacy checks, cross-compilation, and artifact privacy checks. It does not install the app, trust certificates, change a firewall, or run a network listener. Checksums and normalized build information contain no machine paths or identity.

`smoke_binary.py --build` separately compiles the native target and starts it on loopback using temporary synthetic files and a temporary private state directory. It verifies HTTPS, login, downloads, ranges, and dashboard authentication, then stops only its own process. It never connects to or restarts an existing sharing app. To test an actual matching Windows/macOS ZIP or Linux TAR.GZ, supply its path instead of `--build`. It copies only the exact regular executable member; archive paths are never extracted.

The privacy scanner checks source filenames/content, staged blobs, all local Git blobs/commit/tag objects (including older staged objects), release filenames/members, and executable contents. Local identity checks include UTF-8 and UTF-16 representations. Private identifiers are evaluated locally and redacted from reports. Automated scanning is evidence, not a guarantee against every possible credential format; review the complete allowlist and history before publication.

The output is six portable desktop archives, one allowlisted source archive, `SHA256SUMS`, and `BUILD-INFO.json`. The Windows launch helper, macOS command helper, or Linux shell helper runs the bundled executable. No administrator installation or background service is required.

## Release gates before publication

- Obtain explicit owner approval for each public release. The initial repository and unsigned alpha were authorized under `dibyapp/pc-to-iphone-share`; local build tools do not publish them.
- Re-run source/archive/binary privacy scans, including machine-specific deny strings passed locally. Inspect the exact archive manifest; do not rely solely on `.gitignore`.
- Review history and commit metadata. The owner requested sole authorship using the public GitHub identity and noreply email, without co-author trailers. Private computer identity and private email must stay out.
- Enable GitHub private vulnerability reporting; community reports containing sensitive details must use an appropriate private channel.
- Complete native OS/CPU tests in [PLATFORMS.md](PLATFORMS.md), plus mobile certificate/download checks.
- Decide whether the first release remains clearly labeled unsigned pre-release or obtain Windows signing and macOS signing/notarization credentials. Never commit signing keys.
- Review version numbers, release notes, artifact names, checksums, and Go runtime notices.
- Verify all repository/release URLs. The static site remains undeployed: do not invent a canonical domain or claim it is hosted.
- Review search content for accurate feature/support claims. SEO does not guarantee rankings.

## Prepared workflows

The CI workflow tests source and performs an isolated native executable smoke check on Windows, macOS, and Linux runners. These workflows have not been run on a hosting service during local preparation. The build workflow is **manual only** and uploads build outputs to its run only when a maintainer deliberately dispatches it after publication approval. It has no release-creation, site-deployment, or auto-publishing job. Hosting a static content site is a separate, later action.
