package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPendingUpdateRollbackRequiresSameAttemptAtGuardStart(t *testing.T) {
	dir := t.TempDir()
	agent := filepath.Join(dir, "bin", "honeynet-agent")
	statePath := filepath.Join(dir, "custom-state", "update-state.json")
	if err := os.MkdirAll(filepath.Dir(agent), 0700); err != nil {
		t.Fatal(err)
	}
	state := updateState{PreviousVersion: "1.0.0", TargetVersion: "2.0.0", ExecutablePath: agent, BackupPath: agent + ".old", Status: "awaiting_health"}
	if err := saveState(statePath, state); err != nil {
		t.Fatal(err)
	}
	if !rollbackCandidate(state, agent) {
		t.Fatal("unhealthy updated target should roll back")
	}
	if !pendingUpdateStillUnhealthy(state, state, agent, nil) {
		t.Fatal("same pending attempt should roll back regardless of child exit status")
	}
	healthy := state
	healthy.Status = "healthy"
	if err := saveState(statePath, healthy); err != nil {
		t.Fatal(err)
	}
	if pendingUpdateStillUnhealthy(state, healthy, agent, nil) {
		t.Fatal("healthy target crash must not roll back")
	}
	if pendingUpdateStillUnhealthy(updateState{}, state, agent, nil) {
		t.Fatal("ordinary first install without update state must not roll back")
	}
	next := state
	next.TargetVersion = "3.0.0"
	if pendingUpdateStillUnhealthy(state, next, agent, nil) {
		t.Fatal("a different pending update must not be rolled back with stale metadata")
	}
	if !pendingUpdateStillUnhealthy(state, updateState{}, agent, errors.New("corrupt state")) {
		t.Fatal("missing health confirmation for the same startup must fail closed")
	}
}

func TestRestoreUsesAtomicSiblingReplacement(t *testing.T) {
	dir := t.TempDir()
	backup, target := filepath.Join(dir, "backup"), filepath.Join(dir, "honeynet-agent")
	if err := os.WriteFile(backup, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("bad-new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := restore(backup, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "old" {
		t.Fatalf("restored=%q err=%v", data, err)
	}
}
