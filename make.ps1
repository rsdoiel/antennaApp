param (
    [string]$action = "build"
)

$jsonContent = Get-Content -Raw -Path "codemeta.json" | ConvertFrom-Json
$projectName = $jsonContent.name
$versionNo = $jsonContent.version

function Make-Man {
    $markdownFiles = Get-ChildItem -File *.1.md 
    foreach ($file in $markdownFiles) {
        $manName = [System.IO.Path]::GetFileNameWithoutExtension($file.Name)
        
        if (-not (Test-Path -Path man\man1)) {
            New-Item -ItemType Directory -Path man\man1 | Out-Null
        }

        Write-Host "Rending $file as man\man1\$manName"
        pandoc -f Markdown -t man -o man\man1\$manName -s $file
    }
}

function Build-It {
    param (
        [string]$OutPath,
        [string]$SourcePath,
        [string]$TargetOS,
        [string]$TargetArch
    )
    if ([string]::IsNullOrEmpty($OutPath)) {
        throw "missing required output path"
    }
    if ([string]::IsNullOrEmpty($SourcePath)) {
        throw "missing required go source path"
    }

    # Default to host OS if not specified
    if ([string]::IsNullOrEmpty($TargetOS)) {
        if ($IsWindows) {
            $TargetOS = "windows"
        }
        elseif ($IsMacOS) {
            $TargetOS = "darwin"
        }
        elseif ($IsLinux) {
            $TargetOS = "linux"
        }
        else {
            throw "Unsupported host OS."
        }
    }

    # Default to host architecture if not specified
    if ([string]::IsNullOrEmpty($TargetArch)) {
        $arch = $env:PROCESSOR_ARCHITECTURE
        $archMap = @{
            "AMD64"  = "amd64"
            "x86_64" = "amd64"
            "ARM64"  = "arm64"
            "aarch64" = "arm64"
        }
        if ($archMap.ContainsKey($arch)) {
            $TargetArch = $archMap[$arch]
        }
        else {
            # Fallback for Unix-like systems
            $uname = uname -m
            if ($uname -eq "x86_64") { $TargetArch = "amd64" }
            elseif ($uname -eq "aarch64") { $TargetArch = "arm64" }
            else { throw "Unsupported host architecture: $arch" }
        }
    }

    # Validate supported OS and architecture
    $supportedOS = @("windows", "darwin", "linux")
    $supportedArch = @("amd64", "arm64")

    if ($TargetOS -notin $supportedOS) {
        Write-Error "Unsupported OS: $TargetOS. Use windows, darwin, or linux."
        exit 1
    }
    if ($TargetArch -notin $supportedArch) {
        Write-Error "Unsupported architecture: $TargetArch. Use amd64 or arm64."
        exit 1
    }

    Write-Host "Building $OutPath from $SourcePath for OS: $TargetOS, Architecture: $TargetArch"

    # Set GOOS and GOARCH for cross-compilation
    $env:GOOS = $TargetOS
    $env:GOARCH = $TargetArch
    # Run the Go build command
    go build -o "$OutPath" "$SourcePath" 
    if (-not $?) {
        throw "Build failed for $TargetOS/$TargetArch."
    }
}

# Make the man pages
if ($action -eq "man") {
    Make-Man
}

# Default build action
if ($action -eq "build") {
    Write-Host "Building antenna..."
    Build-It -OutPath bin/antenna.exe -SourcePath cmd\antenna\antenna.go
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed"
        exit $LASTEXITCODE
    }

    Write-Host "Generating help documentation..."
    .\bin\antenna.exe --help > antenna.1.md
    Write-Host "Build and documentation generation complete."
}

# Install action
if ($action -eq "install") {
    $binDir = Join-Path $HOME "bin"
    $exePath = Join-Path $PSScriptRoot "bin\antenna.exe"
    $destPath = Join-Path $binDir "antenna.exe"
    Build-It -OutPath $exePath -SourcePath cmd\antenna\antenna.go

    # Create bin directory if it doesn't exist
    if (-not (Test-Path $binDir)) {
        Write-Host "Creating $binDir directory..."
        New-Item -ItemType Directory -Path $binDir | Out-Null
    }

    # Copy executable
    Write-Host "Copying antenna.exe to $binDir..."
    Copy-Item -Path $exePath -Destination $destPath -Force

    # Check if $HOME\bin is in PATH
    $pathEnv = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($pathEnv -notlike "*$binDir*") {
        Write-Host "`n$binDir is not in your PATH. To add it, run the following command:"
        Write-Host "[Environment]::SetEnvironmentVariable('PATH', `'$pathEnv;$binDir`' + ';', 'User')"
        Write-Host "After running the above command, restart your terminal or run:"
        Write-Host "refreshenv"
    } else {
        Write-Host "`n$binDir is already in your PATH."
    }
}

