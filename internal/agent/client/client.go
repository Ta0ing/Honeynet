package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/honeynet/honeynet/internal/agent/config"
	"github.com/honeynet/honeynet/internal/agent/decoy"
	"github.com/honeynet/honeynet/internal/agent/potcert"
	"github.com/honeynet/honeynet/internal/agent/pots"
	"github.com/honeynet/honeynet/internal/agent/protocol"
	"github.com/honeynet/honeynet/internal/agent/queue"
	agentruntime "github.com/honeynet/honeynet/internal/agent/runtime"
	"github.com/honeynet/honeynet/internal/agent/sense"
	"github.com/honeynet/honeynet/internal/agentupdate"
	"github.com/honeynet/honeynet/internal/detection"
)

type Client struct {
	cfg          *config.Config
	version      string
	clientMu     sync.RWMutex
	http         *http.Client
	tls          *tls.Config
	queue        *queue.Queue
	runtime      *agentruntime.Runtime
	decoys       *decoy.Manager
	sense        *sense.Manager
	updater      *agentupdate.Manager
	rulesMu      sync.RWMutex
	matcher      *detection.Matcher
	ruleRevision int64
	// serverSilenceTimeout bounds a half-open control channel. It is configurable
	// internally so the watchdog can be exercised without a 90-second test.
	serverSilenceTimeout time.Duration
}

