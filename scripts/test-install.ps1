# Installer regression tests. No changes to the user's PATH, shortcuts, or Claude data.
param([Parameter(Mandatory = $true)][string]$Binary)
$ErrorActionPreference = 'Stop'
$Binary = (Resolve-Path -LiteralPath $Binary).Path
$installer = Join-Path $PSScriptRoot '..\install.ps1'
$root = Join-Path ([IO.Path]::GetTempPath()) ('claude-switcher-install-test-' + [guid]::NewGuid())
$dir = Join-Path $root 'installation with spaces'
$beforePath = [Environment]::GetEnvironmentVariable('Path', 'User')
$beforeSessionPath = $env:Path
try {
    & $installer -Binary $Binary -InstallDir $dir -NoPath -NoShortcut
    $exe = Join-Path $dir 'claude-switcher.exe'
    if ((Get-FileHash -LiteralPath $Binary).Hash -ne (Get-FileHash -LiteralPath $exe).Hash) { throw 'Installed file differs' }
    & $installer -Binary $Binary -InstallDir $dir -NoPath -NoShortcut
    $before = (Get-FileHash -LiteralPath $exe).Hash
    $locked = [IO.File]::Open($exe, [IO.FileMode]::Open, [IO.FileAccess]::Read, [IO.FileShare]::Read)
    $failed = $false
    try { & $installer -Binary $Binary -InstallDir $dir -NoPath -NoShortcut }
    catch { $failed = $true }
    finally { $locked.Dispose() }
    if (-not $failed) { throw 'Installing over a locked executable must fail' }
    if ((Get-FileHash -LiteralPath $exe).Hash -ne $before) { throw 'Failed update changed the installation' }
    $bad = Join-Path $root 'invalid.exe'
    [IO.File]::WriteAllText($bad, 'not an executable')
    $failed = $false
    try { & $installer -Binary $bad -InstallDir $dir -NoPath -NoShortcut }
    catch { $failed = $true }
    if (-not $failed -or (Get-FileHash -LiteralPath $exe).Hash -ne $before) { throw 'Invalid download replaced a working installation' }
    Set-Content -LiteralPath (Join-Path $dir 'keep.txt') -Value 'unrelated file'
    & $installer -Uninstall -InstallDir $dir -NoPath -NoShortcut
    if (Test-Path -LiteralPath $exe) { throw 'Uninstall left the executable' }
    if (-not (Test-Path -LiteralPath (Join-Path $dir 'keep.txt'))) { throw 'Uninstall removed an unrelated file' }
    & $installer -Uninstall -InstallDir $dir -NoPath -NoShortcut
    if ([Environment]::GetEnvironmentVariable('Path', 'User') -ne $beforePath -or $env:Path -ne $beforeSessionPath) { throw 'Isolated test changed PATH' }
    Write-Host 'Installer regression tests passed.'
}
finally {
    $resolved = [IO.Path]::GetFullPath($root)
    $temp = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\') + '\'
    if (-not $resolved.StartsWith($temp, [StringComparison]::OrdinalIgnoreCase)) { throw 'Unexpected cleanup path' }
    if (Test-Path -LiteralPath $resolved) { Remove-Item -LiteralPath $resolved -Recurse -Force }
}
