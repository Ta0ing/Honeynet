package agentupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	ErrRestartRequired  = errors.New("Agent restart required to activate update")
	ErrRollbackRequired = errors.New("Agent update health check failed; rollback activated")
)

type Command struct {
	TaskID    string `json:"task_id"`
	RolloutID string `json:"rollout_id"`
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	URL       string `json:"url"`
	SHA256    string `json:"sha256"`
	Signature string `json:"signature"`
	Size      int64  `json:"size"`
}

type State struct {
	TaskID          string    `json:"task_id"`
	RolloutID       string    `json:"rollout_id"`
	PreviousVersion string    `json:"previous_version"`
	TargetVersion   string    `json:"target_version"`
	ExecutablePath  string    `json:"executable_path"`
	BackupPath      string    `json:"backup_path"`
	Status          string    `json:"status"`
	BootAttempts    int       `json:"boot_attempts"`
	InstalledAt     time.Time `json:"installed_at"`
	HealthDeadline  time.Time `json:"health_deadline"`
	LastError       string    `json:"last_error,omitempty"`
}

type Manager struct {
	StateDir        string
	CurrentVersion  string
	PublicKey       string
	DownloadBaseURL string
	ExecutablePath  string
	ServiceName     string
}

func (m *Manager) CheckStartup() error {
	state, err := m.loadState()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if state.Status == "rolling_back" && m.CurrentVersion == state.PreviousVersion {
		state.Status = "rolled_back"
		return m.saveState(state)
	}
	if state.Status != "awaiting_health" {
		return nil
	}
	if m.CurrentVersion != state.TargetVersion {
		state.LastError = fmt.Sprintf("started version %s, expected %s", m.CurrentVersion, state.TargetVersion)
		if err := m.rollback(&state); err != nil {
			return err
		}
		return ErrRollbackRequired
	}
	state.BootAttempts++
	if state.BootAttempts > 3 || time.Now().After(state.HealthDeadline) {
		state.LastError = "new Agent did not confirm a healthy control connection"
		if err := m.rollback(&state); err != nil {
			return err
		}
		return ErrRollbackRequired
	}
	return m.saveState(state)
}

// WaitForHealth enforces the signed update health deadline even when the new
// Agent process stays alive but never establishes its authenticated control
// connection. ConfirmHealthy changes the state after hello.ack, making this a
// no-op when startup succeeds normally.
func (m *Manager) WaitForHealth(ctx context.Context) error {
	state, err := m.loadState()
	if errors.Is(err, os.ErrNotExist) || err == nil && state.Status != "awaiting_health" {
		return nil
	}
	if err != nil {
		return err
	}
	delay := time.Until(state.HealthDeadline)
	if delay < 0 {
		delay = 0
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil
	case <-timer.C:
	}
	state, err = m.loadState()
	if err != nil || state.Status != "awaiting_health" {
		return err
	}
	state.LastError = "new Agent did not establish a healthy control connection before the deadline"
	if err := m.rollback(&state); err != nil {
		return err
	}
	return ErrRollbackRequired
}

