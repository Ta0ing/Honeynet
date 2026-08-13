package threatintel

import (
	"archive/zip"
	"compress/flate"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const maxArchiveBytes = 128 << 20

var ErrUpdateInProgress = errors.New("intelligence database update already in progress")

type Config struct {
	Enabled         bool
	DatabasePath    string
	DownloadURL     string
	ArchivePassword string
	UpdateInterval  time.Duration
	HTTPClient      *http.Client
}

type Status struct {
	Enabled          bool       `json:"enabled"`
	Loaded           bool       `json:"loaded"`
	Updating         bool       `json:"updating"`
	Source           string     `json:"source"`
	RecordCount      int        `json:"record_count"`
	IPv4IPv6         bool       `json:"ipv4_ipv6"`
	LookupMode       string     `json:"lookup_mode,omitempty"`
	DatabaseUpdated  *time.Time `json:"database_updated_at,omitempty"`
	InstalledAt      *time.Time `json:"installed_at,omitempty"`
	LastAttemptAt    *time.Time `json:"last_attempt_at,omitempty"`
	LastSuccessfulAt *time.Time `json:"last_successful_at,omitempty"`
	NextUpdateAt     *time.Time `json:"next_update_at,omitempty"`
	DownloadReady    bool       `json:"download_ready"`
	LastError        string     `json:"last_error,omitempty"`
}

type metadata struct {
	SourceUpdatedAt time.Time `json:"source_updated_at"`
	InstalledAt     time.Time `json:"installed_at"`
	ArchiveSHA256   string    `json:"archive_sha256"`
}

type Manager struct {
	cfg      Config
	database atomic.Pointer[Database]
	updateMu sync.Mutex
	statusMu sync.RWMutex
	status   Status
}

func NewManager(cfg Config) (*Manager, error) {
	if cfg.UpdateInterval == 0 {
		cfg.UpdateInterval = 24 * time.Hour
	}
	if cfg.UpdateInterval < time.Hour || cfg.UpdateInterval > 30*24*time.Hour {
		return nil, errors.New("threat intelligence update interval must be between 1h and 720h")
	}
	if cfg.Enabled && strings.TrimSpace(cfg.DatabasePath) == "" {
		return nil, errors.New("threat intelligence database path is required when enabled")
	}
	if cfg.DownloadURL != "" {
		parsed, err := url.Parse(cfg.DownloadURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			return nil, errors.New("threat intelligence download URL must be an HTTPS URL without user information")
		}
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 5 * time.Minute, CheckRedirect: secureRedirect}
	}
	manager := &Manager{cfg: cfg, status: Status{Enabled: cfg.Enabled, Source: "免费社区威胁情报库", IPv4IPv6: true, DownloadReady: strings.TrimSpace(cfg.DownloadURL) != "" && cfg.ArchivePassword != ""}}
	if !cfg.Enabled {
		return manager, nil
	}
	if _, err := os.Stat(cfg.DatabasePath); err == nil {
		database, loadErr := Load(cfg.DatabasePath)
		if loadErr != nil {
			manager.status.LastError = loadErr.Error()
		} else {
			manager.installInMemory(database)
			manager.loadMetadata()
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		manager.status.LastError = fmt.Sprintf("inspect intelligence database: %v", err)
	}
	return manager, nil
}

func (m *Manager) Start(ctx context.Context) {
	if !m.cfg.Enabled {
		return
	}
	m.statusMu.Lock()
	if !m.status.DownloadReady && m.status.LastError == "" {
		m.status.LastError = "自动更新尚未配置完整；需要 HTTPS 下载地址和受保护的解压密码环境变量"
	}
	m.statusMu.Unlock()
	go func() {
		if m.database.Load() == nil && m.canDownload() {
			updateCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			_ = m.Update(updateCtx)
			cancel()
		}
		ticker := time.NewTicker(m.cfg.UpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if m.canDownload() {
					updateCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
					_ = m.Update(updateCtx)
					cancel()
				}
			}
		}
	}()
}

func (m *Manager) Lookup(address string) (Result, bool) {
	if database := m.database.Load(); database != nil {
		return database.Lookup(address)
	}
	return Result{}, false
}

func (m *Manager) Status() Status {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	return m.status
}