if ($action -eq "release") {
    $releasePath = "dist"
    if (Test-Path -Path $releasePath) {
        Write-Host "Removing stale $releasePath"
        Remove-Item -Path "$releasePath" -Recurse -Force
    }
    New-Item -ItemType Directory -Path $releasePath | Out-Null
    Write-Host "Created directory: $releasePath"

    # Copy in the documentation files
    copy README.md dist\
    copy INSTALL.md dist\
    copy codemeta.json dist\
    copy LICENSE dist\
    copy *.?.md dist\
    copy -Recurse man dist\

    # Build Windows on x86_64
    New-Item -ItemType Directory -Path "$releasePath\bin"
    Build-It -OutPath dist\bin\antenna.exe `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS windows `
      -TargetArch amd64
    cd dist
    # Get all items (files and directories) in the current directory
    $filesToZip = @(
        "bin\",
        "man\",
        "*.md",
        "codemeta.json",
        "INSTALL.md",
        "LICENSE",
        "README.md"
    )
    $targetZip = "$projectName-v$versionNo-Windows-x86_64.zip"
    if (Test-Path -Path $targetZip) {
        Remove-Item -Path "$targetZip" -Force
    }
    # Zip everything, preserving paths
    Compress-Archive -Path $filesToZip -DestinationPath  $targetZip -CompressionLevel Optimal
    cd ..
    Remove-Item -Path "dist\bin" -Recurse -Force

    # Build Windows on arm64
    New-Item -ItemType Directory -Path "$releasePath\bin"
    Build-It -OutPath dist\bin\antenna.exe `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS windows `
      -TargetArch arm64
    cd dist
    $filesToZip = @(
        "bin\",
        "man\",
        "*.md",
        "codemeta.json",
        "INSTALL.md",
        "LICENSE",
        "README.md"
    )
    $targetZip = "$projectName-v$versionNo-Windows-arm64.zip"
    if (Test-Path -Path $targetZip) {
        Remove-Item -Path "$targetZip" -Force
    }
    Compress-Archive -Path $filesToZip -DestinationPath  $targetZip -CompressionLevel Optimal
    cd ..
    Remove-Item -Path "dist\bin" -Recurse -Force

    # Build macOS on x86_64
    New-Item -ItemType Directory -Path "$releasePath\bin"
    Build-It -OutPath dist\bin\antenna `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS darwin `
      -TargetArch amd64
    cd dist
    $filesToZip = @(
        "bin\",
        "man\",
        "*.md",
        "codemeta.json",
        "INSTALL.md",
        "LICENSE",
        "README.md"
    )
    $targetZip = "$projectName-v$versionNo-macOS-x86_64.zip"
    if (Test-Path -Path $targetZip) {
        Remove-Item -Path "$targetZip" -Force
    }
    Compress-Archive -Path $filesToZip -DestinationPath  $targetZip -CompressionLevel Optimal
    cd ..
    Remove-Item -Path "dist\bin" -Recurse -Force

    # Build macOS on arm64
    New-Item -ItemType Directory -Path "$releasePath\bin"
    Build-It -OutPath dist\bin\antenna `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS darwin `
      -TargetArch arm64
    cd dist
    $filesToZip = @(
        "bin\",
        "man\",
        "*.md",
        "codemeta.json",
        "INSTALL.md",
        "LICENSE",
        "README.md"
    )
    $targetZip = "$projectName-v$versionNo-macOS-arm64.zip"
    if (Test-Path -Path $targetZip) {
        Remove-Item -Path "$targetZip" -Force
    }
    Compress-Archive -Path $filesToZip -DestinationPath  $targetZip -CompressionLevel Optimal
    cd ..
    Remove-Item -Path "dist\bin" -Recurse -Force

    # Build Linux on x86_64
    New-Item -ItemType Directory -Path "$releasePath\bin"
    Build-It -OutPath dist\bin\antenna `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS linux `
      -TargetArch amd64
    cd dist
    $filesToZip = @(
        "bin\",
        "man\",
        "*.md",
        "codemeta.json",
        "INSTALL.md",
        "LICENSE",
        "README.md"
    )
    $targetZip = "$projectName-v$versionNo-Linux-x86_64.zip"
    if (Test-Path -Path $targetZip) {
        Remove-Item -Path "$targetZip" -Force
    }
    Compress-Archive -Path $filesToZip -DestinationPath  $targetZip -CompressionLevel Optimal
    cd ..
    Remove-Item -Path "dist\bin" -Recurse -Force

    # Build Linux on arm64
    New-Item -ItemType Directory -Path "$releasePath\bin"
    Build-It -OutPath dist\bin\antenna `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS linux `
      -TargetArch arm64
    cd dist
    $filesToZip = @(
        "bin\",
        "man\",
        "*.md",
        "codemeta.json",
        "INSTALL.md",
        "LICENSE",
        "README.md"
    )
    $targetZip = "$projectName-v$versionNo-Linux-arm64.zip"
    if (Test-Path -Path $targetZip) {
        Remove-Item -Path "$targetZip" -Force
    }
    Compress-Archive -Path $filesToZip -DestinationPath  $targetZip -CompressionLevel Optimal
    cd ..
    Remove-Item -Path "dist\bin" -Recurse -Force
    Write-Host "Check the zip files, then do release.ps1 if all is OK"
}