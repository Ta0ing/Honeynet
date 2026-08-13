param(
  [string]$DatabaseDSN = $env:HONEYPOT_DATABASE_DSN,
  [string]$PublicURL = $env:HONEYPOT_PUBLIC_URL,
  [string]$Listen = ":8080",
  [switch]$ConsoleTLS,
  [switch]$NoConsoleTLS,
  [string]$AgentListen = ":8443",
  [string]$AgentPublicURL = $env:HONEYPOT_AGENT_PUBLIC_URL,
  [string]$AdminUser = $(if ($env:HONEYPOT_ADMIN_USERNAME) { $env:HONEYPOT_ADMIN_USERNAME } else { "admin" }),
  [string]$AdminPassword = $env:HONEYPOT_ADMIN_PASSWORD,
  [string]$JWTSecret = $env:HONEYPOT_JWT_SECRET,
  [string]$BuiltinToken = $env:HONEYPOT_BUILTIN_AGENT_TOKEN,
  [string]$IPIPDBPath = $env:HONEYPOT_IPIP_DB_PATH,
  [string]$IPIPLanguage = $(if ($env:HONEYPOT_IPIP_LANGUAGE) { $env:HONEYPOT_IPIP_LANGUAGE } else { "CN" }),
  [bool]$ThreatIntelEnabled = $(if ($env:HONEYPOT_THREAT_INTEL_ENABLED) { $env:HONEYPOT_THREAT_INTEL_ENABLED -in @('1','true','yes','on') } else { $false }),
  [string]$ThreatIntelDBPath = $(if ($env:HONEYPOT_THREAT_INTEL_DB_PATH) { $env:HONEYPOT_THREAT_INTEL_DB_PATH } else { "$env:ProgramData\Honeynet\threat-intelligence.db" }),
  [string]$ThreatIntelDownloadURL = $env:HONEYPOT_THREAT_INTEL_DOWNLOAD_URL,
  [string]$ThreatIntelUpdateInterval = $(if ($env:HONEYPOT_THREAT_INTEL_UPDATE_INTERVAL) { $env:HONEYPOT_THREAT_INTEL_UPDATE_INTERVAL } else { "24h" }),
  [switch]$ServerOnly,
  [switch]$SkipBuiltinAgent,
  [switch]$NoStart
)

$ErrorActionPreference = "Stop"
if ($ServerOnly) { $SkipBuiltinAgent = $true }
if ($ConsoleTLS -and $NoConsoleTLS) { throw 'ConsoleTLS and NoConsoleTLS cannot be used together.' }
$consoleTLSEnabled = $ConsoleTLS.IsPresent
if ($NoConsoleTLS) {
  $consoleTLSEnabled = $false
} elseif (-not $ConsoleTLS -and $env:HONEYPOT_TLS_ENABLED) {
  switch ($env:HONEYPOT_TLS_ENABLED.Trim().ToLowerInvariant()) {
    { $_ -in @('1', 'true', 'yes', 'on') } { $consoleTLSEnabled = $true; break }
    { $_ -in @('0', 'false', 'no', 'off') } { $consoleTLSEnabled = $false; break }
    default { throw 'HONEYPOT_TLS_ENABLED must be true or false.' }
  }
}
$principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
  throw "Run this installer from an Administrator PowerShell."
}

function ConvertFrom-HoneynetListenAddress([string]$Value, [string]$Name) {
  $hostName = ""
  $portText = ""
  if ($Value -match '^:(\d+)$') {
    $portText = $Matches[1]
  } elseif ($Value -match '^\[([^\]]+)\]:(\d+)$') {
    $hostName, $portText = $Matches[1], $Matches[2]
  } elseif ($Value -match '^([^:]+):(\d+)$') {
    $hostName, $portText = $Matches[1], $Matches[2]
  } else {
    throw "$Name must use host:PORT syntax; bracket IPv6, for example [::]:8080"
  }
  $port = [int]$portText
  if ($port -lt 1 -or $port -gt 65535) { throw "$Name port must be between 1 and 65535" }
  return [PSCustomObject]@{ Host = $hostName; Port = $port }
}

function Get-HoneynetProbeHost([string]$HostName) {
  if (-not $HostName -or $HostName -eq '0.0.0.0') { return '127.0.0.1' }
  if ($HostName -eq '::') { return '::1' }
  return $HostName
}

function Format-HoneynetURLHost([string]$HostName) {
  if ($HostName.Contains(':')) { return "[$HostName]" }
  return $HostName
}

