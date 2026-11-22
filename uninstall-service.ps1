#Requires -RunAsAdministrator
Stop-ScheduledTask -TaskName 'DapodikAPI' -ErrorAction SilentlyContinue
Unregister-ScheduledTask -TaskName 'DapodikAPI' -Confirm:$false -ErrorAction SilentlyContinue
Get-Process -Name 'dapodik-api' -ErrorAction SilentlyContinue | Stop-Process -Force