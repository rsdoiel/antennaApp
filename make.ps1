param (
    [string]$action = "build"
)

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
    $releasePath = "dist\bin"
    if (-not (Test-Path -Path $releasePath)) {
        New-Item -ItemType Directory -Path $releasePath | Out-Null
        Write-Host "Created directory: $releasePath"
    } else {
        Write-Host "Directory already exists: $releasePath"
    }
    # Build Windows on x86_64
    Build-It -OutPath dist\bin\antenna.exe `
      -SourcePath cmd\antenna\antenna.go `
      -TargetOS windows `
      -TargetArch amd64
}