$listenEndpoint = ConvertFrom-HoneynetListenAddress $Listen 'Listen'
$agentListenEndpoint = ConvertFrom-HoneynetListenAddress $AgentListen 'AgentListen'
$listenPort = $listenEndpoint.Port
$agentPort = $agentListenEndpoint.Port
$consoleProbeHost = Get-HoneynetProbeHost $listenEndpoint.Host
$agentProbeHost = Get-HoneynetProbeHost $agentListenEndpoint.Host

function New-RandomHex {
  $bytes = New-Object byte[] 32
  $random = [Security.Cryptography.RandomNumberGenerator]::Create()
  try { $random.GetBytes($bytes) } finally { $random.Dispose() }
  return ([BitConverter]::ToString($bytes) -replace '-', '').ToLowerInvariant()
}

function ConvertTo-YAMLQuoted([string]$Value) {
  return "'" + $Value.Replace("'", "''") + "'"
}

function Get-HoneynetServerTLSEnabled([string]$Path) {
  $inServer = $false
  foreach ($line in Get-Content -LiteralPath $Path) {
    if ($line -match '^\s*#') { continue }
    if ($line -match '^\S') {
      $inServer = $line -match '^server\s*:'
      continue
    }
    if ($inServer -and $line -match '^\s+tls_enabled\s*:\s*([^#]+)') {
      $value = $Matches[1].Trim().Trim([char[]]@([char]39, [char]34)).ToLowerInvariant()
      return $value -in @('1', 'true', 'yes', 'on')
    }
  }
  return $false
}

function Test-HoneynetCAHealth([Uri]$URI, [string]$CAPath) {
  if (-not (Test-Path -LiteralPath $CAPath -PathType Leaf)) { return $false }
  $curl = Get-Command curl.exe -ErrorAction SilentlyContinue
  if ($curl) {
    & $curl.Source --fail --silent --show-error --noproxy '*' --max-time 2 --cacert $CAPath $URI.AbsoluteUri 2>$null | Out-Null
    return $LASTEXITCODE -eq 0
  }
  $caCertificate = New-Object -TypeName Security.Cryptography.X509Certificates.X509Certificate2 -ArgumentList $CAPath
  $validationCallback = {
    param($sender, $certificate, $chain, $policyErrors)
    if (($policyErrors -band [Net.Security.SslPolicyErrors]::RemoteCertificateNameMismatch) -ne 0) { return $false }
    $leaf = New-Object -TypeName Security.Cryptography.X509Certificates.X509Certificate2 -ArgumentList $certificate
    $trustedChain = New-Object Security.Cryptography.X509Certificates.X509Chain
    try {
      $trustedChain.ChainPolicy.RevocationMode = [Security.Cryptography.X509Certificates.X509RevocationMode]::NoCheck
      $trustedChain.ChainPolicy.VerificationFlags = [Security.Cryptography.X509Certificates.X509VerificationFlags]::AllowUnknownCertificateAuthority
      $trustedChain.ChainPolicy.ExtraStore.Add($caCertificate)
      $serverAuthOID = New-Object -TypeName Security.Cryptography.Oid -ArgumentList @('1.3.6.1.5.5.7.3.1')
      $trustedChain.ChainPolicy.ApplicationPolicy.Add($serverAuthOID)
      if (-not $trustedChain.Build($leaf)) { return $false }
      $root = $trustedChain.ChainElements[$trustedChain.ChainElements.Count - 1].Certificate
      return $root.Thumbprint -eq $caCertificate.Thumbprint
    } finally {
      $trustedChain.Dispose()
      $leaf.Dispose()
    }
  }.GetNewClosure()

  $client = New-Object Net.Sockets.TcpClient
  try {
    $client.ReceiveTimeout = 2000
    $client.SendTimeout = 2000
    $client.Connect($URI.DnsSafeHost.Trim([char[]]'[]'), $URI.Port)
    $stream = New-Object Net.Security.SslStream($client.GetStream(), $false, ([Net.Security.RemoteCertificateValidationCallback]$validationCallback))
    try {
      $stream.ReadTimeout = 2000
      $stream.WriteTimeout = 2000
      # The Go listener accepts TLS 1.3 only. A successful handshake therefore
      # also proves that the local Windows TLS stack negotiated TLS 1.3.
      $stream.AuthenticateAsClient($URI.DnsSafeHost.Trim([char[]]'[]'))
      $request = "GET $($URI.PathAndQuery) HTTP/1.1`r`nHost: $($URI.Authority)`r`nConnection: close`r`n`r`n"
      $requestBytes = [Text.Encoding]::ASCII.GetBytes($request)
      $stream.Write($requestBytes, 0, $requestBytes.Length)
      $stream.Flush()
      $reader = New-Object IO.StreamReader($stream, [Text.Encoding]::ASCII, $false, 1024, $true)
      try {
        return ($reader.ReadLine() -match '^HTTP/\d(?:\.\d)? 200(?: |$)')
      } finally {
        $reader.Dispose()
      }
    } finally {
      $stream.Dispose()
    }
  } catch {
    return $false
  } finally {
    $client.Dispose()
    $caCertificate.Dispose()
  }
}