type apiEnvelope[T any] struct {
	Data    T      `json:"data"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type wsEnvelope struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	TS      int64           `json:"ts"`
	Payload json.RawMessage `json:"payload"`
}
type wsSession struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

const (
	agentHeartbeatInterval = 30 * time.Second
	agentServerSilence     = 90 * time.Second
)

func New(cfg *config.Config, version string) (*Client, error) {
	q, err := queue.Open(filepath.Join(cfg.StateDir, "pending-events.json"))
	if err != nil {
		return nil, err
	}
	client := &Client{cfg: cfg, version: version, queue: q, serverSilenceTimeout: agentServerSilence}
	client.updater = &agentupdate.Manager{StateDir: cfg.StateDir, CurrentVersion: version, PublicKey: cfg.UpdatePublicKey, DownloadBaseURL: cfg.AgentURL, ServiceName: "HoneynetAgent"}
	if err := client.updater.CheckStartup(); err != nil {
		return nil, err
	}
	if cfg.HasCertificate() {
		if err := client.configureMTLS(); err != nil {
			return nil, fmt.Errorf("load node certificate: %w", err)
		}
	}
	potCertificates, err := potcert.New(cfg.StateDir)
	if err != nil {
		return nil, fmt.Errorf("initialize pot TLS certificates: %w", err)
	}
	client.runtime = agentruntime.NewWithTLSAndTemplates(client.enqueueEvent, potCertificates, cfg.TemplateRoot)
	client.decoys, err = decoy.New(cfg.StateDir, func(event protocol.Event) {
		client.enqueueEvent(event)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize decoy manager: %w", err)
	}
	client.sense = sense.New(func(event protocol.Event) {
		client.enqueueEvent(event)
	})
	return client, nil
}

func (c *Client) enqueueEvent(event protocol.Event) {
	c.rulesMu.RLock()
	matcher := c.matcher
	ruleRevision := c.ruleRevision
	c.rulesMu.RUnlock()
	event.RuleRevision = ruleRevision
	if matcher != nil {
		hits := matcher.Match(detection.Event{EventType: event.EventType, Service: event.Service, RawPacket: event.RawPacket, Payload: event.Payload}, "agent")
		if len(hits) > 0 {
			event.Detections = append(event.Detections, hits...)
			event.Tags = appendUnique(event.Tags, "detected", "agent-matched")
		}
	}
	if err := c.queue.Add(event); err != nil {
		log.Printf("queue event: %v", err)
	}
}

func appendUnique(values []string, additions ...string) []string {
	seen := make(map[string]bool, len(values)+len(additions))
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

func (c *Client) Run(ctx context.Context) error {
	if err := c.Enroll(ctx); err != nil {
		return err
	}
	if time.Until(c.cfg.CertificateExpiry) <= c.cfg.RenewBefore {
		if err := c.RenewCertificate(ctx); err != nil {
			return fmt.Errorf("renew node certificate: %w", err)
		}
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	healthErrors := make(chan error, 1)
	go func() {
		if err := c.updater.WaitForHealth(runCtx); err != nil {
			healthErrors <- err
			cancel()
		}
	}()
	go c.runUploader(runCtx)
	go c.runCertificateRenewal(runCtx)
	defer c.runtime.Shutdown()
	defer c.decoys.Shutdown()
	defer c.sense.Shutdown()
	if saved, exists, err := c.loadSenseConfig(); err != nil {
		log.Printf("load passive sensing config: %v", err)
	} else if exists {
		if _, err := c.sense.Apply(runCtx, saved); err != nil {
			log.Printf("start passive sensing: %v", err)
		}
	}
	delay := time.Second
	for {
		if ctx.Err() != nil {
			return nil
		}
		established, err := c.connect(runCtx)
		if established {
			// A healthy handshake resets outage backoff. Without this, a node
			// that had several historical disconnects could wait a full minute
			// after every later transient network break.
			delay = time.Second
		}
		if err != nil {
			select {
			case healthErr := <-healthErrors:
				return healthErr
			default:
			}
			if errors.Is(err, agentupdate.ErrRestartRequired) || errors.Is(err, agentupdate.ErrRollbackRequired) {
				return err
			}
			if !errors.Is(err, context.Canceled) {
				log.Printf("control connection closed: %v", err)
			}
		}
		jitter := time.Duration(rand.Int63n(int64(delay/3 + 1)))
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(delay + jitter):
		}
		if delay < time.Minute {
			delay *= 2
		}
		if delay > time.Minute {
			delay = time.Minute
		}
	}
}

func (c *Client) Enroll(ctx context.Context) error {
	if c.cfg.HasCertificate() {
		return nil
	}
	return c.register(ctx)
}

func (c *Client) register(ctx context.Context) error {
	if c.cfg.RegistrationToken == "" {
		return errors.New("registration_token is required for first connection")
	}
	addresses := localIPs()
	body, _ := json.Marshal(map[string]any{"version": c.version, "ip": firstLocalIP(addresses), "ips": addresses, "os": runtime.GOOS, "arch": runtime.GOARCH})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AgentURL+"/agent/v1/register", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-ID", c.cfg.NodeID)
	req.Header.Set("X-Node-Token", c.cfg.RegistrationToken)
	bootstrapClient, err := c.bootstrapHTTPClient()
	if err != nil {
		return fmt.Errorf("prepare registration TLS: %w", err)
	}
	resp, err := bootstrapClient.Do(req)
	if err != nil {
		return fmt.Errorf("register node: %w", err)
	}
	defer resp.Body.Close()
	var result apiEnvelope[struct {
		AgentURL           string    `json:"agent_url"`
		CACertificate      string    `json:"ca_certificate"`
		ClientCertificate  string    `json:"client_certificate"`
		ClientKey          string    `json:"client_key"`
		CertificateExpires time.Time `json:"certificate_expires_at"`
		RenewBeforeSeconds int64     `json:"renew_before_seconds"`
		UpdatePublicKey    string    `json:"update_public_key"`
		UpdateKeyID        string    `json:"update_key_id"`
	}]
	if json.NewDecoder(resp.Body).Decode(&result) != nil || resp.StatusCode != http.StatusOK || result.Data.ClientCertificate == "" || result.Data.ClientKey == "" || result.Data.CACertificate == "" {
		return fmt.Errorf("register node: status %d: %s", resp.StatusCode, result.Message)
	}
	c.cfg.UpdatePublicKey = result.Data.UpdatePublicKey
	c.cfg.UpdateKeyID = result.Data.UpdateKeyID
	c.updater.PublicKey = result.Data.UpdatePublicKey
	if err := c.cfg.SaveEnrollment(c.cfg.AgentURL, []byte(result.Data.CACertificate), []byte(result.Data.ClientCertificate), []byte(result.Data.ClientKey), result.Data.CertificateExpires, time.Duration(result.Data.RenewBeforeSeconds)*time.Second); err != nil {
		return fmt.Errorf("save node credential: %w", err)
	}
	if err := c.configureMTLS(); err != nil {
		return fmt.Errorf("activate node certificate: %w", err)
	}
	log.Printf("node %s registered", c.cfg.NodeID)
	return nil
}

func (c *Client) bootstrapHTTPClient() (*http.Client, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13, ServerName: c.cfg.TLSServerName}
	if c.cfg.CACertPath != "" {
		certificate, err := os.ReadFile(c.cfg.CACertPath)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(certificate) {
			return nil, errors.New("CA certificate is invalid")
		}
		tlsConfig.RootCAs = pool
	} else {
		expected := c.cfg.CAFingerprint
		if expected == "" && !c.cfg.InsecureTLS {
			return nil, errors.New("ca_sha256 or ca_cert_path is required for secure enrollment")
		}
		tlsConfig.InsecureSkipVerify = true // Verification is replaced by the pinned-CA callback below.
		tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
			if expected == "" && c.cfg.InsecureTLS {
				return nil
			}
			return verifyPinnedCA(state, expected, c.cfg.TLSServerName, c.cfg.AgentURL)
		}
	}
	return &http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig}}, nil
}

func verifyPinnedCA(state tls.ConnectionState, expected, serverName, agentURL string) error {
	var authority *x509.Certificate
	for _, certificate := range state.PeerCertificates {
		sum := sha256.Sum256(certificate.Raw)
		if subtleHexEqual(hex.EncodeToString(sum[:]), expected) {
			authority = certificate
			break
		}
	}
	if authority == nil || !authority.IsCA {
		return errors.New("Agent gateway CA fingerprint mismatch")
	}
	if serverName == "" {
		parsed, err := url.Parse(agentURL)
		if err != nil {
			return err
		}
		serverName = parsed.Hostname()
	}
	roots := x509.NewCertPool()
	roots.AddCert(authority)
	intermediates := x509.NewCertPool()
	for _, certificate := range state.PeerCertificates[1:] {
		if !certificate.Equal(authority) {
			intermediates.AddCert(certificate)
		}
	}
	if len(state.PeerCertificates) == 0 {
		return errors.New("Agent gateway did not present a certificate")
	}
	_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{
		Roots: roots, Intermediates: intermediates, DNSName: serverName,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	return err
}

func subtleHexEqual(left, right string) bool {
	leftBytes, leftErr := hex.DecodeString(left)
	rightBytes, rightErr := hex.DecodeString(right)
	if leftErr != nil || rightErr != nil || len(leftBytes) != len(rightBytes) {
		return false
	}
	var different byte
	for i := range leftBytes {
		different |= leftBytes[i] ^ rightBytes[i]
	}
	return different == 0
}

func (c *Client) configureMTLS() error {
	caPEM, err := os.ReadFile(c.cfg.CACertPath)
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return errors.New("node CA certificate is invalid")
	}
	certificate, err := tls.LoadX509KeyPair(c.cfg.ClientCertPath, c.cfg.ClientKeyPath)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      pool,
		Certificates: []tls.Certificate{certificate},
		ServerName:   c.cfg.TLSServerName,
	}
	httpClient := &http.Client{Timeout: 20 * time.Second, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig}}
	c.clientMu.Lock()
	oldClient := c.http
	c.http = httpClient
	c.tls = tlsConfig
	c.clientMu.Unlock()
	if oldClient != nil {
		oldClient.CloseIdleConnections()
	}
	return nil
}

func (c *Client) httpClient() *http.Client {
	c.clientMu.RLock()
	defer c.clientMu.RUnlock()
	return c.http
}

func (c *Client) RenewCertificate(ctx context.Context) error {
	c.clientMu.RLock()
	ready := c.http != nil && c.tls != nil
	c.clientMu.RUnlock()
	if !ready {
		return errors.New("node mTLS client is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AgentURL+"/agent/v1/certificates/renew", http.NoBody)
	if err != nil {
		return err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result apiEnvelope[struct {
		AgentURL           string    `json:"agent_url"`
		CACertificate      string    `json:"ca_certificate"`
		ClientCertificate  string    `json:"client_certificate"`
		ClientKey          string    `json:"client_key"`
		CertificateExpires time.Time `json:"certificate_expires_at"`
		RenewBeforeSeconds int64     `json:"renew_before_seconds"`
		UpdatePublicKey    string    `json:"update_public_key"`
		UpdateKeyID        string    `json:"update_key_id"`
	}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, result.Message)
	}
	if result.Data.UpdatePublicKey != "" {
		c.cfg.UpdatePublicKey = result.Data.UpdatePublicKey
		c.cfg.UpdateKeyID = result.Data.UpdateKeyID
		c.updater.PublicKey = result.Data.UpdatePublicKey
	}
	if err := c.cfg.SaveEnrollment(c.cfg.AgentURL, []byte(result.Data.CACertificate), []byte(result.Data.ClientCertificate), []byte(result.Data.ClientKey), result.Data.CertificateExpires, time.Duration(result.Data.RenewBeforeSeconds)*time.Second); err != nil {
		return err
	}
	if err := c.configureMTLS(); err != nil {
		return err
	}
	if err := c.activateCertificate(ctx); err != nil {
		return fmt.Errorf("activate renewed certificate: %w", err)
	}
	log.Printf("node %s certificate renewed until %s", c.cfg.NodeID, c.cfg.CertificateExpiry.Format(time.RFC3339))
	return nil
}

func (c *Client) activateCertificate(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AgentURL+"/agent/v1/certificates/activate", http.NoBody)
	if err != nil {
		return err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) runCertificateRenewal(ctx context.Context) {
	for {
		wait := time.Until(c.cfg.CertificateExpiry.Add(-c.cfg.RenewBefore))
		if wait > 6*time.Hour {
			wait = 6 * time.Hour
		}
		if wait < time.Minute {
			wait = time.Minute
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		if time.Until(c.cfg.CertificateExpiry) > c.cfg.RenewBefore {
			continue
		}
		if err := c.RenewCertificate(ctx); err != nil && ctx.Err() == nil {
			log.Printf("renew node certificate: %v", err)
		}
	}
}

func (c *Client) connect(ctx context.Context) (bool, error) {
	wsURL, err := url.Parse(c.cfg.AgentURL)
	if err != nil {
		return false, err
	}
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	wsURL.Path = "/agent/v1/ws"
	wsURL.RawQuery = ""
	c.clientMu.RLock()
	tlsConfig := c.tls.Clone()
	c.clientMu.RUnlock()
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second, TLSClientConfig: tlsConfig, Proxy: http.ProxyFromEnvironment}
	conn, response, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		if response != nil {
			return false, fmt.Errorf("websocket handshake: status %d", response.StatusCode)
		}
		return false, err
	}
	established := true
	session := &wsSession{conn: conn}
	defer conn.Close()
	silenceTimeout := c.serverSilenceTimeout
	if silenceTimeout <= 0 {
		silenceTimeout = agentServerSilence
	}
	refreshReadDeadline := func() error { return conn.SetReadDeadline(time.Now().Add(silenceTimeout)) }
	if err := refreshReadDeadline(); err != nil {
		return established, err
	}
	conn.SetPongHandler(func(string) error { return refreshReadDeadline() })
	log.Printf("control channel connected to %s", wsURL.Redacted())
	capabilities := agentCapabilities(c.cfg.TemplateRoot)
	ruleRevision, ruleCount := c.detectionRuleStatus()
	if err := session.write("hello", map[string]any{"version": c.version, "hostname": hostname(), "os": runtime.GOOS, "arch": runtime.GOARCH, "capabilities": capabilities, "rule_revision": ruleRevision, "rule_count": ruleCount}); err != nil {
		return established, err
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	heartbeatErrors := make(chan error, 1)
	go c.heartbeats(done, session, heartbeatErrors)
	for {
		select {
		case err := <-heartbeatErrors:
			return established, err
		default:
		}
		var envelope wsEnvelope
		if err := conn.ReadJSON(&envelope); err != nil {
			return established, err
		}
		if err := refreshReadDeadline(); err != nil {
			return established, err
		}
		switch envelope.Type {
		case "hello.ack":
			var hello struct {
				UpdatePublicKey string `json:"update_public_key"`
				UpdateKeyID     string `json:"update_key_id"`
			}
			if json.Unmarshal(envelope.Payload, &hello) == nil && hello.UpdatePublicKey != "" && (hello.UpdatePublicKey != c.cfg.UpdatePublicKey || hello.UpdateKeyID != c.cfg.UpdateKeyID) {
				c.cfg.UpdatePublicKey = hello.UpdatePublicKey
				c.cfg.UpdateKeyID = hello.UpdateKeyID
				c.updater.PublicKey = hello.UpdatePublicKey
				if err := c.cfg.Save(); err != nil {
					log.Printf("save Agent update trust: %v", err)
				}
			}
			if err := c.updater.ConfirmHealthy(); err != nil {
				log.Printf("confirm Agent update health: %v", err)
			}
			if state := c.updater.Status(); state != nil && (state.Status == "rolled_back" || state.Status == "rollback_failed") {
				errorText := state.LastError
				if errorText == "" {
					errorText = "Agent reverted to the previous version"
				}
				if err := session.write("result", map[string]any{"operation": "upgrade", "task_id": state.TaskID, "rollout_id": state.RolloutID, "status": state.Status, "success": false, "error": errorText}); err != nil {
					return established, err
				}
			}
		case "cmd.pot.apply":
			var payload struct {
				Pots []protocol.PotTarget `json:"pots"`
			}
			if json.Unmarshal(envelope.Payload, &payload) != nil {
				continue
			}
			for _, result := range c.runtime.Apply(ctx, payload.Pots) {
				value := map[string]any{"ref_id": envelope.ID, "pot_id": result.PotID, "status": result.Status, "success": result.Success, "error": result.Error}
				if err := session.write("result", value); err != nil {
					return established, err
				}
			}
		case "cmd.decoy.apply":
			var payload struct {
				Decoys []protocol.DecoyTarget `json:"decoys"`
			}
			if json.Unmarshal(envelope.Payload, &payload) != nil {
				continue
			}
			for _, result := range c.decoys.Apply(ctx, payload.Decoys) {
				value := map[string]any{
					"ref_id": envelope.ID, "operation": "decoy", "decoy_id": result.DecoyID,
					"status": result.Status, "success": result.Success, "error": result.LastError,
					"managed_path": result.ManagedPath,
				}
				if err := session.write("result", value); err != nil {
					return established, err
				}
			}
		case "cmd.sense.apply":
			var payload struct {
				Config protocol.SenseConfig `json:"config"`
			}
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				continue
			}
			normalized, err := sense.NormalizeConfig(payload.Config)
			if err == nil {
				err = c.saveSenseConfig(normalized)
			}
			status := c.sense.Status()
			if err == nil {
				status, err = c.sense.Apply(ctx, normalized)
			}
			result := map[string]any{"ref_id": envelope.ID, "operation": "sense", "status": status.ActualStatus, "success": err == nil, "sense": status}
			if err != nil {
				result["error"] = err.Error()
			}
			if writeErr := session.write("result", result); writeErr != nil {
				return established, writeErr
			}
		case "cmd.rules.apply":
			var payload struct {
				Revision int64            `json:"revision"`
				Rules    []detection.Rule `json:"rules"`
			}
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				continue
			}
			matcher, err := detection.Compile(payload.Rules)
			if err == nil {
				c.rulesMu.Lock()
				c.matcher = matcher
				c.ruleRevision = payload.Revision
				c.rulesMu.Unlock()
			}
			result := map[string]any{"ref_id": envelope.ID, "operation": "rules", "success": err == nil, "revision": payload.Revision, "rule_count": len(payload.Rules)}
			if err != nil {
				result["error"] = err.Error()
			}
			if writeErr := session.write("result", result); writeErr != nil {
				return established, writeErr
			}
		case "cmd.agent.upgrade":
			var command agentupdate.Command
			if err := json.Unmarshal(envelope.Payload, &command); err != nil {
				continue
			}
			err := c.updater.Install(ctx, c.httpClient(), command)
			success := errors.Is(err, agentupdate.ErrRestartRequired)
			result := map[string]any{"operation": "upgrade", "task_id": command.TaskID, "rollout_id": command.RolloutID, "success": success}
			if err != nil && !success {
				result["error"] = err.Error()
			}
			if writeErr := session.write("result", result); writeErr != nil {
				return established, writeErr
			}
			if success {
				return established, agentupdate.ErrRestartRequired
			}
		case "cmd.revoke":
			return established, errors.New("node credential revoked")
		}
	}
}

func (c *Client) heartbeats(done <-chan struct{}, session *wsSession, errs chan<- error) {
	ticker := time.NewTicker(agentHeartbeatInterval)
	defer ticker.Stop()
	send := func() error {
		addresses := localIPs()
		ruleRevision, ruleCount := c.detectionRuleStatus()
		return session.write("heartbeat", map[string]any{"version": c.version, "ip": firstLocalIP(addresses), "ips": addresses, "os": runtime.GOOS, "arch": runtime.GOARCH, "capabilities": agentCapabilities(c.cfg.TemplateRoot), "healthy": true, "pots": c.runtime.Statuses(), "decoys": c.decoys.Statuses(), "sense": c.sense.Status(), "queued_events": c.queue.Len(), "rule_revision": ruleRevision, "rule_count": ruleCount, "upgrade": c.updater.Status()})
	}
	if err := send(); err != nil {
		_ = session.conn.Close()
		select {
		case errs <- err:
		default:
		}
		return
	}
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := send(); err != nil {
				_ = session.conn.Close()
				select {
				case errs <- err:
				default:
				}
				return
			}
		}
	}
}

func (c *Client) detectionRuleStatus() (int64, int) {
	c.rulesMu.RLock()
	defer c.rulesMu.RUnlock()
	return c.ruleRevision, c.matcher.Count()
}

func agentCapabilities(templateRoots ...string) []string {
	codes := pots.SupportedCodesAt("")
	if len(templateRoots) > 0 {
		codes = pots.SupportedCodesAt(templateRoots[0])
	}
	capabilities := make([]string, 0, len(codes)+5)
	for _, code := range codes {
		capabilities = append(capabilities, "pot."+code)
	}
	capabilities = append(capabilities, "decoy.file", "decoy.credential", "decoy.network.passive", "network.ipv6")
	if runtime.GOOS == "linux" {
		capabilities = append(capabilities, "sense.passive")
	}
	return capabilities
}
func (s *wsSession) write(messageType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	envelope := wsEnvelope{ID: uuid.NewString(), Type: messageType, TS: time.Now().Unix(), Payload: data}
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return s.conn.WriteJSON(envelope)
}
func localIPs() []string {
	interfaces, _ := net.Interfaces()
	var ipv4, ipv6 []string
	seen := map[string]bool{}
	for _, item := range interfaces {
		if item.Flags&net.FlagUp == 0 || item.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := item.Addrs()
		for _, address := range addrs {
			var ip net.IP
			switch value := address.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			value := ip.String()
			if seen[value] {
				continue
			}
			seen[value] = true
			if ip.To4() != nil {
				ipv4 = append(ipv4, value)
			} else {
				ipv6 = append(ipv6, value)
			}
		}
	}
	return append(ipv4, ipv6...)
}

func firstLocalIP(addresses []string) string {
	if len(addresses) == 0 {
		return ""
	}
	return addresses[0]
}
func hostname() string {
	value, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(value)
}
