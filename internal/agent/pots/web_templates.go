package pots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

// This list mirrors every static Web entry in
// honeypot-templates-server/services/config.json. The configuration file and
// files on disk remain authoritative at runtime; the list only makes the
// service codes available to the Agent factory before a pack is loaded.
var webStaticTemplateList = []string{
	"ac-sangfor", "baota", "canal", "cisco-vpn", "cloudreve", "confluence", "coremail", "cpanel",
	"edr-sangfor", "electric", "esxi", "exchange", "filebrowser", "fw-360", "fw-haofeng", "fw-nsfocus",
	"fw-topsec", "fw-zkww", "gitlab", "gophish", "huorong-zd", "iis", "intel-am", "iot-hikcam",
	"isport", "jenkins", "jira", "joomla", "jspspy", "jumpserver", "kelai-qll", "kibana", "mailu",
	"nagios", "nas-qnap", "nginx", "oa", "oa-gov", "oa-tongda", "oa-yy", "phpadmin", "portainer",
	"poste", "printer-dell", "qzsec", "router-aruba", "router-cmcc", "router-h3c", "router-ikuai",
	"router-openwrt", "router-ruijie", "router-tplink", "routos", "ruoyi", "sangfor-fcg", "sangfor-vpn",
	"synology-nas", "tdp", "thinkphp", "tomcat", "uniaccess-lr", "weblogic", "webmin", "websphere",
	"wordpress", "zabbix", "zhongke-kongzhi", "zimbra",
}

var webTemplateCodes = func() map[string]struct{} {
	items := make(map[string]struct{}, len(webStaticTemplateList))
	for _, code := range webStaticTemplateList {
		items[code] = struct{}{}
	}
	return items
}()

func init() {
	for _, item := range webStaticTemplateList {
		code := item
		if _, exists := factories[code]; exists {
			panic("Web template conflicts with native service: " + code)
		}
		factories[code] = func(provider TLSProvider) Service {
			root := ""
			if source, ok := provider.(TemplateRootProvider); ok {
				root = source.TemplateRoot()
			}
			return &WebTemplateService{code: code, templateRoot: root, tlsProvider: provider}
		}
	}
}

type webTemplateConfig struct {
	Addr  string          `json:"addr"`
	HTTPS json.RawMessage `json:"https"`
	Index string          `json:"index"`
	Root  string          `json:"root"`
}

type resolvedWebTemplate struct {
	Code       string
	Root       string
	Index      string
	Secure     bool
	Configured string
}

func loadWebTemplateManifest(templateRoot string) (map[string]webTemplateConfig, error) {
	templateRoot = strings.TrimSpace(templateRoot)
	if templateRoot == "" {
		return nil, errors.New("Web template root is not configured")
	}
	raw, err := os.ReadFile(filepath.Join(templateRoot, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("read Web template manifest: %w", err)
	}
	manifest := make(map[string]webTemplateConfig)
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("parse Web template manifest: %w", err)
	}
	return manifest, nil
}

func resolveWebTemplate(templateRoot, code string) (resolvedWebTemplate, error) {
	manifest, err := loadWebTemplateManifest(templateRoot)
	if err != nil {
		return resolvedWebTemplate{}, err
	}
	entry, exists := manifest[code]
	if !exists || entry.Root == "" || entry.Index == "" {
		return resolvedWebTemplate{}, fmt.Errorf("Web template %q is not configured", code)
	}
	if !safeRelative(entry.Root) || !safeRelative(entry.Index) {
		return resolvedWebTemplate{}, fmt.Errorf("Web template %q contains an unsafe root or index", code)
	}
	root := filepath.Join(templateRoot, code, filepath.FromSlash(entry.Root))
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return resolvedWebTemplate{}, fmt.Errorf("Web template %q root is unavailable", code)
	}
	index := filepath.Join(root, filepath.FromSlash(entry.Index))
	rootInfo, rootErr := os.Stat(root)
	indexInfo, indexErr := os.Stat(index)
	if rootErr != nil || !rootInfo.IsDir() || indexErr != nil || indexInfo.IsDir() {
		return resolvedWebTemplate{}, fmt.Errorf("Web template %q resources are incomplete", code)
	}
	return resolvedWebTemplate{Code: code, Root: root, Index: filepath.ToSlash(entry.Index), Secure: templateBool(entry.HTTPS), Configured: entry.Addr}, nil
}

