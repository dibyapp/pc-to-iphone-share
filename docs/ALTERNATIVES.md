# Choosing a local file-transfer workflow

PC to iPhone Share is useful when a computer owner wants a phone browser to pull files from one password-protected, read-only folder. It is not a replacement for every file-transfer or synchronization tool.

| Your task | Option to investigate | Why |
| --- | --- | --- |
| Browse one computer folder from a phone browser | PC to iPhone Share | Portable host, read-only HTTPS downloads, generated password, manual local-certificate pairing |
| Send files between many nearby devices | [LocalSend](https://github.com/localsend/localsend) | Broader local device-to-device experience, including browser-sharing features in current versions |
| Browser-based peer transfers and pairing | [PairDrop](https://github.com/schlagmichdoch/PairDrop) | WebRTC/WebSocket browser workflow and self-hosting options |
| Transfer with short codes, including remote use | [croc](https://github.com/schollz/croc) | Native CLI, web client, and configurable transport/relay workflows |
| Run a rich file server with uploads and multiple protocols | [copyparty](https://github.com/9001/copyparty) | File-server features beyond this project's download-only scope |
| Quickly serve non-sensitive development files | [Python http.server](https://docs.python.org/3/library/http.server.html) | Built-in basic development server; production security is explicitly limited |

These tools overlap. No claim is made that PC to iPhone Share is faster, more secure, or the only phone-app-free option. Its implementation has not had an independent audit. Consider a mature alternative when you need device discovery, uploads, remote access, synchronization, or established platform coverage. Consult each project's current docs before choosing.