func (m *Manager) Update(ctx context.Context) error {
	if !m.cfg.Enabled {
		return errors.New("threat intelligence database is disabled")
	}
	if !m.updateMu.TryLock() {
		return ErrUpdateInProgress
	}
	defer m.updateMu.Unlock()
	now := time.Now()
	m.statusMu.Lock()
	m.status.Updating = true
	m.status.LastAttemptAt = &now
	m.status.LastError = ""
	m.statusMu.Unlock()
	err := m.update(ctx)
	m.statusMu.Lock()
	m.status.Updating = false
	if err != nil {
		m.status.LastError = err.Error()
	} else {
		success := time.Now()
		next := success.Add(m.cfg.UpdateInterval)
		m.status.LastSuccessfulAt = &success
		m.status.NextUpdateAt = &next
	}
	m.statusMu.Unlock()
	return err
}

func (m *Manager) update(ctx context.Context) error {
	if !m.canDownload() {
		return errors.New("download URL and archive password must be configured through Server settings")
	}
	directory := filepath.Dir(m.cfg.DatabasePath)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create intelligence database directory: %w", err)
	}
	archive, err := os.CreateTemp(directory, ".threat-intel-*.zip")
	if err != nil {
		return fmt.Errorf("create temporary intelligence archive: %w", err)
	}
	archivePath := archive.Name()
	defer os.Remove(archivePath)
	if err := archive.Chmod(0o600); err != nil {
		archive.Close()
		return fmt.Errorf("protect temporary intelligence archive: %w", err)
	}
	hash := sha256.New()
	if err := m.download(ctx, io.MultiWriter(archive, hash)); err != nil {
		archive.Close()
		return err
	}
	if err := archive.Sync(); err != nil {
		archive.Close()
		return fmt.Errorf("sync temporary intelligence archive: %w", err)
	}
	if err := archive.Close(); err != nil {
		return fmt.Errorf("close temporary intelligence archive: %w", err)
	}
	candidate, err := os.CreateTemp(directory, ".threat-intel-*.db")
	if err != nil {
		return fmt.Errorf("create temporary intelligence database: %w", err)
	}
	candidatePath := candidate.Name()
	defer os.Remove(candidatePath)
	if err := candidate.Chmod(0o600); err != nil {
		candidate.Close()
		return fmt.Errorf("protect temporary intelligence database: %w", err)
	}
	exportedAt, err := extractDatabaseArchive(archivePath, m.cfg.ArchivePassword, candidate)
	if err != nil {
		candidate.Close()
		return err
	}
	if err := candidate.Sync(); err != nil {
		candidate.Close()
		return fmt.Errorf("sync temporary intelligence database: %w", err)
	}
	if err := candidate.Close(); err != nil {
		return fmt.Errorf("close temporary intelligence database: %w", err)
	}
	database, err := Load(candidatePath)
	if err != nil {
		return fmt.Errorf("validate downloaded intelligence database: %w", err)
	}
	if err := replaceFile(candidatePath, m.cfg.DatabasePath); err != nil {
		return err
	}
	database.path = m.cfg.DatabasePath
	m.installInMemory(database)
	installedAt := time.Now()
	meta := metadata{SourceUpdatedAt: exportedAt, InstalledAt: installedAt, ArchiveSHA256: hex.EncodeToString(hash.Sum(nil))}
	_ = writeMetadata(m.cfg.DatabasePath+".meta.json", meta)
	m.statusMu.Lock()
	m.status.DatabaseUpdated = timePointer(exportedAt)
	m.status.InstalledAt = &installedAt
	m.statusMu.Unlock()
	return nil
}

func (m *Manager) download(ctx context.Context, destination io.Writer) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.DownloadURL, nil)
	if err != nil {
		return errors.New("create intelligence database download request")
	}
	request.Header.Set("User-Agent", "Honeynet-Server/Threat-Intelligence-Updater")
	response, err := m.cfg.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("download intelligence database: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download intelligence database: unexpected HTTP status %d", response.StatusCode)
	}
	if response.ContentLength > maxArchiveBytes {
		return errors.New("download intelligence database: archive exceeds size limit")
	}
	limited := io.LimitReader(response.Body, maxArchiveBytes+1)
	written, err := io.Copy(destination, limited)
	if err != nil {
		return fmt.Errorf("download intelligence database: %w", err)
	}
	if written == 0 || written > maxArchiveBytes {
		return errors.New("download intelligence database: invalid archive size")
	}
	return nil
}

func (m *Manager) installInMemory(database *Database) {
	m.database.Store(database)
	m.statusMu.Lock()
	m.status.Loaded = true
	m.status.RecordCount = database.Count()
	if database.SlowMode() {
		m.status.LookupMode = "兼容查询"
	} else {
		m.status.LookupMode = "快速查询"
	}
	loaded := database.LoadedAt()
	m.status.InstalledAt = &loaded
	m.status.LastError = ""
	m.statusMu.Unlock()
}

