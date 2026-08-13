// agent-guard is the stable Linux A/B launcher for Honeynet Agent. It is not
// replaced by an Agent rollout, so an invalid, non-executable or immediately
// crashing signed target cannot trap systemd in a restart loop. The guard
// restores the updater's verified local backup and records an auditable state
// before returning control to systemd.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type updateState struct {
	PreviousVersion string    `json:"previous_version"`
	TargetVersion   string    `json:"target_version"`
	ExecutablePath  string    `json:"executable_path"`
	BackupPath      string    `json:"backup_path"`
	Status          string    `json:"status"`
	BootAttempts    int       `json:"boot_attempts"`
	HealthDeadline  time.Time `json:"health_deadline"`
	LastError       string    `json:"last_error,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fatal(errors.New("usage: honeynet-agent-guard AGENT [args...]"))
	}
	agent, err := filepath.Abs(os.Args[1])
	if err != nil {
		fatal(err)
	}
	stateDir := agentStateDir(os.Args[2:])
	statePath := filepath.Join(stateDir, "update-state.json")
	initial, initialErr := loadState(statePath)
	state := initial
	pendingAtStart := initialErr == nil && rollbackCandidate(initial, agent)
	command := exec.Command(agent, os.Args[2:]...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	command.Env = os.Environ()
	err = command.Start()
	if err == nil {
		done := make(chan error, 1)
		go func() { done <- command.Wait() }()
		if pendingAtStart && !initial.HealthDeadline.IsZero() {
			delay := time.Until(initial.HealthDeadline)
			if delay < 0 {
				delay = 0
			}
			timer := time.NewTimer(delay)
			select {
			case err = <-done:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			case <-timer.C:
				fresh, loadErr := loadState(statePath)
				if pendingUpdateStillUnhealthy(initial, fresh, agent, loadErr) {
					if loadErr == nil {
						state = fresh
					}
					_ = command.Process.Kill()
					err = fmt.Errorf("updated Agent missed health deadline")
					<-done
				} else {
					err = <-done
				}
			}
		} else {
			err = <-done
		}
	}
	fresh, freshErr := loadState(statePath)
	if freshErr == nil {
		state = fresh
	}
	// An old Agent writes awaiting_health only while this guard invocation is
	// already running. That invocation must be allowed to exit so systemd can
	// start the newly installed target. Rollback is therefore legal only when
	// this guard invocation itself started for the same pending target.
	if pendingAtStart && pendingUpdateStillUnhealthy(initial, fresh, agent, freshErr) {
		if freshErr != nil {
			state = initial
		}
		state.Status = "rolling_back"
		state.LastError = fmt.Sprintf("updated Agent exited before health confirmation: %v", err)
		_ = saveState(statePath, state)
		if rollbackErr := restore(state.BackupPath, agent); rollbackErr != nil {
			state.Status = "rollback_failed"
			state.LastError = rollbackErr.Error()
			_ = saveState(statePath, state)
			fatal(fmt.Errorf("automatic Agent rollback failed: %w", rollbackErr))
		}
		state.Status = "rolled_back"
		_ = saveState(statePath, state)
		// Exit non-zero. systemd Restart=always starts the restored previous
		// binary, which reports rolled_back to Server on hello.ack.
		os.Exit(75)
	}
	if err == nil {
		return
	}
	if exit := new(exec.ExitError); errors.As(err, &exit) {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				os.Exit(128 + int(status.Signal()))
			}
			os.Exit(status.ExitStatus())
		}
	}
	fatal(err)
}

func agentStateDir(args []string) string {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == "--state-dir" {
			return args[index+1]
		}
	}
	if value := os.Getenv("HONEYPOT_AGENT_STATE_DIR"); value != "" {
		return value
	}
	return "/var/lib/honeynet-agent"
}

func rollbackCandidate(state updateState, agent string) bool {
	if state.Status != "awaiting_health" {
		return false
	}
	if current, err := filepath.Abs(state.ExecutablePath); err != nil || current != agent {
		return false
	}
	if state.BackupPath == "" || state.PreviousVersion == "" || state.TargetVersion == "" {
		return false
	}
	return true
}

func pendingUpdateStillUnhealthy(initial, fresh updateState, agent string, freshErr error) bool {
	if !rollbackCandidate(initial, agent) {
		return false
	}
	// Losing or corrupting the health record cannot be treated as confirmation.
	if freshErr != nil {
		return true
	}
	return fresh.Status == "awaiting_health" &&
		fresh.ExecutablePath == initial.ExecutablePath &&
		fresh.BackupPath == initial.BackupPath &&
		fresh.PreviousVersion == initial.PreviousVersion &&
		fresh.TargetVersion == initial.TargetVersion
}

func loadState(path string) (updateState, error) {
	var state updateState
	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	return state, json.Unmarshal(data, &state)
}

func saveState(path string, state updateState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".guard.tmp"
	if err := os.WriteFile(temporary, data, 0600); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func restore(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	temporary := target + ".guard.rollback"
	output, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	if _, err = output.ReadFrom(input); err == nil {
		err = output.Sync()
	}
	if closeErr := output.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(temporary)
		return err
	}
	if err := os.Chmod(temporary, 0755); err != nil {
		return err
	}
	return os.Rename(temporary, target)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
