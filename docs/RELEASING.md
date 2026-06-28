# Releasing harnest

Distribution channels: **GitHub Releases**, **Homebrew tap**, **npm** (macOS/Linux), **winget** (Windows).

## 1. Build artifacts

```bash
make release
```

Produces in `dist/` for darwin/linux/windows × amd64/arm64, plus `checksums.txt`.
Windows targets are `harnest-windows-amd64.exe` and `harnest-windows-arm64.exe`.

## 2. GitHub Release

Tag `vX.Y.Z`, upload every file in `dist/` (binaries + `.exe` + `checksums.txt`) as release assets.
The npm `postinstall` and winget manifests download these raw assets by URL, so the file names must match exactly.

## 3. Homebrew tap

Update `alexgladkov/homebrew-tap/harnest.rb` with the new version + sha256 of the darwin/linux tarballs.

## 4. npm

```bash
cd npm && npm version X.Y.Z --no-git-tag-version && npm publish
```

npm is macOS/Linux only (`os` field in `package.json`). Windows users use winget.

## 5. winget (Windows)

Manifests live in `winget/manifests/a/AlexGladkov/Harnest/<version>/` and are generated from the
built Windows binaries so the sha256 never drifts:

```bash
make release          # must run first — produces dist/harnest-windows-*.exe
make winget           # regenerates manifests with sha256 from those exes
```

Manifest type is **portable** (raw `.exe`, no installer/zip). `PortableCommandAlias: harnest`
makes `harnest` available on PATH after `winget install`.

### Submit to the public catalog

The package becomes installable only after the manifests are merged into
[`microsoft/winget-pkgs`](https://github.com/microsoft/winget-pkgs). Prerequisite: the GitHub
Release for this version must already expose the two `.exe` assets at their public URLs
(the manifests reference them, and winget CI downloads + hash-checks them).

Easiest path — [`wingetcreate`](https://github.com/microsoft/winget-create) (Windows / `dotnet tool`):

```powershell
wingetcreate update AlexGladkov.Harnest `
  --version X.Y.Z `
  --urls "https://github.com/AlexGladkov/harnest/releases/download/vX.Y.Z/harnest-windows-amd64.exe" `
         "https://github.com/AlexGladkov/harnest/releases/download/vX.Y.Z/harnest-windows-arm64.exe" `
  --submit
```

This recomputes hashes, regenerates the manifest, and opens the PR. Alternatively, copy the files
from `winget/manifests/...` into a fork of `microsoft/winget-pkgs` and open the PR manually, then
`winget validate` / sandbox-test before submitting.

## Recommended: automate with GoReleaser

All four channels are currently updated by hand → version drift risk grows with each channel.
[GoReleaser](https://goreleaser.com/) builds every target, creates the GitHub Release, and has
native `brews:` and `winget:` blocks that publish to the tap and open the winget-pkgs PR in one
`goreleaser release` run. Migrating the manual steps above into `.goreleaser.yaml` + a
`release.yml` GitHub Action is the long-term fix.
