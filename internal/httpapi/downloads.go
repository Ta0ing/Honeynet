package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (a *API) downloadAgent(c *gin.Context) {
	path, name, ok := a.agentBuild(c.Param("os"), c.Param("arch"))
	if !ok {
		fail(c, 404, "BUILD_NOT_FOUND", "暂不支持该操作系统或架构")
		return
	}
	if _, err := os.Stat(path); err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "该平台 Agent 尚未构建")
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.FileAttachment(path, name)
}

func (a *API) downloadAgentChecksum(c *gin.Context) {
	path, _, ok := a.agentBuild(c.Param("os"), c.Param("arch"))
	if !ok {
		fail(c, 404, "BUILD_NOT_FOUND", "暂不支持该操作系统或架构")
		return
	}
	file, err := os.Open(path)
	if err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "该平台 Agent 尚未构建")
		return
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		fail(c, 500, "DOWNLOAD_ERROR", "Agent 校验和生成失败")
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.String(http.StatusOK, "%s\n", hex.EncodeToString(hash.Sum(nil)))
}

func (a *API) downloadAgentGuard(c *gin.Context) {
	path, name, ok := a.agentGuardBuild(c.Param("os"), c.Param("arch"))
	if !ok {
		fail(c, 404, "BUILD_NOT_FOUND", "该平台不使用独立 Agent 守护器")
		return
	}
	if _, err := os.Stat(path); err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "该平台 Agent 守护器尚未构建")
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.FileAttachment(path, name)
}

func (a *API) downloadAgentGuardChecksum(c *gin.Context) {
	path, _, ok := a.agentGuardBuild(c.Param("os"), c.Param("arch"))
	if !ok {
		fail(c, 404, "BUILD_NOT_FOUND", "该平台不使用独立 Agent 守护器")
		return
	}
	file, err := os.Open(path)
	if err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "该平台 Agent 守护器尚未构建")
		return
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		fail(c, 500, "DOWNLOAD_ERROR", "Agent 守护器校验和生成失败")
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.String(http.StatusOK, "%s\n", hex.EncodeToString(hash.Sum(nil)))
}

func (a *API) downloadTemplateBundle(c *gin.Context) {
	path, name, ok := a.templateBundle(c.Param("format"))
	if !ok {
		fail(c, 404, "BUILD_NOT_FOUND", "蜜罐模板资源包不存在")
		return
	}
	if _, err := os.Stat(path); err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "蜜罐模板资源包尚未构建")
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.FileAttachment(path, name)
}

func (a *API) downloadTemplateBundleChecksum(c *gin.Context) {
	path, _, ok := a.templateBundle(c.Param("format"))
	if !ok {
		fail(c, 404, "BUILD_NOT_FOUND", "蜜罐模板资源包不存在")
		return
	}
	file, err := os.Open(path)
	if err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "蜜罐模板资源包尚未构建")
		return
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		fail(c, 500, "DOWNLOAD_ERROR", "蜜罐模板资源包校验和生成失败")
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.String(http.StatusOK, "%s\n", hex.EncodeToString(hash.Sum(nil)))
}

func (a *API) templateBundle(format string) (string, string, bool) {
	name := "honeypot-templates-server."
	switch format {
	case "tar.gz":
		name += "tar.gz"
	case "zip":
		name += "zip"
	default:
		return "", "", false
	}
	base, err := filepath.Abs(a.cfg.DownloadsDir)
	if err != nil {
		return "", "", false
	}
	return filepath.Join(base, name), name, true
}

func (a *API) agentBuild(osName, arch string) (string, string, bool) {
	allowed := map[string]bool{"linux/386": true, "linux/amd64": true, "linux/arm": true, "linux/arm64": true, "linux/loong64": true, "windows/386": true, "windows/amd64": true}
	if !allowed[osName+"/"+arch] {
		return "", "", false
	}
	name := "honeynet-agent-" + osName + "-" + arch
	if osName == "windows" {
		name += ".exe"
	}
	base, err := filepath.Abs(a.cfg.DownloadsDir)
	if err != nil {
		return "", "", false
	}
	return filepath.Join(base, name), name, true
}