func (m *Manager) loadMetadata() {
	data, err := os.ReadFile(m.cfg.DatabasePath + ".meta.json")
	if err != nil {
		return
	}
	var meta metadata
	if json.Unmarshal(data, &meta) != nil {
		return
	}
	m.statusMu.Lock()
	m.status.DatabaseUpdated = timePointer(meta.SourceUpdatedAt)
	m.status.InstalledAt = timePointer(meta.InstalledAt)
	m.statusMu.Unlock()
}

func (m *Manager) canDownload() bool {
	return strings.TrimSpace(m.cfg.DownloadURL) != "" && m.cfg.ArchivePassword != ""
}

func secureRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("too many redirects")
	}
	if request.URL.Scheme != "https" || request.URL.Host == "" || request.URL.User != nil {
		return errors.New("intelligence database redirect must remain on HTTPS")
	}
	if len(via) > 0 {
		previousHost := strings.ToLower(via[len(via)-1].URL.Hostname())
		nextHost := strings.ToLower(request.URL.Hostname())
		// The publisher currently redirects its download endpoint to a dedicated
		// object-storage hostname. Keep that one explicit transition and reject
		// any arbitrary cross-host redirect or subsequent host hopping.
		allowedPublisherRedirect := strings.HasSuffix(previousHost, ".rivers.chaitin.cn") && strings.HasSuffix(nextHost, ".aliyuncs.com")
		if nextHost != previousHost && (!allowedPublisherRedirect || len(via) != 1) {
			return errors.New("intelligence database redirect changed to an untrusted host")
		}
	}
	return nil
}

func replaceFile(candidate, destination string) error {
	backup := destination + ".previous"
	_ = os.Remove(backup)
	hadPrevious := false
	if _, err := os.Stat(destination); err == nil {
		if err := os.Rename(destination, backup); err != nil {
			return fmt.Errorf("backup current intelligence database: %w", err)
		}
		hadPrevious = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect current intelligence database: %w", err)
	}
	if err := os.Rename(candidate, destination); err != nil {
		if hadPrevious {
			_ = os.Rename(backup, destination)
		}
		return fmt.Errorf("activate intelligence database: %w", err)
	}
	return nil
}

func writeMetadata(path string, value metadata) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".threat-intel-meta-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copy := value
	return &copy
}

func extractDatabaseArchive(archivePath, password string, destination io.Writer) (time.Time, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("open intelligence database archive: %w", err)
	}
	defer reader.Close()
	var databaseFile *zip.File
	var readmeFile *zip.File
	for _, file := range reader.File {
		name := filepath.Base(strings.ReplaceAll(file.Name, "\\", "/"))
		if name != file.Name || file.FileInfo().IsDir() {
			return time.Time{}, errors.New("intelligence database archive contains an unsafe path")
		}
		switch {
		case strings.HasSuffix(strings.ToLower(name), ".db"):
			if databaseFile != nil {
				return time.Time{}, errors.New("intelligence database archive contains multiple database files")
			}
			databaseFile = file
		case strings.EqualFold(name, "readme.txt"):
			readmeFile = file
		}
	}
	if databaseFile == nil {
		return time.Time{}, errors.New("intelligence database archive contains no database file")
	}
	if databaseFile.UncompressedSize64 == 0 || databaseFile.UncompressedSize64 > maxDatabaseBytes {
		return time.Time{}, errors.New("intelligence database archive expands beyond the accepted size")
	}
	if _, err := copyZipFile(databaseFile, password, destination, maxDatabaseBytes); err != nil {
		return time.Time{}, fmt.Errorf("extract intelligence database archive: %w", err)
	}
	if readmeFile == nil {
		return time.Time{}, nil
	}
	var readme strings.Builder
	if _, err := copyZipFile(readmeFile, password, &readme, 4096); err != nil {
		return time.Time{}, fmt.Errorf("read intelligence database metadata: %w", err)
	}
	text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(readme.String()), "exported at"))
	exportedAt, _ := time.ParseInLocation("2006-01-02 15:04", text, time.Local)
	return exportedAt, nil
}

