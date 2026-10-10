# Search and discovery research

Research date: 2026-10-11. Audience: English-speaking developers, Linux desktop users, and people who want to download a computer folder on an iPhone or Android device. Geography: global English; no country-specific volume claim is made.

## What was researched, and what the numbers mean

Live searches covered Windows-to-iPhone transfers, Ubuntu/Linux-to-iPhone transfers, password-protected browser file sharing, local/cloud-free sharing, read-only folder access, and Safari connection problems. Primary project documentation was reviewed for LocalSend, PairDrop, croc, copyparty, Python's built-in file server, Google Search, GitHub discovery, and Apple certificate pairing.

Search results show competing pages and language people use. They do **not** provide reliable monthly volume, keyword difficulty, or a forecast of rankings. No paid Keyword Planner/Ahrefs/Semrush dataset was exported for this research. Every priority below is an editorial judgment about intent and product fit. `KEYWORDS.csv` is a reproducible keyword backlog, not a fabricated volume report. Search result order can differ by location, time, and personalization.

## Positioning that matches the product

**Download one computer folder on your phone over local Wi-Fi, with a generated password and HTTPS. No phone app or cloud upload.**

The differentiating scope is small and inspectable: one selected interface, one read-only folder, portable standard-library Go binaries, constrained local certificates, explicit pairing, and no relay. Do not claim a unique invention, universal OS support, automatic discovery, phone-to-computer uploads, or a frictionless AirDrop replacement. The certificate pairing step is a conversion obstacle and must be visible before installation.

## Keyword priorities

| Cluster | Primary phrase | Priority | Why it fits | Destination |
| --- | --- | --- | --- | --- |
| Browser folder download | download PC files on iPhone without an app | P1 | Exact receiver experience and clear installation tradeoff | README + Windows guide |
| Windows/iPhone | transfer files from Windows to iPhone over WiFi | P1 | Recognizable problem; narrower than generic file sharing | Windows guide |
| Linux/iPhone | transfer files from Ubuntu to iPhone | P1 | Linux/iOS interoperability questions are visible in live results | Ubuntu guide + Linux setup |
| Local privacy | local file sharing without cloud | P1 | Matches direct LAN transport and no accounts | Home + private-sharing guide |
| Folder scope | password protected folder sharing over WiFi | P1 | Relevant to an owner-controlled temporary folder | README + security doc |
| Browser server | read only HTTPS file server | P1 | Developer intent aligned with the architecture | Architecture + README |
| Mac/Android | transfer files from Mac to Android over WiFi | P2 | Supported cross-ecosystem use; hardware validation still explicit | Mac/Android guide |
| Windows/Android | download Windows files on Android browser | P2 | Browser download intent rather than synchronization | Workflow guide |
| Setup help | iPhone local HTTPS certificate trust | P2 | Critical product adoption obstacle | Certificate guide |
| Connection help | Safari local file server cross site requests blocked | P2 | Specific symptom; fix by using the correct same-origin URL, not disabling privacy | Troubleshooting |
| Large downloads | large file transfer PC to iPhone without cloud | P2 | Streaming and ranges fit; client/storage limits remain | Windows guide + FAQ |
| Broad discovery | AirDrop alternative Windows Linux | P3 | Established competitors and broader expectations | Balanced alternatives comparison |
| Brand navigation | PC to iPhone Share GitHub | P1 after publication | Helps users find source and verified releases | Repository |

P1 means product-fit focus, **not** proven high volume or low competition. Use one main destination per intent. Windows/Ubuntu/Linux guides share a setup foundation but must carry distinct OS instructions and troubleshooting; do not create dozens of near-identical pages to catch slight spelling variants.

## Competitive context