func (m *Manager) Install(ctx context.Context, client *http.Client, command Command) error {
	if strings.TrimSpace(m.PublicKey) == "" {
		return errors.New("trusted update public key is not configured")
	}
	if command.OS != runtime.GOOS || command.Arch != runtime.GOARCH {
		return fmt.Errorf("update platform %s/%s does not match %s/%s", command.OS, command.Arch, runtime.GOOS, runtime.GOARCH)
	}
	if command.Version == m.CurrentVersion {
		return errors.New("target version is already running")
	}
	descriptor := Descriptor{Version: command.Version, OS: command.OS, Arch: command.Arch, SHA256: command.SHA256, Size: command.Size}
	if err := Verify(m.PublicKey, command.Signature, descriptor); err != nil {
		return err
	}
	downloadURL, err := m.resolveDownloadURL(command.URL)
	if err != nil {
		return err
	}
	parsed, err := url.Parse(downloadURL)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return errors.New("update URL must use HTTPS")
	}
	if command.Size <= 0 || command.Size > 256<<20 {
		return errors.New("update size is outside the 256 MiB limit")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download update: status %d", response.StatusCode)
	}
	updateDir := filepath.Join(m.StateDir, "updates")
	if err := os.MkdirAll(updateDir, 0700); err != nil {
		return err
	}
	downloadPath := filepath.Join(updateDir, "agent-"+safeVersion(command.Version)+".download")
	file, err := os.OpenFile(downloadPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(response.Body, command.Size+1))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written != command.Size || !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), command.SHA256) {
		return errors.New("downloaded Agent size or SHA-256 does not match signed manifest")
	}
	current, err := m.currentExecutable()
	if err != nil {
		return err
	}
	backupPath := filepath.Join(updateDir, "agent-"+safeVersion(m.CurrentVersion)+".rollback")
	if err := copyExecutable(current, backupPath); err != nil {
		return fmt.Errorf("backup current Agent: %w", err)
	}
	stagedPath := filepath.Join(filepath.Dir(current), ".honeynet-agent-"+safeVersion(command.Version)+".new")
	if err := copyExecutable(downloadPath, stagedPath); err != nil {
		return fmt.Errorf("stage Agent beside executable: %w", err)
	}
	state := State{
		TaskID: command.TaskID, RolloutID: command.RolloutID, PreviousVersion: m.CurrentVersion, TargetVersion: command.Version,
		ExecutablePath: current, BackupPath: backupPath, Status: "awaiting_health", InstalledAt: time.Now(), HealthDeadline: time.Now().Add(2 * time.Minute),
	}
	if err := m.saveState(state); err != nil {
		return err
	}
	if err := replaceExecutable(stagedPath, current, m.statePath(), m.ServiceName); err != nil {
		state.Status = "failed"
		state.LastError = err.Error()
		_ = m.saveState(state)
		return err
	}
	return ErrRestartRequired
}

func (m *Manager) resolveDownloadURL(value string) (string, error) {
	reference, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", errors.New("update URL is invalid")
	}
	if reference.IsAbs() {
		return reference.String(), nil
	}
	base, err := url.Parse(strings.TrimSpace(m.DownloadBaseURL))
	if err != nil || !base.IsAbs() {
		return "", errors.New("relative update URL requires the Agent gateway base URL")
	}
	return base.ResolveReference(reference).String(), nil
}

func (m *Manager) ConfirmHealthy() error {
	state, err := m.loadState()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if state.Status != "awaiting_health" || state.TargetVersion != m.CurrentVersion {
		return nil
	}
	state.Status = "healthy"
	state.LastError = ""
	if err := m.saveState(state); err != nil {
		return err
	}
	_ = os.Remove(state.BackupPath)
	return nil
}

func (m *Manager) Status() *State {
	state, err := m.loadState()
	if err != nil {
		return nil
	}
	return &state
}

func (m *Manager) rollback(state *State) error {
	state.Status = "rolling_back"
	if err := m.saveState(*state); err != nil {
		return err
	}
	if err := restoreExecutable(state.BackupPath, state.ExecutablePath, m.statePath(), m.ServiceName); err != nil {
		state.Status = "rollback_failed"
		state.LastError = err.Error()
		_ = m.saveState(*state)
		return fmt.Errorf("restore previous Agent: %w", err)
	}
	return nil
}

func (m *Manager) currentExecutable() (string, error) {
	if m.ExecutablePath != "" {
		return filepath.Abs(m.ExecutablePath)
	}
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Abs(path)
}

func (m *Manager) statePath() string {
	return filepath.Join(m.StateDir, "update-state.json")
}

func (m *Manager) loadState() (State, error) {
	data, err := os.ReadFile(m.statePath())
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func (m *Manager) saveState(state State) error {
	if err := os.MkdirAll(m.StateDir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(m.statePath(), data, 0600)
}

func copyExecutable(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	temporary := target + ".tmp"
	output, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Chmod(temporary, 0755); err != nil {
		return err
	}
	return os.Rename(temporary, target)
}

func safeVersion(version string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("._+-", r) {
			return r
		}
		return '-'
	}, version)
}