func copyZipFile(file *zip.File, password string, destination io.Writer, maxBytes int64) (int64, error) {
	var source io.ReadCloser
	if file.Flags&1 == 0 {
		opened, err := file.Open()
		if err != nil {
			return 0, err
		}
		source = opened
	} else {
		if password == "" {
			return 0, errors.New("archive password is required")
		}
		raw, err := file.OpenRaw()
		if err != nil {
			return 0, err
		}
		decryptor := newZipCryptoReader(raw, []byte(password))
		var encryptionHeader [12]byte
		if _, err := io.ReadFull(decryptor, encryptionHeader[:]); err != nil {
			return 0, errors.New("invalid encrypted ZIP header")
		}
		checkByte := byte(file.CRC32 >> 24)
		if file.Flags&8 != 0 {
			checkByte = byte(file.ModifiedTime >> 8)
		}
		if encryptionHeader[11] != checkByte {
			return 0, errors.New("archive password is incorrect")
		}
		switch file.Method {
		case zip.Store:
			source = io.NopCloser(decryptor)
		case zip.Deflate:
			source = flate.NewReader(decryptor)
		default:
			return 0, fmt.Errorf("unsupported ZIP compression method %d", file.Method)
		}
	}
	defer source.Close()
	hash := crc32.NewIEEE()
	written, err := io.Copy(io.MultiWriter(destination, hash), io.LimitReader(source, maxBytes+1))
	if err != nil {
		return written, err
	}
	if written > maxBytes || uint64(written) != file.UncompressedSize64 {
		return written, errors.New("ZIP entry size does not match its manifest")
	}
	if hash.Sum32() != file.CRC32 {
		return written, errors.New("ZIP entry checksum does not match its manifest")
	}
	return written, nil
}

type zipCryptoReader struct {
	reader io.Reader
	keys   [3]uint32
}

func newZipCryptoReader(reader io.Reader, password []byte) *zipCryptoReader {
	z := &zipCryptoReader{reader: reader, keys: [3]uint32{0x12345678, 0x23456789, 0x34567890}}
	for _, value := range password {
		z.update(value)
	}
	return z
}

func (z *zipCryptoReader) Read(buffer []byte) (int, error) {
	n, err := z.reader.Read(buffer)
	for index := 0; index < n; index++ {
		temporary := z.keys[2] | 2
		plain := buffer[index] ^ byte((temporary*(temporary^1))>>8)
		buffer[index] = plain
		z.update(plain)
	}
	return n, err
}

func (z *zipCryptoReader) update(value byte) {
	z.keys[0] = zipCRC32(z.keys[0], value)
	z.keys[1] += z.keys[0] & 0xff
	z.keys[1] = z.keys[1]*134775813 + 1
	z.keys[2] = zipCRC32(z.keys[2], byte(z.keys[1]>>24))
}

var zipCRCTable = crc32.MakeTable(crc32.IEEE)

func zipCRC32(current uint32, value byte) uint32 {
	return zipCRCTable[byte(current)^value] ^ (current >> 8)
}

// EventTags returns bounded human-readable tags so attack events remain useful
// even before a dedicated intelligence panel is opened.
func EventTags(result Result) []string {
	tags := []string{"威胁情报命中", "情报等级：" + levelLabel(result.Level)}
	seen := map[string]struct{}{}
	for _, tag := range tags {
		seen[tag] = struct{}{}
	}
	appendUnique := func(prefix, value string) {
		value = localizedValue(value)
		if value == "" {
			return
		}
		tag := prefix + value
		if _, exists := seen[tag]; exists {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	for _, label := range result.Labels {
		if len(tags) >= 6 {
			break
		}
		appendUnique("情报标签：", label)
	}
	for _, behavior := range result.Behaviors {
		if len(tags) >= 12 {
			break
		}
		appendUnique("情报行为：", behavior)
	}
	return tags
}

func levelLabel(level int) string {
	switch {
	case level >= 3:
		return "高危"
	case level == 2:
		return "中危"
	case level == 1:
		return "低危"
	default:
		return "信息"
	}
}

func localizedValue(value string) string {
	translations := map[string]string{
		"Hosting": "托管主机", "Home Broadband": "家庭宽带", "Company": "企业网络", "Unused": "闲置地址", "Institution": "机构网络",
		"Unrouted": "未路由地址", "Mobile Network": "移动网络", "WLAN": "无线网络", "University": "高校网络", "Unallocated": "未分配地址",
		"Special Export": "特殊出口", "Anycast": "任播网络", "Infrastructure": "基础设施", "Satellite Communication": "卫星通信", "CDN": "内容分发网络",
		"Port Scanning": "端口扫描", "Vulnerability Scanning": "漏洞扫描", "Spammers": "垃圾信息", "Spider": "网络爬虫", "Abuse": "滥用行为",
		"Proxy": "代理访问", "Exploit": "漏洞利用", "Bruteforce": "暴力破解", "Brute Force": "暴力破解", "Information Gathering": "信息收集",
		"Command Injection": "命令注入", "SSH Bruteforce": "SSH 暴力破解",
	}
	if translated := translations[value]; translated != "" {
		return translated
	}
	return strings.TrimSpace(value)
}
