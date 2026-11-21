#Requires -RunAsAdministrator

$ErrorActionPreference = 'Stop'
$serviceName = 'DapodikAPI'
$service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
if ($null -eq $service) {
    Write-Output "Service $serviceName tidak ditemukan."
    exit 0
}
if ($service.Status -ne 'Stopped') {
    Stop-Service -Name $serviceName -Force
    (Get-Service -Name $serviceName).WaitForStatus('Stopped', [TimeSpan]::FromSeconds(30))
}
& sc.exe delete $serviceName | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Gagal menghapus service $serviceName (sc.exe exit $LASTEXITCODE)"
}
Write-Output "Service $serviceName telah dihapus. Binary dan .env tidak dihapus."