func (a *API) agentGuardBuild(osName, arch string) (string, string, bool) {
	allowed := map[string]bool{"linux/386": true, "linux/amd64": true, "linux/arm": true, "linux/arm64": true, "linux/loong64": true}
	if !allowed[osName+"/"+arch] {
		return "", "", false
	}
	name := "honeynet-agent-guard-" + osName + "-" + arch
	base, err := filepath.Abs(a.cfg.DownloadsDir)
	if err != nil {
		return "", "", false
	}
	return filepath.Join(base, name), name, true
}

func (a *API) installShell(c *gin.Context) {
	c.Header("Content-Type", "text/x-shellscript; charset=utf-8")
	c.String(http.StatusOK, `#!/bin/sh
set -eu
SERVER=""
AGENT_URL=""
CA_SHA256=""
NODE_ID=""
TOKEN=""
CONSOLE_CA=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --server) SERVER="$2"; shift 2 ;;
    --agent-url) AGENT_URL="$2"; shift 2 ;;
    --ca-sha256) CA_SHA256="$2"; shift 2 ;;
    --node-id) NODE_ID="$2"; shift 2 ;;
    --token) TOKEN="$2"; shift 2 ;;
    --console-ca) CONSOLE_CA="$2"; shift 2 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done
[ -n "$SERVER" ] && [ -n "$AGENT_URL" ] && [ -n "$CA_SHA256" ] && [ -n "$NODE_ID" ] && [ -n "$TOKEN" ] || { echo "server, agent-url, ca-sha256, node-id and token are required" >&2; exit 2; }
[ -z "$CONSOLE_CA" ] || [ -r "$CONSOLE_CA" ] || { echo "console CA is not readable" >&2; exit 2; }
[ "$(id -u)" -eq 0 ] || { echo "Honeynet Agent installer must run as root" >&2; exit 1; }
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
MACHINE=$(uname -m)
case "$MACHINE" in
  x86_64|amd64) ARCH=amd64 ;;
  i386|i686) ARCH=386 ;;
  aarch64|arm64) ARCH=arm64 ;;
  armv7*|armv6*) ARCH=arm ;;
  loongarch64) ARCH=loong64 ;;
  *) echo "unsupported architecture: $MACHINE" >&2; exit 1 ;;
esac
PREFIX=/opt/honeynet-agent
BIN=$PREFIX/bin/honeynet-agent
GUARD=$PREFIX/libexec/honeynet-agent-guard
LEGACY_BIN=/usr/local/bin/honeynet-agent
BIN_DOWNLOAD=/tmp/honeynet-agent-$$
CONFIG_DIR=/etc/honeynet
STATE_DIR=/var/lib/honeynet-agent
TEMPLATE_DIR=$PREFIX/templates/web
TEMPLATE_ARCHIVE=/tmp/honeynet-web-templates-$$.tar.gz
GUARD_DOWNLOAD=/tmp/honeynet-agent-guard-$$
trap 'rm -f "$BIN_DOWNLOAD" "$TEMPLATE_ARCHIVE" "$GUARD_DOWNLOAD"' EXIT INT TERM
mkdir -p "$CONFIG_DIR" "$STATE_DIR" "$PREFIX/bin" "$TEMPLATE_DIR"
if command -v curl >/dev/null 2>&1; then
  download_file() { if [ -n "$CONSOLE_CA" ]; then curl -fsSL --cacert "$CONSOLE_CA" "$1" -o "$2"; else curl -fsSL "$1" -o "$2"; fi; }
  download_text() { if [ -n "$CONSOLE_CA" ]; then curl -fsSL --cacert "$CONSOLE_CA" "$1"; else curl -fsSL "$1"; fi; }
  download_file "$SERVER/download/agent/$OS/$ARCH" "$BIN_DOWNLOAD"
  EXPECTED=$(download_text "$SERVER/download/agent/$OS/$ARCH/sha256")
  download_file "$SERVER/download/templates/tar.gz" "$TEMPLATE_ARCHIVE"
  TEMPLATE_EXPECTED=$(download_text "$SERVER/download/templates/tar.gz/sha256")
  download_file "$SERVER/download/agent-guard/$OS/$ARCH" "$GUARD_DOWNLOAD"
  GUARD_EXPECTED=$(download_text "$SERVER/download/agent-guard/$OS/$ARCH/sha256")
elif command -v wget >/dev/null 2>&1; then
  download_file() { if [ -n "$CONSOLE_CA" ]; then wget --ca-certificate="$CONSOLE_CA" -qO "$2" "$1"; else wget -qO "$2" "$1"; fi; }
  download_text() { if [ -n "$CONSOLE_CA" ]; then wget --ca-certificate="$CONSOLE_CA" -qO- "$1"; else wget -qO- "$1"; fi; }
  download_file "$SERVER/download/agent/$OS/$ARCH" "$BIN_DOWNLOAD"
  EXPECTED=$(download_text "$SERVER/download/agent/$OS/$ARCH/sha256")
  download_file "$SERVER/download/templates/tar.gz" "$TEMPLATE_ARCHIVE"
  TEMPLATE_EXPECTED=$(download_text "$SERVER/download/templates/tar.gz/sha256")
  download_file "$SERVER/download/agent-guard/$OS/$ARCH" "$GUARD_DOWNLOAD"
  GUARD_EXPECTED=$(download_text "$SERVER/download/agent-guard/$OS/$ARCH/sha256")
else
  echo "curl or wget is required" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  printf '%s  %s\n' "$EXPECTED" "$BIN_DOWNLOAD" | sha256sum -c - >/dev/null
  printf '%s  %s\n' "$TEMPLATE_EXPECTED" "$TEMPLATE_ARCHIVE" | sha256sum -c - >/dev/null
  printf '%s  %s\n' "$GUARD_EXPECTED" "$GUARD_DOWNLOAD" | sha256sum -c - >/dev/null
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "$BIN_DOWNLOAD" | awk '{print $1}')
  [ "$ACTUAL" = "$EXPECTED" ] || { echo "Agent checksum verification failed" >&2; exit 1; }
  TEMPLATE_ACTUAL=$(shasum -a 256 "$TEMPLATE_ARCHIVE" | awk '{print $1}')
  [ "$TEMPLATE_ACTUAL" = "$TEMPLATE_EXPECTED" ] || { echo "Template checksum verification failed" >&2; exit 1; }
  GUARD_ACTUAL=$(shasum -a 256 "$GUARD_DOWNLOAD" | awk '{print $1}')
  [ "$GUARD_ACTUAL" = "$GUARD_EXPECTED" ] || { echo "Agent guard checksum verification failed" >&2; exit 1; }
else
  echo "sha256sum or shasum is required" >&2
  exit 1
fi
if command -v systemctl >/dev/null 2>&1; then
  systemctl stop honeynet-agent.service >/dev/null 2>&1 || true
fi
mkdir -p "$PREFIX/bin" "$PREFIX/libexec"
if [ -f "$BIN" ]; then
  cp -p "$BIN" "$BIN.install-backup"
elif [ -f "$LEGACY_BIN" ]; then
  cp -p "$LEGACY_BIN" "$BIN.install-backup"
fi
restore_agent_install() {
  if [ -f "$BIN.install-backup" ]; then
    mv -f "$BIN.install-backup" "$BIN"
    if command -v systemctl >/dev/null 2>&1; then
      systemctl start honeynet-agent.service >/dev/null 2>&1 || true
    fi
  fi
}
trap 'status=$?; if [ "$status" -ne 0 ]; then restore_agent_install; fi; rm -f "$BIN_DOWNLOAD" "$TEMPLATE_ARCHIVE" "$GUARD_DOWNLOAD"; exit "$status"' EXIT INT TERM
install -m 0755 "$BIN_DOWNLOAD" "$BIN"
# The stable guard is downloaded and verified before stopping an existing
# service, so a transient network failure cannot leave a node offline.
install -m 0755 "$GUARD_DOWNLOAD" "$GUARD"
rm -f "$GUARD_DOWNLOAD"
rm -rf "$TEMPLATE_DIR/services"
tar -xzf "$TEMPLATE_ARCHIVE" -C "$TEMPLATE_DIR"
[ -f "$TEMPLATE_DIR/services/config.json" ] || { echo "Template package is incomplete" >&2; exit 1; }
"$BIN" --config "$CONFIG_DIR/agent.json" --server "$SERVER" --agent-url "$AGENT_URL" --ca-sha256 "$CA_SHA256" --node-id "$NODE_ID" --registration-token "$TOKEN" --force-enroll --state-dir "$STATE_DIR" --template-root "$TEMPLATE_DIR/services" --init-only
"$BIN" --config "$CONFIG_DIR/agent.json" --enroll-only
if command -v systemctl >/dev/null 2>&1; then
  cat > /etc/systemd/system/honeynet-agent.service <<'UNIT'
[Unit]
Description=Honeynet Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/opt/honeynet-agent/libexec/honeynet-agent-guard /opt/honeynet-agent/bin/honeynet-agent --config /etc/honeynet/agent.json --state-dir /var/lib/honeynet-agent
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/etc/honeynet /var/lib/honeynet-agent /opt/honeynet-agent/bin
ReadOnlyPaths=-/opt/honeynet-agent/templates/web

[Install]
WantedBy=multi-user.target
UNIT
  systemctl daemon-reload
  systemctl enable --now honeynet-agent
else
  nohup "$BIN" --config "$CONFIG_DIR/agent.json" >/var/log/honeynet-agent.log 2>&1 &
fi
rm -f "$BIN.install-backup"
echo "Honeynet Agent installed for node $NODE_ID"
`)
}

