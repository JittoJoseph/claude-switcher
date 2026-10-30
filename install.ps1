# Claude Switcher - current-user installer. Compatible with Windows PowerShell 5.1.
# Install: irm https://raw.githubusercontent.com/JittoJoseph/claude-switcher/main/install.ps1 | iex
param(
    [switch]$Uninstall,
    [switch]$NoShortcut,
    [switch]$NoPath,
    [string]$Version = 'latest',
    [string]$Binary = '',
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\claude-switcher"
)
$previousError = $ErrorActionPreference
$previousProgress = $ProgressPreference
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$repo = 'JittoJoseph/claude-switcher'

function Set-UserPath([bool]$Add) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $parts = @($current -split ';' | Where-Object { $_ -and $_.Trim().TrimEnd('\') -ine $dir })
    if ($Add) { $parts += $dir }
    $updated = $parts -join ';'
    if ($updated -ne $current) { [Environment]::SetEnvironmentVariable('Path', $updated, 'User') }
    $sessionParts = @($env:Path -split ';' | Where-Object { $_ -and $_.Trim().TrimEnd('\') -ine $dir })
    if ($Add) { $sessionParts += $dir }
    $env:Path = $sessionParts -join ';'
}

function Remove-StagingFolder([string]$Path) {
    $resolved = [IO.Path]::GetFullPath($Path)
    if (-not $resolved.StartsWith($dir + '\', [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Refusing to remove a staging folder outside the installation directory.'
    }
    if (Test-Path -LiteralPath $resolved) { Remove-Item -LiteralPath $resolved -Recurse -Force }
}

try {
    if (-not [IO.Path]::IsPathRooted($InstallDir)) { throw 'InstallDir must be an absolute path.' }
    $dir = [IO.Path]::GetFullPath($InstallDir).TrimEnd('\')
    $exe = Join-Path $dir 'claude-switcher.exe'
    $shortcut = Join-Path ([Environment]::GetFolderPath('Programs')) 'Claude Switcher.lnk'
    if ($Uninstall) {
        Get-Process -Name claude-switcher -ErrorAction SilentlyContinue |
            Where-Object { $_.Path -ieq $exe } | Stop-Process
        # Remove only files owned by this installer. Preserve unrelated files and saved profiles.
        foreach ($name in @('claude-switcher.exe', 'claude-switcher.exe.old', 'claude-switcher.exe.new', 'claude-switcher.cmd', 'claude-switcher.ps1')) {
            $file = Join-Path $dir $name
            if (Test-Path -LiteralPath $file -PathType Leaf) { Remove-Item -LiteralPath $file -Force }
        }
        if (-not $NoShortcut -and (Test-Path -LiteralPath $shortcut)) {
            $shell = New-Object -ComObject WScript.Shell
            if ($shell.CreateShortcut($shortcut).TargetPath -ieq $exe) { Remove-Item -LiteralPath $shortcut -Force }
        }
        if ((Test-Path -LiteralPath $dir) -and @(Get-ChildItem -LiteralPath $dir -Force).Count -eq 0) {
            Remove-Item -LiteralPath $dir
        }
        if (-not $NoPath) { Set-UserPath $false }
        Write-Host 'Claude Switcher removed. Saved profiles and Claude chats remain in place.' -ForegroundColor Green
        return
    }
    if (Get-Process -Name claude-switcher -ErrorAction SilentlyContinue | Where-Object { $_.Path -ieq $exe }) {
        throw 'Close Claude Switcher before you install an update.'
    }
    if ($Binary -and -not (Test-Path -LiteralPath $Binary -PathType Leaf)) { throw "Binary not found: $Binary" }
    if (-not $Binary) {
        [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
        if ($Version -eq 'latest') {
            $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -TimeoutSec 30
            $Version = $release.tag_name
        }
        if ($Version -notmatch '^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$') {
            throw 'Version must be latest or a release version such as v3.0.1.'
        }
        if (-not $Version.StartsWith('v')) { $Version = 'v' + $Version }
    }
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    $stage = Join-Path $dir ('.install-' + [guid]::NewGuid())
    New-Item -ItemType Directory -Path $stage | Out-Null
    $keepStage = $false
    try {
        $next = Join-Path $stage 'claude-switcher.exe'
        if ($Binary) { Copy-Item -LiteralPath $Binary -Destination $next }
        else {
            $nativeArch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
            $arch = switch ($nativeArch) { 'ARM64' { 'arm64' } 'AMD64' { 'amd64' } default { throw "Unsupported Windows architecture: $nativeArch" } }
            $asset = "claude-switcher_windows_$arch.zip"
            $base = "https://github.com/$repo/releases/download/$Version"
            $zip = Join-Path $stage $asset
            $checksums = Join-Path $stage 'checksums.txt'
            Invoke-WebRequest -UseBasicParsing -Uri "$base/$asset" -OutFile $zip -TimeoutSec 120
            Invoke-WebRequest -UseBasicParsing -Uri "$base/checksums.txt" -OutFile $checksums -TimeoutSec 30
            $lines = @(Get-Content -LiteralPath $checksums | Where-Object { $_ -match ('^[0-9a-fA-F]{64}\s+\*?' + [regex]::Escape($asset) + '$') })
            if ($lines.Count -ne 1) { throw "$asset needs exactly one valid SHA-256 entry in checksums.txt." }
            $want = ($lines[0] -split '\s+')[0]
            if ((Get-FileHash -LiteralPath $zip -Algorithm SHA256).Hash -ine $want) { throw 'Checksum mismatch. The existing installation remains unchanged.' }
            $expanded = Join-Path $stage 'archive'
            Expand-Archive -LiteralPath $zip -DestinationPath $expanded
            Copy-Item -LiteralPath (Join-Path $expanded 'claude-switcher.exe') -Destination $next
        }
        Unblock-File -LiteralPath $next
        $installed = & $next version --short
        if ($LASTEXITCODE -ne 0 -or -not $installed) { throw 'The downloaded executable did not pass its version check.' }
        $installed = ([string]$installed).Trim()
        if (-not $Binary -and $installed -ne $Version.Substring(1)) { throw 'The executable version does not match the requested release.' }
        $old = Join-Path $stage 'previous.exe'
        $hadOld = Test-Path -LiteralPath $exe
        if ($hadOld) { Move-Item -LiteralPath $exe -Destination $old }
        try { Move-Item -LiteralPath $next -Destination $exe }
        catch {
            if ($hadOld) {
                try { Move-Item -LiteralPath $old -Destination $exe }
                catch {
                    $keepStage = $true
                    throw "Installation rollback failed. The previous executable remains at $old. $($_.Exception.Message)"
                }
            }
            throw
        }
    }
    finally { if (-not $keepStage) { Remove-StagingFolder $stage } }
    foreach ($name in @('claude-switcher.cmd', 'claude-switcher.ps1')) {
        $legacy = Join-Path $dir $name
        if (Test-Path -LiteralPath $legacy -PathType Leaf) { Remove-Item -LiteralPath $legacy -Force }
    }
    if (-not $NoPath) { Set-UserPath $true }
    if (-not $NoShortcut) {
        $shell = New-Object -ComObject WScript.Shell
        $link = $shell.CreateShortcut($shortcut)
        $link.TargetPath = $exe
        $link.WorkingDirectory = $env:USERPROFILE
        $link.Description = 'Switch Claude Desktop accounts and manage Claude Code chats'
        $link.Save()
    }
    Write-Host "Claude Switcher $installed installed." -ForegroundColor Green
    if ($NoPath) { Write-Host "Run: & '$exe'" }
    else { Write-Host 'Run: claude-switcher' }
    if (-not $NoShortcut) { Write-Host 'You can also open Claude Switcher from the Start menu.' }
}
finally {
    $ErrorActionPreference = $previousError
    $ProgressPreference = $previousProgress
}
