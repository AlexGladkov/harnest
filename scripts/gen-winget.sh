#!/usr/bin/env bash
# Generate winget manifests for harnest from built Windows binaries in dist/.
# Usage: VERSION=0.11.0 ./scripts/gen-winget.sh
# Requires: dist/harnest-windows-amd64.exe and dist/harnest-windows-arm64.exe (run `make release` first).
set -euo pipefail

VERSION="${VERSION:?set VERSION, e.g. VERSION=0.11.0}"
REPO="AlexGladkov/harnest"
PKG_ID="AlexGladkov.Harnest"
MANIFEST_VERSION="1.6.0"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist"
OUT="$ROOT/winget/manifests/a/AlexGladkov/Harnest/$VERSION"

EXE_AMD64="$DIST/harnest-windows-amd64.exe"
EXE_ARM64="$DIST/harnest-windows-arm64.exe"
for f in "$EXE_AMD64" "$EXE_ARM64"; do
  [ -f "$f" ] || { echo "missing $f — run 'make release' first" >&2; exit 1; }
done

# winget expects uppercase hex sha256.
sha() { shasum -a 256 "$1" | awk '{print toupper($1)}'; }
SHA_AMD64="$(sha "$EXE_AMD64")"
SHA_ARM64="$(sha "$EXE_ARM64")"
BASE_URL="https://github.com/$REPO/releases/download/v$VERSION"

mkdir -p "$OUT"

cat > "$OUT/$PKG_ID.yaml" <<EOF
# yaml-language-server: \$schema=https://aka.ms/winget-manifest.version.$MANIFEST_VERSION.schema.json
PackageIdentifier: $PKG_ID
PackageVersion: $VERSION
DefaultLocale: en-US
ManifestType: version
ManifestVersion: $MANIFEST_VERSION
EOF

cat > "$OUT/$PKG_ID.installer.yaml" <<EOF
# yaml-language-server: \$schema=https://aka.ms/winget-manifest.installer.$MANIFEST_VERSION.schema.json
PackageIdentifier: $PKG_ID
PackageVersion: $VERSION
InstallerType: portable
Commands:
  - harnest
ReleaseDate: ""
Installers:
  - Architecture: x64
    InstallerUrl: $BASE_URL/harnest-windows-amd64.exe
    InstallerSha256: $SHA_AMD64
    PortableCommandAlias: harnest
  - Architecture: arm64
    InstallerUrl: $BASE_URL/harnest-windows-arm64.exe
    InstallerSha256: $SHA_ARM64
    PortableCommandAlias: harnest
ManifestType: installer
ManifestVersion: $MANIFEST_VERSION
EOF

cat > "$OUT/$PKG_ID.locale.en-US.yaml" <<EOF
# yaml-language-server: \$schema=https://aka.ms/winget-manifest.defaultLocale.$MANIFEST_VERSION.schema.json
PackageIdentifier: $PKG_ID
PackageVersion: $VERSION
PackageLocale: en-US
Publisher: Alex Gladkov
PublisherUrl: https://github.com/AlexGladkov
PublisherSupportUrl: https://github.com/$REPO/issues
PackageName: Harnest
PackageUrl: https://github.com/$REPO
License: CC-BY-NC-4.0
LicenseUrl: https://github.com/$REPO/blob/main/LICENSE
ShortDescription: AI coding assistant configurator — generate configs for Claude Code, Cursor, Windsurf, Codex, OpenCode, and Qwen Code.
Tags:
  - ai
  - cli
  - coding-assistant
  - claude-code
  - cursor
  - windsurf
  - codex
  - opencode
  - qwen-code
  - developer-tools
ManifestType: defaultLocale
ManifestVersion: $MANIFEST_VERSION
EOF

# strip empty ReleaseDate (winget rejects empty string); set only if provided
if [ -z "${RELEASE_DATE:-}" ]; then
  grep -v 'ReleaseDate: ""' "$OUT/$PKG_ID.installer.yaml" > "$OUT/$PKG_ID.installer.yaml.tmp"
  mv "$OUT/$PKG_ID.installer.yaml.tmp" "$OUT/$PKG_ID.installer.yaml"
else
  sed -i.bak "s|ReleaseDate: \"\"|ReleaseDate: $RELEASE_DATE|" "$OUT/$PKG_ID.installer.yaml"
  rm -f "$OUT/$PKG_ID.installer.yaml.bak"
fi

echo "Generated manifests in $OUT"
ls -1 "$OUT"
