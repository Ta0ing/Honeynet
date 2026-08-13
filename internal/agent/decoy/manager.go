package decoy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
	"github.com/honeynet/honeynet/internal/decoyconfig"
)

type Sink func(protocol.Event)

type fileMonitor interface {
	Stop() error
}

type ownedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type manifest struct {
	Files map[string]ownedFile `json:"files"`
}

type instance struct {
	target      protocol.DecoyTarget
	monitor     fileMonitor
	status      string
	managedPath string
	lastError   string
	hitCount    atomic.Int64
	lastHitUnix atomic.Int64
}

type Manager struct {
	mu           sync.Mutex
	manifestPath string
	instances    map[string]*instance
	owned        map[string]ownedFile
	sink         Sink
}

func New(stateDir string, sink Sink) (*Manager, error) {
	manager := &Manager{
		manifestPath: filepath.Join(stateDir, "decoys", "manifest.json"),
		instances:    map[string]*instance{}, owned: map[string]ownedFile{}, sink: sink,
	}
	if err := manager.loadManifest(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) Apply(ctx context.Context, targets []protocol.DecoyTarget) []protocol.DecoyResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	desired := make(map[string]protocol.DecoyTarget, len(targets))
	cleanupErrors := map[string]error{}
	for _, target := range targets {
		desired[target.ID] = target
	}
	for id, current := range m.instances {
		target, exists := desired[id]
		if !exists || target.Status != "enabled" || !sameTarget(current.target, target) {
			if current.monitor != nil {
				_ = current.monitor.Stop()
			}
			delete(m.instances, id)
			if !exists || target.Status != "enabled" || current.managedPath != targetPath(target) {
				if err := m.cleanupOwnedLocked(id); err != nil {
					cleanupErrors[id] = err
				}
			}
		}
	}
	for id := range m.owned {
		target, exists := desired[id]
		if !exists || target.Status != "enabled" {
			if err := m.cleanupOwnedLocked(id); err != nil {
				cleanupErrors[id] = err
			}
		}
	}
	results := make([]protocol.DecoyResult, 0, len(targets))
	for _, target := range targets {
		if target.ID == "" || target.Name == "" {
			results = append(results, protocol.DecoyResult{DecoyID: target.ID, Status: "error", Success: false, LastError: "decoy id and name are required"})
			continue
		}
		if target.Status != "enabled" {
			if err := cleanupErrors[target.ID]; err != nil {
				results = append(results, protocol.DecoyResult{DecoyID: target.ID, Status: "error", Success: false, LastError: "remove managed decoy: " + err.Error()})
				continue
			}
			results = append(results, protocol.DecoyResult{DecoyID: target.ID, Status: "stopped", Success: true})
			continue
		}
		if current := m.instances[target.ID]; current != nil {
			results = append(results, current.result(true))
			continue
		}
		current, err := m.startLocked(ctx, target)
		if err != nil {
			results = append(results, protocol.DecoyResult{DecoyID: target.ID, Status: "error", Success: false, LastError: err.Error()})
			continue
		}
		m.instances[target.ID] = current
		results = append(results, current.result(true))
	}
	return results
}

func (m *Manager) startLocked(_ context.Context, target protocol.DecoyTarget) (*instance, error) {
	spec, err := decoyconfig.ParseMap(target.Type, target.Config)
	if err != nil {
		return nil, err
	}
	current := &instance{target: target}
	if target.Type == "network" {
		current.status = "passive"
		return current, nil
	}
	if !filepath.IsAbs(spec.Path) || filepath.Clean(spec.Path) == filepath.VolumeName(spec.Path)+string(filepath.Separator) {
		return nil, errors.New("decoy path is not an absolute non-root path on this Agent platform")
	}
	spec.Path = filepath.Clean(spec.Path)
	current.managedPath = spec.Path
	content := decoyconfig.Render(target.Type, target.ID, target.Name, spec)
	mode, _ := decoyconfig.FileMode(spec)
	if err := m.ensureFileLocked(target.ID, spec, content, os.FileMode(mode)); err != nil {
		return nil, err
	}
	monitor, err := newFileMonitor(spec.Path, func(action string) {
		now := time.Now()
		current.hitCount.Add(1)
		current.lastHitUnix.Store(now.UnixNano())
		event := protocol.NewEvent("decoy."+target.Type, protocol.Endpoint{}, protocol.Endpoint{}, map[string]any{
			"decoy_id": target.ID, "decoy_name": target.Name, "decoy_type": target.Type,
			"path": spec.Path, "action": action,
		}, "decoy", "local")
		event.DecoyID = target.ID
		event.Service = "decoy"
		m.sink(event)
	})
	if err != nil {
		_ = m.cleanupOwnedLocked(target.ID)
		return nil, fmt.Errorf("monitor decoy file: %w", err)
	}
	current.monitor = monitor
	current.status = "monitoring"
	return current, nil
}