func (a *API) installPowerShell(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, `function Install-HoneynetAgent {
	param([Parameter(Mandatory=$true)][string]$Server,[Parameter(Mandatory=$true)][string]$AgentURL,[Parameter(Mandatory=$true)][string]$CASHA256,[Parameter(Mandatory=$true)][string]$NodeID,[Parameter(Mandatory=$true)][string]$Token)
  $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
  $base = Join-Path $env:ProgramData "Honeynet"
  $bin = Join-Path $base "honeynet-agent.exe"
  $binDownload = Join-Path $env:TEMP "honeynet-agent.exe.new"
  $config = Join-Path $base "agent.json"
	$templateBase = Join-Path $base "templates\web"
	$templateRoot = Join-Path $templateBase "services"
  $templateArchive = Join-Path $env:TEMP "honeynet-web-templates.zip"
  New-Item -ItemType Directory -Force -Path $base,$templateBase | Out-Null
  Invoke-WebRequest -UseBasicParsing "$Server/download/agent/windows/$arch" -OutFile $binDownload
  $expected = (Invoke-WebRequest -UseBasicParsing "$Server/download/agent/windows/$arch/sha256").Content.Trim()
  $actual = (Get-FileHash -Algorithm SHA256 $binDownload).Hash.ToLowerInvariant()
  if ($actual -ne $expected.ToLowerInvariant()) { throw "Honeynet Agent checksum verification failed" }
	Invoke-WebRequest -UseBasicParsing "$Server/download/templates/zip" -OutFile $templateArchive
	$templateExpected = (Invoke-WebRequest -UseBasicParsing "$Server/download/templates/zip/sha256").Content.Trim()
	$templateActual = (Get-FileHash -Algorithm SHA256 $templateArchive).Hash.ToLowerInvariant()
	if ($templateActual -ne $templateExpected.ToLowerInvariant()) { throw "Honeypot template checksum verification failed" }
  & sc.exe stop HoneynetAgent 2>$null | Out-Null
  & sc.exe delete HoneynetAgent 2>$null | Out-Null
  for ($i = 0; $i -lt 20; $i++) {
    & sc.exe query HoneynetAgent 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) { break }
    Start-Sleep -Milliseconds 250
  }
  Move-Item -Force $binDownload $bin
	if (Test-Path $templateRoot) { Remove-Item -Recurse -Force $templateRoot }
	Expand-Archive -Force $templateArchive $templateBase
	Remove-Item -Force $templateArchive
  if (-not (Test-Path (Join-Path $templateRoot "config.json"))) { throw "Honeypot template package is incomplete" }
  & $bin --config $config --server $Server --agent-url $AgentURL --ca-sha256 $CASHA256 --node-id $NodeID --registration-token $Token --force-enroll --state-dir (Join-Path $base "state") --template-root $templateRoot --init-only
  if ($LASTEXITCODE -ne 0) { throw "Failed to initialize Honeynet Agent" }
  & $bin --config $config --enroll-only
  if ($LASTEXITCODE -ne 0) { throw "Failed to enroll Honeynet Agent with the IPv4/IPv6 gateway" }
  $command = '"' + $bin + '" --config "' + $config + '"'
  & sc.exe create HoneynetAgent binPath= $command start= auto DisplayName= 'Honeynet Agent' | Out-Null
  & sc.exe failure HoneynetAgent reset= 86400 actions= restart/5000/restart/5000/restart/5000 | Out-Null
  & sc.exe start HoneynetAgent | Out-Null
  Write-Host "Honeynet Agent installed for node $NodeID"
}
`)
}
