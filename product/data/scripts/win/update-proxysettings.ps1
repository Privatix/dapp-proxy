<#
.SYNOPSIS
    Set/Restore system proxy configuration
.DESCRIPTION
    Set/Restore system proxy configuration

.PARAMETER Action
    "set" or "restore" system proxy configuration

.PARAMETER ProxyOffSettingsPath
    Path to file, where previous proxy configuration settings are stored (backup).

.PARAMETER LocalSocksPort
    Port number of inbound local socks proxy

.EXAMPLE
    .\update-proxysettings.ps1 -Action set -ProxyOffSettingsPath "C:\Program Files\Privatix\Client\product\881da45b-ce8c-46bf-943d-730e9cee5740\config\proxysettings_backup.json" -LocalSocksPort 10081

    Description
    -----------
    Set Windows system proxy settings and backup previous settings to file.

.EXAMPLE
    .\update-proxysettings.ps1 -Action restore -ProxyOffSettingsPath "C:\Program Files\Privatix\Client\product\881da45b-ce8c-46bf-943d-730e9cee5740\config\proxysettings_backup.json"

    Description
    -----------
    Restore previous Windows system proxy settings from backup file.
#>

[cmdletbinding(
        DefaultParameterSetName='set'
    )]
param (
    [Parameter(Mandatory = $true)]
    [ValidateSet('set','restore')]
    [string]$Action,
    [Parameter(Mandatory = $true)]
    [string]$ProxyOffSettingsPath,
    [ValidateRange(0, 65535)] 
    [int]$LocalSocksPort
)

$error.Clear()

if ($Action -eq 'set') {
    # Get current proxy configuration
    $ProxyEnable = Get-ItemPropertyValue -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyEnable
    $ProxyOverride = Get-ItemPropertyValue -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyOverride
    $ProxyServer = Get-ItemPropertyValue -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyServer

    try {
        # backup previous proxy configuration
        $ProxyOffSettings = [PSCustomObject]@{
            ProxyEnable = $ProxyEnable
            ProxyOverride = $ProxyOverride
            ProxyServer = $ProxyServer
        }
        $ProxyOffSettingsJson = ConvertTo-Json -InputObject $ProxyOffSettings 
        $Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False    
        [System.IO.File]::WriteAllLines($ProxyOffSettingsPath, $ProxyOffSettingsJson, $Utf8NoBomEncoding)
        # Set new proxy configuration
        Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyEnable -Value 1
        Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyOverride -Value '<local>'
        Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyServer -Value "socks=127.0.0.1:$LocalSocksPort"
    }
    catch {
        if ($ProxyEnable) {Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyEnable -Value $ProxyEnable}
        if ($ProxyOverride) {Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyOverride -Value $ProxyOverride}
        if ($ProxyServer) {Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyServer -Value $ProxyServer}
    }
    finally {
        if ($Error.Count -gt 0) {   
            Write-Error $error[0].Exception
        }
    }
}

if ($Action -eq 'restore') {
    # Get previous settings from backup
    $ProxyOffSettings = Get-Content -Path $ProxyOffSettingsPath  | ConvertFrom-Json
    # Set proxy configuration from backup
    Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyEnable -Value $ProxyOffSettings.ProxyEnable
    Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyOverride -Value $ProxyOffSettings.ProxyOverride
    Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -Name ProxyServer -Value $ProxyOffSettings.ProxyServer
}