| Project | Established use case | How to explain PC to iPhone Share fairly |
| --- | --- | --- |
| [LocalSend](https://github.com/localsend/localsend), [official site](https://localsend.org/) | Cross-platform local transfers with a broader device-to-device experience | A focused folder browser is an alternative workflow. Current LocalSend also has browser-sharing features; never claim phone-app-free sharing is unique. |
| [PairDrop](https://github.com/schlagmichdoch/PairDrop) | Browser transfers using WebRTC/WebSockets, pairing and self-hosting options | PC to iPhone Share uses a desktop HTTPS folder server and does not need a discovery/signaling service. It requires manual local-certificate trust. |
| [croc](https://github.com/schollz/croc) | Code-based secure transfers, native CLI and a web client, local/public relay options | PC to iPhone Share is for a browsable local folder; remote code-based transfers are outside its scope. |
| [copyparty](https://github.com/9001/copyparty) | A feature-rich portable file server with uploads, many protocols, and media features | PC to iPhone Share intentionally offers fewer operations: download-only access to one folder. |
| [Python http.server](https://docs.python.org/3/library/http.server.html) | A convenient basic development file server; newer Python supports TLS configuration | PC to iPhone Share adds generated password authentication, confined access, sessions, pairing, and subnet binding. Do not imply modern Python cannot serve HTTPS. |

These are descriptions of workflows, not performance benchmarks or an exhaustive security comparison. Verify competing features again before using a comparison in a future launch.

## Search implementation

- Repository name: `pc-to-iphone-share`. Description: “Password-protected Wi-Fi file downloads from one computer folder to iPhone/Android browsers. Windows, macOS, Linux. No cloud or phone app.”
- Topics: `file-sharing`, `wifi-file-transfer`, `local-file-sharing`, `iphone`, `android`, `windows`, `macos`, `linux`, `ubuntu`, `https`, `self-hosted`, `golang`, `privacy`, `open-source`.
- README leads with the problem, real screenshots, release downloads, pairing requirement, and tested-platform status. Screenshots use only synthetic data and are explicitly labeled.
- HTML pages use unique descriptive titles, readable headings, concise descriptions, crawlable links, useful captions/alt text, and OS-specific instructions. No hidden keyword blocks or meta-keywords tag.
- The repository link is the approved live destination. The static site remains undeployed; no fake canonical domain, sitemap domain, hosted installation page, or Search Console ownership claim is created.
- Software/FAQ JSON-LD describes visible facts only. It contains no invented review count, rating, sales price, or star count. Schema markup does not guarantee a rich result; FAQ visibility rules change.

Google recommends useful original content, descriptive titles and links, and images with context. A public repository can be discovered through GitHub topics and external links, but GitHub stars and search rankings are separate outcomes. [Helpful content](https://developers.google.com/search/docs/fundamentals/creating-helpful-content), [titles](https://developers.google.com/search/docs/appearance/title-link), [snippets](https://developers.google.com/search/docs/appearance/snippet), [links](https://developers.google.com/search/docs/crawling-indexing/links-crawlable), [images](https://developers.google.com/search/docs/appearance/google-images), [GitHub topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics).

## What to measure after publication

Record weekly repository unique visitors, clones, release asset downloads, new issues/PRs, installation failures, and genuine stars. These metrics have different meanings and retention windows; archive only aggregate counts. GitHub traffic is available to people with appropriate repository permissions. [GitHub traffic documentation](https://docs.github.com/en/repositories/viewing-activity-and-data-for-your-repository/viewing-traffic-to-a-repository).

If a separate site is later approved and deployed, verify it in Search Console, collect impressions/clicks/queries, and revise titles/content from actual data. Keyword Planner can provide estimates with an appropriate Google Ads account; volume definitions and ranges need interpretation. [Keyword Planner](https://support.google.com/google-ads/answer/7337243).

Stop expanding keywords when visitors cannot install or pair the app. A clear first download and reliable documentation are more valuable than a larger keyword list. The first month's 100-star target is an experiment, not a promise; see [GROWTH-PLAN.md](GROWTH-PLAN.md).
