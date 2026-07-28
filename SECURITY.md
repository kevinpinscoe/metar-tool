# Security Policy

## Supported versions

Only the latest released version receives fixes. There are no long-term support
branches — if you are behind, upgrade before reporting.

| Version | Supported |
|---|---|
| latest `v1.1.x` release | yes |
| anything older | no |

## Reporting a vulnerability

Report privately through GitHub's
[security advisories](https://github.com/kevinpinscoe/metar-tool/security/advisories/new)
form. Please do not open a public issue for a suspected vulnerability.

This is a personal project maintained by one person in his spare time. Expect
an acknowledgement within a week. Fixes are best-effort, and there is no
guaranteed response window.

## Verifying a release

Release artifacts are signed with [cosign](https://github.com/sigstore/cosign)
keyless signing, and each release publishes an SPDX SBOM alongside the binaries.
The SBOMs are listed in `checksums.txt`, so verifying the signed checksum file
covers them too:

```bash
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp \
    'https://github.com/kevinpinscoe/metar-tool/.github/workflows/release.yml@refs/tags/v.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
```

Then check any downloaded artifact against the verified `checksums.txt`:

```bash
sha256sum -c checksums.txt --ignore-missing
```
