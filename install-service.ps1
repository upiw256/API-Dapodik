#Requires -RunAsAdministrator

$ErrorActionPreference = 'Stop'
$serviceName = 'DapodikAPI'
$exePath = Join-Path $PSScriptRoot 'dapodik-api.exe'
$envPath = Join-Path $PSScriptRoot '.env'

if (-not (Test-Path -LiteralPath $exePath -PathType Leaf)) {
    throw "Binary API tidak ditemukan: $exePath"
}
if (-not (Test-Path -LiteralPath $envPath -PathType Leaf)) {
    throw "File .env tidak ditemukan: $envPath"
}

$service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
if ($null -eq $service) {
    New-Service -Name $serviceName `
        -DisplayName 'Dapodik API' `
        -Description 'Read-only REST API untuk data Dapodik' `
        -BinaryPathName ('"{0}"' -f $exePath) `
        -StartupType Automatic | Out-Null
} else {
    if ($service.Status -ne 'Stopped') {
        Stop-Service -Name $serviceName -Force
        (Get-Service -Name $serviceName).WaitForStatus('Stopped', [TimeSpan]::FromSeconds(30))
    }
    $binaryPath = '"{0}"' -f $exePath
    & sc.exe config $serviceName "binPath= $binaryPath" 'start= auto' | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Gagal memperbarui konfigurasi service $serviceName (sc.exe exit $LASTEXITCODE)"
    }
    Set-Service -Name $serviceName -StartupType Automatic
}

& sc.exe failure $serviceName reset= 86400 actions= restart/60000/restart/60000/restart/60000 | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Gagal mengatur restart otomatis saat API berhenti (sc.exe exit $LASTEXITCODE)"
}

Start-Service -Name $serviceName
Get-Service -Name $serviceName | Select-Object Name, DisplayName, Status, StartType