function Remove-HoneynetService([string]$Name) {
  & sc.exe query $Name 2>$null | Out-Null
  if ($LASTEXITCODE -eq 0) {
    & sc.exe stop $Name 2>$null | Out-Null
    Start-Sleep -Seconds 2
    & sc.exe delete $Name 2>$null | Out-Null
    for ($i = 0; $i -lt 20; $i++) {
      & sc.exe query $Name 2>$null | Out-Null
      if ($LASTEXITCODE -ne 0) { return }
      Start-Sleep -Milliseconds 250
    }
    throw "Timed out deleting Windows service $Name"
  }
}

$packageRoot = Split-Path -Parent $PSScriptRoot
if (Test-Path (Join-Path $packageRoot "SERVER_ONLY")) { $SkipBuiltinAgent = $true }
$prefix = Join-Path $env:ProgramFiles "Honeynet"
$configDir = Join-Path $env:ProgramData "Honeynet"
$configPath = Join-Path $configDir "server.yaml"
$agentConfig = Join-Path $configDir "builtin-agent.json"
$agentState = Join-Path $configDir "agent-state"
$pkiDir = Join-Path $configDir "pki"
$serverSource = Join-Path $packageRoot "bin\honeynet-server.exe"
$agentSource = Join-Path $packageRoot "bin\honeynet-agent.exe"
$webSource = Join-Path $packageRoot "web\dist"
$templateSource = Join-Path $packageRoot "templates\web\services"
$ruleSource = Join-Path $packageRoot "rules\builtin"
$bundledIPIPDB = Join-Path $packageRoot "geoip\ipip.ipdb"
if (Test-Path -LiteralPath $configPath -PathType Leaf) {
  # Upgrade installations preserve server.yaml; probe the protocol selected by
  # that file, regardless of command-line switches supplied for this run.
  $consoleTLSEnabled = Get-HoneynetServerTLSEnabled $configPath
}
if (-not $IPIPDBPath -and (Test-Path $bundledIPIPDB)) { $IPIPDBPath = $bundledIPIPDB }
if ($IPIPDBPath -and -not (Test-Path $IPIPDBPath -PathType Leaf)) { throw "IPIP database does not exist: $IPIPDBPath" }
foreach ($required in @($serverSource, (Join-Path $webSource "index.html"))) {
  if (-not (Test-Path $required)) { throw "Incomplete release package: missing $required" }
}
if (-not (Test-Path $ruleSource -PathType Container)) { throw "Incomplete release package: missing $ruleSource" }
if (-not $SkipBuiltinAgent) {
  foreach ($required in @($agentSource, (Join-Path $templateSource "config.json"))) {
    if (-not (Test-Path $required)) { throw "Incomplete built-in Agent package: missing $required (use -ServerOnly to install only Server)" }
  }
}

if (-not $PublicURL) {
  $consoleScheme = if ($consoleTLSEnabled) { 'https' } else { 'http' }
  $PublicURL = if ($listenPort -in @(80, 443)) { "$consoleScheme`://$env:COMPUTERNAME" } else { "$consoleScheme`://$env:COMPUTERNAME`:$listenPort" }
}
if (-not $AgentPublicURL) { $AgentPublicURL = "https://$env:COMPUTERNAME`:$agentPort" }
$publicURI = [Uri]$PublicURL
if (-not $publicURI.IsAbsoluteUri -or $publicURI.Scheme -notin @('http', 'https') -or -not $publicURI.Host) { throw 'PublicURL must be an absolute HTTP URL' }
if ($consoleTLSEnabled -and $publicURI.Scheme -ne 'https') { throw 'PublicURL must use https when ConsoleTLS is enabled.' }
$agentURI = [Uri]$AgentPublicURL
if (-not $agentURI.IsAbsoluteUri -or $agentURI.Scheme -ne 'https' -or -not $agentURI.Host) { throw 'AgentPublicURL must be an absolute HTTPS URL' }
$agentHost = $agentURI.DnsSafeHost.Trim([char[]]'[]')
$consoleProbeScheme = if ($consoleTLSEnabled) { 'https' } else { 'http' }
$consoleProbeURL = "$consoleProbeScheme`://$(Format-HoneynetURLHost $consoleProbeHost):$listenPort"
$agentProbeURL = "https://$(Format-HoneynetURLHost $agentProbeHost):$agentPort"
$newConfig = -not (Test-Path $configPath)
$generatedPassword = $false
if ($newConfig) {
  if (-not $DatabaseDSN) { throw "DatabaseDSN is required for the first installation." }
  if (-not $AdminPassword) { $AdminPassword = New-RandomHex; $generatedPassword = $true }
  if (-not $JWTSecret) { $JWTSecret = New-RandomHex }
  if (-not $BuiltinToken) { $BuiltinToken = New-RandomHex }
}

