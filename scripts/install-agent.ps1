param(
  [Parameter(Mandatory = $true)][string]$Server,
  [Parameter(Mandatory = $true)][string]$AgentURL,
  [Parameter(Mandatory = $true)][string]$CASHA256,
  [Parameter(Mandatory = $true)][string]$NodeID,
  [Parameter(Mandatory = $true)][string]$Token,
  [switch]$NoStart
)

$ErrorActionPreference = "Stop"
$principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
  throw "Run this installer from an Administrator PowerShell."
}

function Remove-HoneynetAgentService {
  & sc.exe query HoneynetAgent 2>$null | Out-Null
  if ($LASTEXITCODE -ne 0) { return }
  & sc.exe stop HoneynetAgent 2>$null | Out-Null
  & sc.exe delete HoneynetAgent 2>$null | Out-Null
  for ($i = 0; $i -lt 20; $i++) {
    & sc.exe query HoneynetAgent 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) { return }
    Start-Sleep -Milliseconds 250
  }
  throw "Timed out deleting HoneynetAgent"
}

$packageRoot = Split-Path -Parent $PSScriptRoot
$sourceBin = Join-Path $packageRoot "bin\honeynet-agent.exe"
$sourceTemplates = Join-Path $packageRoot "templates\web\services"
foreach ($required in @($sourceBin, (Join-Path $sourceTemplates "config.json"))) {
  if (-not (Test-Path $required -PathType Leaf)) { throw "Incomplete Agent package: missing $required" }
}

$prefix = Join-Path $env:ProgramFiles "HoneynetAgent"
$dataDir = Join-Path $env:ProgramData "HoneynetAgent"
$bin = Join-Path $prefix "honeynet-agent.exe"
$config = Join-Path $dataDir "agent.json"
$stateDir = Join-Path $dataDir "state"
$templateBase = Join-Path $prefix "templates\web"
$templateRoot = Join-Path $templateBase "services"

Remove-HoneynetAgentService
New-Item -ItemType Directory -Force -Path $prefix, $dataDir, $stateDir, $templateBase | Out-Null
Copy-Item -Force $sourceBin $bin
if (Test-Path $templateRoot) { Remove-Item -Recurse -Force $templateRoot }
Copy-Item -Recurse -Force $sourceTemplates $templateRoot

& $bin --config $config --server $Server --agent-url $AgentURL --ca-sha256 $CASHA256 --node-id $NodeID --registration-token $Token --force-enroll --state-dir $stateDir --template-root $templateRoot --init-only
if ($LASTEXITCODE -ne 0) { throw "Failed to initialize Honeynet Agent" }
& $bin --config $config --enroll-only
if ($LASTEXITCODE -ne 0) { throw "Failed to enroll Honeynet Agent with the IPv4/IPv6 gateway" }

$command = '"' + $bin + '" --config "' + $config + '"'
& sc.exe create HoneynetAgent binPath= $command start= auto DisplayName= "Honeynet Agent" | Out-Null
if ($LASTEXITCODE -ne 0) { throw "Failed to register HoneynetAgent" }
& sc.exe failure HoneynetAgent reset= 86400 actions= restart/5000/restart/5000/restart/5000 | Out-Null
if (-not $NoStart) { & sc.exe start HoneynetAgent | Out-Null }

Write-Host "Honeynet Agent installed for node $NodeID"
Write-Host "Status: Get-Service HoneynetAgent"
