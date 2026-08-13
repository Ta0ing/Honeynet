//go:build windows

package agentupdate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func replaceExecutable(staged, current, statePath, serviceName string) error {
	return startWindowsHelper(staged, staged, current, statePath, serviceName)
}

func restoreExecutable(backup, current, statePath, serviceName string) error {
	return startWindowsHelper(backup, backup, current, statePath, serviceName)
}

func startWindowsHelper(helper, source, target, statePath, serviceName string) error {
	command := exec.Command(helper,
		"--update-helper", "--update-source", source, "--update-target", target,
		"--update-state", statePath, "--update-service", serviceName, "--update-parent", strconv.Itoa(os.Getpid()),
	)
	return command.Start()
}

func RunUpdateHelper(source, target, statePath, serviceName string, parentPID int) error {
	prepared := target + ".replace"
	_ = os.Remove(prepared)
	if err := copyExecutable(source, prepared); err != nil {
		return fmt.Errorf("prepare Windows Agent replacement: %w", err)
	}
	if serviceName != "" {
		_ = exec.Command("sc.exe", "stop", serviceName).Run()
	}
	_ = parentPID
	var replaceErr error
	for i := 0; i < 60; i++ {
		// Prepare the full replacement before touching the current executable.
		// If the old process still has the target locked, neither Remove nor
		// Rename succeeds and the existing executable remains intact.
		if removeErr := os.Remove(target); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			replaceErr = removeErr
		} else {
			replaceErr = os.Rename(prepared, target)
		}
		if replaceErr == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if replaceErr != nil {
		return fmt.Errorf("replace Windows Agent: %w", replaceErr)
	}
	if serviceName != "" {
		if err := exec.Command("sc.exe", "start", serviceName).Run(); err != nil {
			return err
		}
	}
	_ = statePath
	return nil
}
