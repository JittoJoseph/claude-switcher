# Security review

Review date: October 31, 2026.
Baseline: `bb03545`, plus the security and performance changes committed with this report.

## Conclusion and scope

The review found no evidence of intentional credential exfiltration, persistence, arbitrary remote-command execution, or Defender tampering in the reviewed application code.
This is a source review and a set of checks for this snapshot. It is not proof that all dependencies or future builds are harmless.

The review covered application networking, process execution, authentication-file handling, installers, exports, embedded Windows resources, dependency integrity, and release workflows.

## Defender alert investigation

A local `Trojan:Win32/ClickFix.CCM!MTB` alert identified a PowerShell command submitted during documentation editing.
The command contained a Python write of README text, including installation examples.
The affected item was that command line, rather than a switcher executable or a repository file.

Defender reported `ActionSuccess=True`, `DidThreatExecute=False`, and `IsActive=False` for that detection during investigation.
This evidence is consistent with a false positive on the documentation-writing command.
The event does not disclose the exact signature logic and does not establish that the entire computer is clean.
No protection setting, exclusion, or threat-allowance rule was changed.
A custom Defender scan of the repository completed.

Microsoft describes the real [ClickFix technique](https://www.microsoft.com/en-us/security/blog/2025/08/21/think-before-you-clickfix-analyzing-the-clickfix-social-engineering-technique/) as social engineering that gets users to execute commands.
An alert must be investigated from its affected item and behavior, rather than dismissed from its name alone.

## Reviewed capabilities

| Area | Observed behavior |
| --- | --- |
| Network | Application update code contacts GitHub release endpoints. No profile or transcript upload path was found. |
| Saved logins | Account switching copies local authentication files into the profile vault. These files remain sensitive credentials. |
| Process execution | The application opens or closes Desktop and its helpers. Explicit VM repair invokes a local disk tool after confirmation. |
| Installation | The installer downloads a release, verifies its checksum and version, updates user PATH, and creates a Start menu shortcut. |
| Persistence | No service, scheduled task, startup-folder entry, or automatic-start registry entry was found. A Start menu shortcut does not start the app at login. |
| Clipboard | The linked text-input dependency reads the clipboard for its paste action. No background clipboard collector was found. |
| Exports | HTML uses Go's escaping template package. The export template contains no script execution path. |
| Windows resources | Both checked-in `.syso` objects contain a non-executable `.rsrc` section. No executable section was present. |
| Dependencies | `go mod verify` passed. `govulncheck v1.9.0` reported no known vulnerabilities for the source build with Go 1.26.10. The older published binary had separate standard-library findings, described below. |

Checksums detect damaged or mismatched downloads. They do not provide independent trust if the release publisher is compromised.
Build attestations identify provenance. They are not a malware verdict.

## Hardening change

The review found that terminal rendering trusted plain-text input to contain no terminal control sequences.
A crafted title, path, or preview could violate that assumption.

Plain-text rendering now removes ANSI control sequences and unsafe control characters before layout and styling.
Regression tests cover clipboard-control sequences, screen controls, hyperlinks, and all supported rendering modes.
The change filters display output. It does not rewrite stored conversations or authentication files.

## Performance and release changes

Short CLI commands read cached update notices without making a network request.
They no longer query the terminal synchronously for its background color.
Multi-list chat reads parse each card once and share one transcript index per request. Doctor reuses its index for orphan detection.
No persistent account-state cache was introduced.

Routine CI tests three operating systems, including race detection on Linux and installer tests on Windows.
Installer tests reuse the Windows build. Linux does not repeat the suite without race detection.
Obsolete runs on the same branch are cancelled.

Release packaging and publication are manual. The release workflow requires successful CI for the exact commit.
Static inspection of the published `v3.0.1` archive verified its SHA-256 checksum and GitHub build attestation without executing the binary.
The binary scan found symbols covered by 14 standard-library advisories because that release used Go 1.26.7.
Binary symbol matches do not prove that each advisory is exploitable in this application.
They do require a patched build, independently of the Defender false-positive investigation.

The build baseline now requires Go 1.26.10, which includes the listed fixes.
Examples include [certificate verification](https://pkg.go.dev/vuln/GO-2026-6809) and [Windows root-directory handling](https://pkg.go.dev/vuln/GO-2026-6604).
The manual release workflow can distribute this security patch after CI passes.
Future pushes do not publish releases automatically.

## Remaining trust boundaries

- Saved profile files must remain private. This tool intentionally handles authentication data.
- Local journals contain file paths used by undo. The profile vault is trusted local state.
- Login replacement spans files and cannot guarantee a complete transaction through a power loss.
- An unrecognized custom data directory disables automatic Desktop detection and launching.
- This review does not replace a full incident investigation or a complete audit of every dependency.