if (-not $SkipBuiltinAgent) { Remove-HoneynetService "HoneynetAgent" }
Remove-HoneynetService "HoneynetServer"
New-Item -ItemType Directory -Force -Path (Join-Path $prefix "bin"), (Join-Path $prefix "web"), (Join-Path $prefix "downloads"), (Join-Path $prefix "templates\web"), (Join-Path $prefix "rules\builtin"), $configDir, $agentState, $pkiDir | Out-Null
$installedIPIPDB = Join-Path $configDir "ipip.ipdb"
if ($IPIPDBPath) {
  $resolvedIPIPDB = (Resolve-Path $IPIPDBPath).Path
  if ($resolvedIPIPDB -ne $installedIPIPDB) { Copy-Item -Force $resolvedIPIPDB $installedIPIPDB }
}
Copy-Item -Force $serverSource (Join-Path $prefix "bin\honeynet-server.exe")
Copy-Item -Recurse -Force $webSource (Join-Path $prefix "web")
$installedRuleRoot = Join-Path $prefix "rules\builtin"
Copy-Item -Recurse -Force (Join-Path $ruleSource '*') $installedRuleRoot
$installedTemplateRoot = Join-Path $prefix "templates\web\services"
if (-not $SkipBuiltinAgent) {
  Copy-Item -Force $agentSource (Join-Path $prefix "bin\honeynet-agent.exe")
  if (Test-Path $installedTemplateRoot) { Remove-Item -Recurse -Force $installedTemplateRoot }
  Copy-Item -Recurse -Force $templateSource $installedTemplateRoot
}
if (Test-Path (Join-Path $packageRoot "downloads")) {
  Get-ChildItem -LiteralPath (Join-Path $packageRoot "downloads") -Force | Copy-Item -Destination (Join-Path $prefix "downloads") -Recurse -Force
}

$serverExe = Join-Path $prefix "bin\honeynet-server.exe"
$agentExe = Join-Path $prefix "bin\honeynet-agent.exe"
if ($newConfig) {
  $agentToken = if ($SkipBuiltinAgent) { "" } else { $BuiltinToken }
  $configuredIPIPDB = if (Test-Path $installedIPIPDB) { $installedIPIPDB } else { "" }
  $yaml = @"
server:
  addr: $(ConvertTo-YAMLQuoted $Listen)
  public_url: $(ConvertTo-YAMLQuoted $PublicURL)
  tls_enabled: $($consoleTLSEnabled.ToString().ToLowerInvariant())
  web_dist: $(ConvertTo-YAMLQuoted (Join-Path $prefix 'web\dist'))
  downloads_dir: $(ConvertTo-YAMLQuoted (Join-Path $prefix 'downloads'))
agent:
  addr: $(ConvertTo-YAMLQuoted $AgentListen)
  public_url: $(ConvertTo-YAMLQuoted $AgentPublicURL)
  pki_dir: $(ConvertTo-YAMLQuoted $pkiDir)
  tls_names:
    - $(ConvertTo-YAMLQuoted $agentHost)
    - $(ConvertTo-YAMLQuoted $agentProbeHost)
    - $(ConvertTo-YAMLQuoted $consoleProbeHost)
    - 'localhost'
    - '127.0.0.1'
    - '::1'
  certificate_validity: '9600h'
  renew_before: '720h'
geoip:
  ipip_db_path: $(ConvertTo-YAMLQuoted $configuredIPIPDB)
  language: $(ConvertTo-YAMLQuoted $IPIPLanguage)
threat_intelligence:
  enabled: $($ThreatIntelEnabled.ToString().ToLowerInvariant())
  database_path: $(ConvertTo-YAMLQuoted $ThreatIntelDBPath)
  download_url: $(ConvertTo-YAMLQuoted $ThreatIntelDownloadURL)
  update_interval: $(ConvertTo-YAMLQuoted $ThreatIntelUpdateInterval)
detection:
  rules_dir: $(ConvertTo-YAMLQuoted $installedRuleRoot)
ai:
  enabled: false
  provider: 'openai-compatible'
  base_url: ''
  api_key: ''
  model: ''
  timeout: '45s'
  send_raw_packet: false
database:
  dsn: $(ConvertTo-YAMLQuoted $DatabaseDSN)
auth:
  jwt_secret: $(ConvertTo-YAMLQuoted $JWTSecret)
  jwt_expires: '8h'
  admin_username: $(ConvertTo-YAMLQuoted $AdminUser)
  admin_password: $(ConvertTo-YAMLQuoted $AdminPassword)
builtin_agent:
  token: $(ConvertTo-YAMLQuoted $agentToken)
cors:
  origins:
    - $(ConvertTo-YAMLQuoted $PublicURL)
"@
  [IO.File]::WriteAllText($configPath, $yaml, (New-Object Text.UTF8Encoding($false)))
  & icacls.exe $configPath /inheritance:r /grant:r '*S-1-5-18:F' '*S-1-5-32-544:F' | Out-Null
  if (-not $SkipBuiltinAgent) {
    & $agentExe --config $agentConfig --server $consoleProbeURL --agent-url $agentProbeURL --node-id "00000000-0000-4000-8000-000000000001" --registration-token $BuiltinToken --ca-cert (Join-Path $pkiDir 'ca.crt') --state-dir $agentState --template-root $installedTemplateRoot --init-only
    if ($LASTEXITCODE -ne 0) { throw "Failed to initialize the built-in Agent" }
  }
} else {
  Write-Host "Preserving existing configuration: $configPath"
}

