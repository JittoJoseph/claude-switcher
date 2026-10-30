# Claude Switcher

Switch Claude Desktop between saved accounts, and manage their local Claude Code chats.
Use a primary account most of the time and a backup account when you need it.

[![Release](https://img.shields.io/github/v/release/JittoJoseph/claude-switcher)](https://github.com/JittoJoseph/claude-switcher/releases/latest)
[![CI](https://github.com/JittoJoseph/claude-switcher/actions/workflows/ci.yml/badge.svg)](https://github.com/JittoJoseph/claude-switcher/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

![Claude Switcher account screen](assets/readme/hero.png)

## Install and open

No Go, Git, Node.js, or administrator account is needed for a release installation.
Claude Desktop must already be installed. Before you save its profile, sign in to your primary account in Desktop.

### Windows: copy into PowerShell

```powershell
irm https://raw.githubusercontent.com/JittoJoseph/claude-switcher/main/install.ps1 | iex
claude-switcher
```

The installer supports Windows PowerShell 5.1 and PowerShell 7 on x64 and ARM64 Windows.
It installs for your Windows user, updates PATH in the current terminal, and creates a Start menu shortcut.
You can also open **Claude Switcher** from the Start menu.

### macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/JittoJoseph/claude-switcher/main/install.sh | sh
~/.local/bin/claude-switcher
```

The installer supports Intel and Apple silicon Macs, plus x64 and ARM64 Linux.
It installs to `~/.local/bin`. It prints a PATH command when your shell does not include that folder.
Linux Desktop process detection can vary between distributions and Desktop packages.

Both installers resolve one release version, verify its SHA-256 checksum, and check the executable before replacement.
They preserve the previous executable if download or validation fails.
You can inspect [install.ps1](install.ps1) or [install.sh](install.sh) before you run them.
Archives and checksums are also available on the [release page](https://github.com/JittoJoseph/claude-switcher/releases/latest).

## Save your primary and backup accounts

**Account switching does not copy Code chats automatically.** Each account keeps its own sidebar.
Use the **Chats** tab to copy the chats that you want in another account.
The switcher does not combine subscriptions or change account usage limits.

Before the first switch, back up the Desktop data folder and `~/.claude/projects`.
Saved profiles contain authentication data. Keep `~/.claude-instances` private and out of cloud sync.

1. In Claude Desktop, sign in to your primary account.
2. Open `claude-switcher` and save that login as `primary` in **Accounts**.
3. Select **Add account** and name the new profile `backup`.
4. When Desktop opens signed out, sign in to your backup account.
5. Return to the switcher and save the backup login.
6. Select `primary` to return to your main account.

On Windows, the switcher asks you to quit Desktop from its tray menu before a login operation.
Closing the Desktop window does not quit the app. The switcher waits until Desktop is closed.
It then changes the login files and reopens Desktop.

For the same setup from a terminal:

```powershell
# Sign in to the primary account in Desktop, then quit Desktop from its tray menu.
claude-switcher save primary
claude-switcher add backup

# Sign in to the backup account in Desktop, then quit Desktop from its tray menu.
claude-switcher save backup
claude-switcher switch primary
```

Later, use **Accounts** or these commands:

```powershell
claude-switcher switch backup
claude-switcher switch primary
```

Claude can expire or revoke a saved login. If that happens, sign in again. Then save the profile again.
Avoid Desktop's **Log out** action between switches because it can revoke the saved session.

## Manage Code chats

![Chats from multiple accounts with a preview](assets/readme/chats.png)

| Action | Result |
| --- | --- |
| Copy | Add the chat to another account's sidebar. Keep the source record. |
| Move | Transfer the sidebar record to another account. Remove the source record. |
| Merge | Copy the newest records from other sidebars into one sidebar. |
| Recover | Restore sidebar records for conversations that remain on disk. |
| Undo | Restore the files from the most recent chat-change journal. |
| Export | Save a conversation as Markdown or HTML. |

The newest sidebar record wins when two accounts have the same chat.
The switcher reads transcript history without rewriting it.
**Lost** lists transcripts that no known sidebar references. Review the list before recovery.
Terminal-only Claude Code conversations can also appear there.

This tool manages local **Code-tab** chats. It does not transfer cloud Chat conversations or manage CLI authentication.
Chat titles support Persian, Arabic, Hebrew, and other languages.
The **Usage** tab counts tokens from local transcripts. It does not show the subscription's remaining weekly allowance.

## Diagnose a setup

```powershell
claude-switcher doctor
claude-switcher list
```

On current Windows Desktop versions, the active data folder is usually:

```text
%LOCALAPPDATA%\Claude-Data\claude_pzs8sxrjxfjjc
```

Older MSIX and standalone installations are also supported.
The switcher prefers the migrated folder for a package and retains its MSIX launch identity.
A settings-only folder left by an older installer does not count as another signed-in installation.

| Symptom | Next step |
| --- | --- |
| No saved profiles | Save the account that is signed in to Desktop. |
| Desktop is still running | Quit it from the tray menu and let the switcher continue. |
| Chats disappear after a switch | Use **Chats** to copy records into the other account's sidebar. |
| Several organizations match a profile | Select the destination list in the app, or use an account/org reference in the CLI. |
| A saved login no longer works | Sign in again in Desktop and save that profile again. |
| An update cannot replace the executable | Close other switcher instances and retry. |

## Update or uninstall

To update, close other switcher instances and run:

```powershell
claude-switcher update
```

You can also repeat the installation command. Profiles and Claude chats remain in place.

To uninstall on Windows:

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/JittoJoseph/claude-switcher/main/install.ps1))) -Uninstall
```

The uninstaller removes the executable, its PATH entry, and its shortcut.
It preserves saved profiles, Claude data, and unrelated files in the installation folder.

To uninstall on macOS or Linux:

```sh
rm -f ~/.local/bin/claude-switcher
```

For a specific release on Windows:

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/JittoJoseph/claude-switcher/main/install.ps1))) -Version v3.0.1
```

For a specific release on macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/JittoJoseph/claude-switcher/main/install.sh | CLAUDE_SWITCHER_VERSION=v3.0.1 sh
```

## Commands

Run `claude-switcher --help` or append `--help` to a command for its arguments.
Run `claude-switcher` without arguments to open the terminal app.

| Command | Purpose |
| --- | --- |
| `list` | Show profiles, accounts, and chat lists. |
| `save <name>` | Save Desktop's current login. |
| `add <name>` | Save the current profile and prepare another account. |
| `switch <profile>` | Restore a saved login and reopen Desktop. |
| `rename <old> <new>` | Rename a profile. |
| `remove <name>` | Remove a saved profile. Keep its chats. |
| `chats [list] [--lost]` | List chats or recoverable conversations. |
| `copy --from A --to B <ids...>` | Copy sidebar records. |
| `move --from A --to B <ids...>` | Move sidebar records. |
| `merge --to <list>` | Merge sidebar records into a list. |
| `rescue [--to <list>] [--all]` | Rebuild missing sidebar records. |
| `export <id> [--format md]` | Export a conversation. |
| `undo` / `history` | Restore a chat change or inspect its journal. |
| `usage` / `doctor` | Inspect local usage or diagnose the setup. |
| `config` / `version` / `update` | Manage preferences or the installed version. |

Common flags include `--json`, `--yes`, `--no-launch`, `--theme`, and `--rtl`.
Exit codes: 0 means success, 1 means an error, and 2 means invalid arguments.
Codes 3, 4, and 5 mean a safety refusal, missing data, and cancellation.

## Data and failure handling

| Data | Windows | macOS | Linux |
| --- | --- | --- | --- |
| Desktop data | `%LOCALAPPDATA%\Claude-Data\claude_pzs8sxrjxfjjc` | `~/Library/Application Support/Claude` | `~/.config/Claude` |
| Code transcripts | `%USERPROFILE%\.claude\projects` | `~/.claude/projects` | `~/.claude/projects` |
| Saved profiles | `%USERPROFILE%\.claude-instances` | `~/.claude-instances` | `~/.claude-instances` |

Login saves finish a staged copy before replacement. Failed login restores and clears attempt to restore the original files.
If rollback fails, the error identifies the recovery folder.
An operating-system lock prevents overlapping mutations in the same profile vault.
Metadata writes use temporary files, and chat changes have journals and backups.

Login replacement spans several files. It cannot guarantee a complete transaction across a power loss.
For an unrecognized custom `--data-dir`, automatic process checks and app launching are disabled.
Normal detected installations and recognized folder overrides retain those checks.

Account and chat operations use local files. Update checks and explicit updates contact GitHub.
The switcher does not upload your profiles or conversations.

## Build and release

Source builds require Go 1.26.7 or later.

```sh
git clone https://github.com/JittoJoseph/claude-switcher.git
cd claude-switcher
go test ./...
go build -o bin/claude-switcher ./cmd/claude-switcher
```

On Windows, use `bin/claude-switcher.exe` for the output name.
`go run ./tools/demodata -out demo` creates sample accounts and conversations for a local demo.
See [the architecture guide](docs/ARCHITECTURE.md) for the package layout.

Successful CI on `main` automatically calls the release workflow.
The workflow creates the next patch tag, starting at `v3.0.1` for this fork.
It builds release archives and Linux packages, publishes checksums, and adds build attestations.
It then installs the published release on Windows, macOS, and Linux, including the README's installation commands.
Pull requests and failed CI runs do not publish releases. An older commit does not release after `main` advances.
A rerun reuses an existing tag and does not replace a published release.
Maintainers can retry the release workflow from the repository's **Actions** tab.

## Credits

This repository is a fork of [Nyxon's Claude Switcher](https://github.com/nyxon-tech/claude-switcher).
The original project started from [claude-profile-switcher](https://github.com/NeezerGu/claude-profile-switcher) by [@NeezerGu](https://github.com/NeezerGu).
The original MIT license and author credits remain in place.

Built with [Bubble Tea, Lip Gloss, and Fang](https://github.com/charmbracelet) by Charm.
Unofficial. Not affiliated with or endorsed by Anthropic.
