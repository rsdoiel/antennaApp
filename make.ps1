param (
    [string]$action = "build"
)

# Default build action
if ($action -eq "build") {
    Write-Host "Building antenna..."
    go build -o bin/antenna.exe cmd\antenna\antenna.go
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
