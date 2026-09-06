<#
.SYNOPSIS
  Capture what the container status re-render needs from a live Docker broker host.

.DESCRIPTION
  Read-only. Writes .\docker-capture\.

  Uses Set-Content -Encoding utf8 rather than `>`: in Windows PowerShell 5.1 `>`
  writes UTF-16 and wraps long lines at the console width, which corrupts JSON.

  Secret VALUES are never captured. `docker inspect` DOES carry the container's
  environment, which on docker is where compose-sourced secrets are passed -- so
  the Env array is stripped from the saved copy before it is written. Please still
  skim the files before sharing.

.EXAMPLE
  .\capture-docker.ps1 -Container solace
#>
[CmdletBinding()]
param(
  [string]$Container = '',
  [string]$OutDir = 'docker-capture'
)

$ErrorActionPreference = 'Continue'
if (-not (Test-Path $OutDir)) { New-Item -ItemType Directory -Path $OutDir | Out-Null }

function Save {
  param([string]$File, [string[]]$Lines)
  Set-Content -Path (Join-Path $OutDir $File) -Value $Lines -Encoding utf8
  Write-Host "  $File -- ok" -ForegroundColor Green
}

Write-Host "capturing docker broker state"

# 1. Every container, so the report knows what a real ps row looks like.
Save 'ps.txt'   ([string[]](& docker ps --all))
Save 'ps-format.txt' ([string[]](& docker ps --all --format '{{.Names}}`t{{.Image}}`t{{.Status}}'))

# 2. The full inspect, MINUS the environment (which carries secret values on docker).
$names = @()
if ($Container) { $names = @($Container) }
else { $names = @(& docker ps --all --format '{{.Names}}') }

foreach ($n in $names) {
  if (-not $n) { continue }
  $json = (& docker inspect $n) -join "`n"
  if ($LASTEXITCODE -ne 0) { Write-Host "  inspect $n -- SKIPPED" -ForegroundColor Yellow; continue }
  try {
    $obj = $json | ConvertFrom-Json
    foreach ($o in $obj) {
      # Strip the secret-bearing environment before anything is written to disk.
      if ($o.Config -and $o.Config.PSObject.Properties.Name -contains 'Env') {
        $o.Config.Env = @('(stripped by capture-docker.ps1: may contain secret values)')
      }
    }
    Save ("inspect-$n.json") ([string[]]($obj | ConvertTo-Json -Depth 40))
  } catch {
    Write-Host "  inspect $n -- could not parse, skipped rather than risk writing secrets" -ForegroundColor Yellow
  }
}

# 3. Version/info, for the report's engine line.
Save 'version.txt' ([string[]](& docker version))
Save 'compose-ps.txt' ([string[]](& docker compose ps 2>$null))

Write-Host ""
Write-Host "wrote $OutDir\ -- the container Env array was stripped, but please still"
Write-Host "skim the files for anything host-specific before sharing."
