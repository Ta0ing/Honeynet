//go:build !windows

package agentupdate

import "os"

func replaceExecutable(staged, current, _ string, _ string) error {
	return os.Rename(staged, current)
}

func restoreExecutable(backup, current, _ string, _ string) error {
	temporary := current + ".rollback"
	if err := copyExecutable(backup, temporary); err != nil {
		return err
	}
	return os.Rename(temporary, current)
}

func RunUpdateHelper(_, _, _, _ string, _ int) error {
	return nil
}