$serverCommand = '"' + $serverExe + '" --config "' + $configPath + '"'
# Windows SCM does not read systemd-style environment files. Administrators
# should store HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD in the machine-level
# environment before starting the service; it is intentionally absent here and
# from server.yaml so the archive password is never written by the installer.
& sc.exe create HoneynetServer binPath= $serverCommand start= auto DisplayName= 'Honeynet Management Server' | Out-Null
if ($LASTEXITCODE -ne 0) { throw "Failed to register HoneynetServer" }
& sc.exe failure HoneynetServer reset= 86400 actions= restart/5000/restart/5000/restart/5000 | Out-Null

if (-not $SkipBuiltinAgent -and (Test-Path $agentConfig)) {
  $agentCommand = '"' + $agentExe + '" --config "' + $agentConfig + '"'
  & sc.exe create HoneynetAgent binPath= $agentCommand start= auto depend= HoneynetServer DisplayName= 'Honeynet Built-in Agent' | Out-Null
  if ($LASTEXITCODE -ne 0) { throw "Failed to register HoneynetAgent" }
  & sc.exe failure HoneynetAgent reset= 86400 actions= restart/5000/restart/5000/restart/5000 | Out-Null
}

if (-not $NoStart) {
  & sc.exe start HoneynetServer | Out-Null
  $healthy = $false
  for ($attempt = 0; $attempt -lt 60; $attempt++) {
    try {
      if ($consoleTLSEnabled) {
        if (-not (Test-HoneynetCAHealth ([Uri]"$consoleProbeURL/healthz") (Join-Path $pkiDir 'ca.crt'))) { throw 'Console TLS health probe failed' }
      } else {
        Invoke-WebRequest -UseBasicParsing -Uri "$consoleProbeURL/healthz" -TimeoutSec 2 | Out-Null
      }
      if (-not (Test-HoneynetCAHealth ([Uri]"$agentProbeURL/healthz") (Join-Path $pkiDir 'ca.crt'))) { throw 'Agent TLS health probe failed' }
      $healthy = $true
      break
    } catch {
      Start-Sleep -Seconds 2
    }
  }
  if (-not $healthy) { throw 'Honeynet console and Agent gateway did not pass startup probes within 120 seconds.' }
  if (-not $SkipBuiltinAgent -and (Test-Path $agentConfig)) { & sc.exe start HoneynetAgent | Out-Null }
}

Write-Host "Honeynet Server installed at $prefix"
Write-Host "Console: $PublicURL"
if ($consoleTLSEnabled) {
  Write-Host "Console TLS uses the Honeynet CA at $(Join-Path $pkiDir 'ca.crt'); import that CA into administrator browsers before opening the console."
}
Write-Host "Administrator: $AdminUser"
if ($generatedPassword) {
  Write-Host "Generated administrator password: $AdminPassword"
  Write-Host "Save this password now; it will not be displayed again."
}
if ($SkipBuiltinAgent) {
  Write-Host "Status: Get-Service HoneynetServer"
} else {
  Write-Host "Status: Get-Service HoneynetServer,HoneynetAgent"
}
