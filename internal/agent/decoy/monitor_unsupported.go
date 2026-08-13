//go:build !linux

package decoy

import (
	"crypto/sha256"
	"errors"
	"os"
	"sync"
	"time"
)

type pollingFileMonitor struct {
	stop chan struct{}
	once sync.Once
	wg   sync.WaitGroup
}

func newFileMonitor(path string, callback func(string)) (fileMonitor, error) {
	baseline, err := fileSnapshot(path)
	if err != nil {
		return nil, err
	}
	monitor := &pollingFileMonitor{stop: make(chan struct{})}
	monitor.wg.Add(1)
	go func() {
		defer monitor.wg.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		current := baseline
		for {
			select {
			case <-monitor.stop:
				return
			case <-ticker.C:
				next, snapshotErr := fileSnapshot(path)
				if snapshotErr != nil {
					if !current.missing && errors.Is(snapshotErr, os.ErrNotExist) {
						callback("deleted")
						current = snapshot{missing: true}
					}
					continue
				}
				if current.missing {
					callback("recreated")
				} else if next.digest != current.digest {
					callback("modified")
				}
				current = next
			}
		}
	}()
	return monitor, nil
}

func (m *pollingFileMonitor) Stop() error {
	m.once.Do(func() {
		close(m.stop)
		m.wg.Wait()
	})
	return nil
}

type snapshot struct {
	digest  [sha256.Size]byte
	missing bool
}

func fileSnapshot(path string) (snapshot, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return snapshot{}, err
	}
	return snapshot{digest: sha256.Sum256(content)}, nil
}