func safeRelative(value string) bool {
	if value == "" || filepath.IsAbs(value) || strings.ContainsRune(value, '\x00') {
		return false
	}
	cleaned := path.Clean(strings.ReplaceAll(value, "\\", "/"))
	return cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func templateBool(raw json.RawMessage) bool {
	var value bool
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	var text string
	return json.Unmarshal(raw, &text) == nil && strings.EqualFold(strings.TrimSpace(text), "true")
}

func availableWebTemplates(templateRoot string) map[string]bool {
	available := make(map[string]bool)
	manifest, err := loadWebTemplateManifest(templateRoot)
	if err != nil {
		return available
	}
	for code, entry := range manifest {
		if entry.Root == "" || entry.Index == "" {
			continue
		}
		if _, known := webTemplateCodes[code]; !known {
			continue
		}
		if _, err := resolveWebTemplateEntry(templateRoot, code, entry); err == nil {
			available[code] = true
		}
	}
	return available
}

func resolveWebTemplateEntry(templateRoot, code string, entry webTemplateConfig) (resolvedWebTemplate, error) {
	if !safeRelative(entry.Root) || !safeRelative(entry.Index) {
		return resolvedWebTemplate{}, errors.New("unsafe root or index")
	}
	root := filepath.Join(templateRoot, code, filepath.FromSlash(entry.Root))
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return resolvedWebTemplate{}, errors.New("unavailable root")
	}
	index := filepath.Join(root, filepath.FromSlash(entry.Index))
	rootInfo, rootErr := os.Stat(root)
	indexInfo, indexErr := os.Stat(index)
	if rootErr != nil || !rootInfo.IsDir() || indexErr != nil || indexInfo.IsDir() {
		return resolvedWebTemplate{}, errors.New("incomplete resources")
	}
	return resolvedWebTemplate{Code: code, Root: root, Index: filepath.ToSlash(entry.Index), Secure: templateBool(entry.HTTPS), Configured: entry.Addr}, nil
}

type WebTemplateService struct {
	listener     net.Listener
	server       *http.Server
	code         string
	templateRoot string
	tlsProvider  TLSProvider
	template     resolvedWebTemplate
}

func (s *WebTemplateService) Start(_ context.Context, target protocol.PotTarget, sink Sink) error {
	template, err := resolveWebTemplate(s.templateRoot, s.code)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	if template.Secure {
		listener, err = wrapTLSListener(listener, s.tlsProvider)
		if err != nil {
			return err
		}
	}
	s.template = template
	s.listener = listener
	s.server = &http.Server{
		Handler: http.MaxBytesHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.handle(w, r, sink)
		}), 128<<10),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    64 << 10,
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, localAddressKey{}, connection.LocalAddr())
		},
	}
	go func() { _ = s.server.Serve(listener) }()
	return nil
}

func (s *WebTemplateService) handle(w http.ResponseWriter, r *http.Request, sink Sink) {
	defer r.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 64<<10+1))
	truncated := len(raw) > 64<<10
	if truncated {
		raw = raw[:64<<10]
	}
	values := webRequestValues(r, raw)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	rawRequest, bodyBase64 := httpRequestSnapshot(r, raw, truncated)
	payload := map[string]any{
		"method": r.Method, "path": r.URL.RequestURI(), "host": r.Host, "scheme": scheme,
		"user_agent": r.UserAgent(), "headers": r.Header, "body": string(raw), "body_truncated": truncated,
		"template_code": s.code, "template_source": "honeypot-templates-server",
		"raw_request": rawRequest,
	}
	if bodyBase64 != "" {
		payload["body"] = "[binary body]"
		payload["body_base64"] = bodyBase64
	}
	sink(protocol.NewEvent("web.request", endpoint(remoteAddr(r.RemoteAddr)), endpoint(localAddr(r.Context())), payload, "http", s.code, "web-template"))

	username, password, mechanism := webCredentials(r, values)
	if username != "" || password != "" {
		sink(protocol.NewEvent("web.credential", endpoint(remoteAddr(r.RemoteAddr)), endpoint(localAddr(r.Context())), map[string]any{
			"username": username, "password": password, "mechanism": mechanism, "path": r.URL.Path,
			"template_code": s.code, "template_source": "honeypot-templates-server",
		}, "credential", s.code, "web-template"))
	}

	asset, err := s.resolveRequestAsset(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	response := readWebTemplateResponse(asset + ".res")
	copyWebTemplateHeaders(w.Header(), response.Headers, r)
	if w.Header().Get("Content-Type") == "" {
		extensionPath := strings.SplitN(filepath.ToSlash(asset), "？", 2)[0]
		contentType := mime.TypeByExtension(filepath.Ext(extensionPath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
	}
	file, err := os.Open(asset)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	if info, err := file.Stat(); err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	}
	status := response.Status
	if status < 200 || status > 599 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = io.Copy(w, file)
	}
}

