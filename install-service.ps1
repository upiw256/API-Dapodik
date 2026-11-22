#Requires -RunAsAdministrator

$ErrorActionPreference = 'Stop'
$taskName = 'DapodikAPI'
$exePath = Join-Path $PSScriptRoot 'dapodik-api.exe'
$envPath = Join-Path $PSScriptRoot '.env'

if (-not (Test-Path -LiteralPath $exePath -PathType Leaf)) {
    throw "Binary API tidak ditemukan: $exePath"
}
if (-not (Test-Path -LiteralPath $envPath -PathType Leaf)) {
    throw "File .env tidak ditemukan: $envPath"
}

$action = New-ScheduledTaskAction `
    -Execute $exePath `
    -WorkingDirectory $PSScriptRoot

$trigger = New-ScheduledTaskTrigger -AtStartup

$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -StartWhenAvailable `
    -RestartCount 3 `
    -RestartInterval (New-TimeSpan -Minutes 1) `
    -ExecutionTimeLimit (New-TimeSpan -Days 3650)

# Jalankan sebagai SYSTEM saat boot tanpa perlu login user
$principal = New-ScheduledTaskPrincipal `
    -UserId 'NT AUTHORITY\SYSTEM' `
    -LogonType ServiceAccount `
    -RunLevel Highest

if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
    Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
}

Register-ScheduledTask `
    -TaskName $taskName `
    -Description 'Read-only REST API untuk data Dapodik' `
    -Action $action `
    -Trigger $trigger `
    -Settings $settings `
    -Principal $principal `
    -Force | Out-Null

Start-ScheduledTask -TaskName $taskName
Start-Sleep -Seconds 2
Get-ScheduledTask -TaskName $taskName | Select-Object TaskName, State
Get-ScheduledTaskInfo -TaskName $taskName | Select-Object TaskName, LastRunTime, LastTaskResult, NextRunTime