#!/bin/sh
# Claude Switcher - current-user installer for macOS and Linux.
set -eu
repo=JittoJoseph/claude-switcher
version=${CLAUDE_SWITCHER_VERSION:-latest}
bin=${CLAUDE_SWITCHER_BIN:-$HOME/.local/bin}
case "$bin" in /*) ;; *) echo 'CLAUDE_SWITCHER_BIN must be an absolute path.' >&2; exit 1 ;; esac
case "$(uname -s)" in
  Darwin) os=darwin arch=all ;;
  Linux)
    os=linux
    case "$(uname -m)" in
      x86_64 | amd64) arch=amd64 ;;
      aarch64 | arm64) arch=arm64 ;;
      *) echo "Unsupported CPU: $(uname -m)" >&2; exit 1 ;;
    esac ;;
  *) echo 'Use install.ps1 on Windows.' >&2; exit 1 ;;
esac
if [ "$version" = latest ]; then
  # Resolve latest once so a concurrent release cannot change either download.
  latest=$(curl --fail --silent --show-error --location --head --max-time 30 \
    --output /dev/null --write-out '%{url_effective}' "https://github.com/$repo/releases/latest")
  version=${latest##*/}
fi
case "$version" in v*) ;; *) version=v$version ;; esac
if ! printf '%s\n' "$version" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$'; then
  echo 'Version must be latest or a release version such as v3.0.1.' >&2
  exit 1
fi
base="https://github.com/$repo/releases/download/$version"
file="claude-switcher_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
next=''
trap 'rm -rf "$tmp"; if [ -n "$next" ]; then rm -f "$next"; fi' EXIT
trap 'exit 1' HUP INT TERM
curl --fail --silent --show-error --location --retry 3 --max-time 120 "$base/$file" -o "$tmp/$file"
curl --fail --silent --show-error --location --retry 3 --max-time 30 "$base/checksums.txt" -o "$tmp/checksums.txt"
# Compare names as strings, not regular expressions. Reject absent or duplicate entries.
want=$(awk -v name="$file" 'NF == 2 { f=$2; sub(/^\*/, "", f); if (f == name) { sum=$1; count++ } } END { if (count == 1 && length(sum) == 64 && sum !~ /[^0-9a-fA-F]/) print tolower(sum); else exit 1 }' "$tmp/checksums.txt") || {
  echo "$file needs exactly one valid SHA-256 entry in checksums.txt." >&2
  exit 1
}
if command -v sha256sum >/dev/null 2>&1; then
  got=$(sha256sum "$tmp/$file" | awk '{print $1}')
else
  got=$(shasum -a 256 "$tmp/$file" | awk '{print $1}')
fi
[ "$got" = "$want" ] || { echo 'Checksum mismatch. The existing installation remains unchanged.' >&2; exit 1; }
# Extract just the executable, then stage it on the destination filesystem.
tar -xzf "$tmp/$file" -C "$tmp" claude-switcher
[ -f "$tmp/claude-switcher" ] && [ ! -L "$tmp/claude-switcher" ] || { echo 'The archive has no regular executable.' >&2; exit 1; }
mkdir -p "$bin"
next=$(mktemp "$bin/.claude-switcher.XXXXXX")
cp "$tmp/claude-switcher" "$next"
chmod 755 "$next"
installed=$("$next" version --short)
[ "$installed" = "${version#v}" ] || { echo 'The executable version does not match the release.' >&2; exit 1; }
mv -f "$next" "$bin/claude-switcher"
next=''
printf '\nClaude Switcher %s installed to %s\n' "$installed" "$bin"
case ":$PATH:" in
  *":$bin:"*) echo 'Run: claude-switcher' ;;
  *) printf 'Run now: "%s/claude-switcher"\n' "$bin"
     printf 'For future terminals, add to your shell profile: export PATH="%s:$PATH"\n' "$bin" ;;
esac