func (s *WebTemplateService) resolveRequestAsset(r *http.Request) (string, error) {
	requestPath := strings.ReplaceAll(r.URL.Path, "\\", "/")
	if strings.ContainsRune(requestPath, '\x00') || strings.HasSuffix(requestPath, ".req") || strings.HasSuffix(requestPath, ".res") {
		return "", os.ErrNotExist
	}
	cleaned := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	candidates := make([]string, 0, 6)
	if cleaned == "." || cleaned == "" {
		candidates = append(candidates, s.template.Index)
	} else {
		if r.URL.RawQuery != "" {
			candidates = append(candidates, cleaned+"？"+r.URL.RawQuery)
		}
		base := path.Base(cleaned)
		directory := path.Dir(cleaned)
		methodPath := path.Join(directory, r.Method+"__"+base, "index.html")
		candidates = append(candidates, methodPath, cleaned)
	}
	for _, candidate := range candidates {
		resolved, err := containedFile(s.template.Root, candidate)
		if err == nil {
			return resolved, nil
		}
		resolved, err = containedFile(s.template.Root, path.Join(candidate, "index.html"))
		if err == nil {
			return resolved, nil
		}
	}
	// Web templates frequently post credentials or bootstrap requests to a
	// dynamic endpoint without shipping a matching static response. Keep any
	// explicit METHOD__ route authoritative, then preserve the decoy page for
	// unmatched mutations instead of navigating the visitor to a bare 404.
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return containedFile(s.template.Root, s.template.Index)
	}
	return "", os.ErrNotExist
}

func containedFile(root, candidate string) (string, error) {
	if !safeRelative(candidate) {
		return "", os.ErrPermission
	}
	resolved := filepath.Join(root, filepath.FromSlash(candidate))
	resolved, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", os.ErrNotExist
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", os.ErrPermission
	}
	info, err := os.Stat(resolved)
	if err != nil || info.IsDir() {
		return "", os.ErrNotExist
	}
	return resolved, nil
}

type webTemplateResponse struct {
	Status  int
	Headers map[string]string
}

func readWebTemplateResponse(file string) webTemplateResponse {
	raw, err := os.ReadFile(file)
	if err != nil {
		return webTemplateResponse{}
	}
	values := make(map[string]any)
	if json.Unmarshal(raw, &values) != nil {
		return webTemplateResponse{}
	}
	result := webTemplateResponse{Headers: make(map[string]string)}
	for key, value := range values {
		text := fmt.Sprint(value)
		if strings.EqualFold(key, "status-code") {
			result.Status, _ = strconv.Atoi(text)
			continue
		}
		result.Headers[key] = text
	}
	return result
}

func copyWebTemplateHeaders(destination http.Header, source map[string]string, request *http.Request) {
	for name, value := range source {
		switch strings.ToLower(name) {
		case "content-length", "content-encoding", "transfer-encoding", "connection", "date":
			continue
		case "location":
			if parsed, err := url.Parse(value); err == nil && parsed.IsAbs() {
				value = parsed.RequestURI()
			}
		}
		destination.Set(name, value)
	}
}

func (s *WebTemplateService) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