func (m *Manager) ensureFileLocked(id string, spec decoyconfig.Spec, content []byte, mode os.FileMode) error {
	info, err := os.Lstat(spec.Path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return errors.New("decoy path exists but is not a regular file")
		}
		owned, known := m.owned[id]
		if known && owned.Path == spec.Path {
			current, readErr := os.ReadFile(spec.Path)
			if readErr != nil {
				return readErr
			}
			if digest(current) != owned.SHA256 {
				return errors.New("Agent-owned decoy was modified; preserving it as evidence and refusing to overwrite it")
			}
			if digest(content) == owned.SHA256 {
				return nil
			}
			if err := replaceOwnedFile(spec.Path, content, mode); err != nil {
				return err
			}
			m.owned[id] = ownedFile{Path: spec.Path, SHA256: digest(content)}
			return m.saveManifestLocked()
		}
		if !spec.MonitorExisting {
			return errors.New("decoy path already exists; enable monitor_existing to watch it without modifying or deleting it")
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	parent := filepath.Dir(spec.Path)
	if parentInfo, statErr := os.Stat(parent); statErr != nil {
		if !errors.Is(statErr, os.ErrNotExist) || !spec.CreateParent {
			return errors.New("decoy parent directory does not exist; enable create_parent to create it")
		}
		if err := os.MkdirAll(parent, 0750); err != nil {
			return fmt.Errorf("create decoy parent: %w", err)
		}
	} else if !parentInfo.IsDir() {
		return errors.New("decoy parent path is not a directory")
	}
	file, err := os.OpenFile(spec.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("create decoy file: %w", err)
	}
	writeErr := error(nil)
	if _, err = file.Write(content); err != nil {
		writeErr = err
	} else if err = file.Sync(); err != nil {
		writeErr = err
	}
	if closeErr := file.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		_ = os.Remove(spec.Path)
		return fmt.Errorf("write decoy file: %w", writeErr)
	}
	if err := os.Chmod(spec.Path, mode); err != nil {
		_ = os.Remove(spec.Path)
		return err
	}
	m.owned[id] = ownedFile{Path: spec.Path, SHA256: digest(content)}
	if err := m.saveManifestLocked(); err != nil {
		delete(m.owned, id)
		_ = os.Remove(spec.Path)
		return fmt.Errorf("persist decoy ownership: %w", err)
	}
	return nil
}

func (m *Manager) cleanupOwnedLocked(id string) error {
	record, exists := m.owned[id]
	if !exists {
		return nil
	}
	content, err := os.ReadFile(record.Path)
	if err == nil && digest(content) == record.SHA256 {
		if err := os.Remove(record.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	delete(m.owned, id)
	return m.saveManifestLocked()
}

func (m *Manager) Statuses() []protocol.DecoyStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	statuses := make([]protocol.DecoyStatus, 0, len(m.instances))
	for _, current := range m.instances {
		status := protocol.DecoyStatus{ID: current.target.ID, Status: current.status, ManagedPath: current.managedPath, LastError: current.lastError, HitCount: current.hitCount.Load()}
		if value := current.lastHitUnix.Load(); value > 0 {
			lastHit := time.Unix(0, value)
			status.LastHitAt = &lastHit
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, current := range m.instances {
		if current.monitor != nil {
			_ = current.monitor.Stop()
		}
		delete(m.instances, id)
	}
}

func (m *Manager) loadManifest() error {
	raw, err := os.ReadFile(m.manifestPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var state manifest
	if err := json.Unmarshal(raw, &state); err != nil {
		quarantinePath := fmt.Sprintf("%s.corrupt-%d", m.manifestPath, time.Now().UTC().UnixNano())
		if renameErr := os.Rename(m.manifestPath, quarantinePath); renameErr != nil {
			return fmt.Errorf("quarantine invalid decoy manifest after parse error %v: %w", err, renameErr)
		}
		// An unreadable ownership record must never be guessed. Quarantine it and
		// continue without ownership so existing evidence cannot be deleted.
		return nil
	}
	if state.Files != nil {
		m.owned = state.Files
	}
	return nil
}

func (m *Manager) saveManifestLocked() error {
	if err := os.MkdirAll(filepath.Dir(m.manifestPath), 0700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(manifest{Files: m.owned}, "", "  ")
	if err != nil {
		return err
	}
	temporary := m.manifestPath + ".tmp"
	if err := os.WriteFile(temporary, raw, 0600); err != nil {
		return err
	}
	if err := os.Rename(temporary, m.manifestPath); err != nil {
		return err
	}
	return os.Chmod(m.manifestPath, 0600)
}

func (current *instance) result(success bool) protocol.DecoyResult {
	result := protocol.DecoyResult{DecoyID: current.target.ID, Status: current.status, Success: success, ManagedPath: current.managedPath, LastError: current.lastError, HitCount: current.hitCount.Load()}
	if value := current.lastHitUnix.Load(); value > 0 {
		lastHit := time.Unix(0, value)
		result.LastHitAt = &lastHit
	}
	return result
}

func sameTarget(left, right protocol.DecoyTarget) bool {
	a, _ := json.Marshal(left)
	b, _ := json.Marshal(right)
	return string(a) == string(b)
}

func targetPath(target protocol.DecoyTarget) string {
	spec, err := decoyconfig.ParseMap(target.Type, target.Config)
	if err != nil {
		return ""
	}
	return spec.Path
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func replaceOwnedFile(path string, content []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".honeynet-decoy-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
