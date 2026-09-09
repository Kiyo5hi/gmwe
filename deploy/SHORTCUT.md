# Apple Shortcut Distribution

`client/public/downloads/GMWE-iPhone.shortcut` is the credential-free, signed
artifact imported successfully on the operator's iPhone on 2026-09-08.
SHA256: `d1f203e842ffea2216dc0da08568b84c90cbc875babd868b1751c7b9393aefb9`.
It was signed through the operator-approved HubSign service using Cherri 2.3.0's
request format. Apple-device import was confirmed; the encrypted signed archive
was not independently decrypted to compare its embedded actions.

The reproducible source, builder, structural tests and signing procedure live in
the private homelab-infra repository (`templates/shortcuts/gmwe.cherri`,
`scripts/build-gmwe-shortcut.py`, `docs/gmwe-shortcut.md`; source at 46383c6).
The website serves the signed artifact byte-for-byte; no runtime credentials,
connection JSON or token values may ever be added to public assets.

The account page provides separate Shortcut and Connection downloads. The first
installs the workflow; the second performs authenticated code/PKCE consent and
downloads a secret per-device connection. The Shortcut first reads its saved
`gmwe-iphone.json`, selecting an initial connection only if that file is absent.
After renewal it saves and reads back the rotating token before a content write.
Files storage may sync with private iCloud Drive; it is not Keychain storage.

To update the workflow, compile and validate the credential-free source, sign it
using the approved process, verify the resulting checksum and iPhone import,
then replace this asset and its checksum together. Preserve the credential file
when reinstalling the workflow. Never publish an authorization download.
