//go:build linux

package decoy

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type linuxFileMonitor struct {
	fd   int
	once sync.Once
	stop chan struct{}
	wg   sync.WaitGroup
}

func newFileMonitor(path string, callback func(string)) (fileMonitor, error) {
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}
	mask := uint32(unix.IN_OPEN | unix.IN_ACCESS | unix.IN_MODIFY | unix.IN_ATTRIB | unix.IN_CLOSE_WRITE | unix.IN_DELETE_SELF | unix.IN_MOVE_SELF)
	if _, err := unix.InotifyAddWatch(fd, path, mask); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	monitor := &linuxFileMonitor{fd: fd, stop: make(chan struct{})}
	monitor.wg.Add(1)
	go monitor.run(callback)
	return monitor, nil
}

func (m *linuxFileMonitor) run(callback func(string)) {
	defer m.wg.Done()
	buffer := make([]byte, 16<<10)
	var lastEvent time.Time
	for {
		select {
		case <-m.stop:
			return
		default:
		}
		count, err := unix.Read(m.fd, buffer)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if errors.Is(err, unix.EAGAIN) {
				select {
				case <-m.stop:
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}
			return
		}
		for offset := 0; offset+unix.SizeofInotifyEvent <= count; {
			event := (*unix.InotifyEvent)(unsafe.Pointer(&buffer[offset]))
			action := inotifyAction(event.Mask)
			if action != "" && time.Since(lastEvent) >= time.Second {
				lastEvent = time.Now()
				callback(action)
			}
			offset += unix.SizeofInotifyEvent + int(event.Len)
		}
	}
}

func (m *linuxFileMonitor) Stop() error {
	var err error
	m.once.Do(func() {
		close(m.stop)
		m.wg.Wait()
		err = unix.Close(m.fd)
	})
	return err
}

func inotifyAction(mask uint32) string {
	switch {
	case mask&unix.IN_DELETE_SELF != 0:
		return "deleted"
	case mask&unix.IN_MOVE_SELF != 0:
		return "moved"
	case mask&unix.IN_MODIFY != 0 || mask&unix.IN_CLOSE_WRITE != 0:
		return "modified"
	case mask&unix.IN_ATTRIB != 0:
		return "metadata_changed"
	case mask&unix.IN_OPEN != 0:
		return "opened"
	case mask&unix.IN_ACCESS != 0:
		return "read"
	default:
		return ""
	}
}
