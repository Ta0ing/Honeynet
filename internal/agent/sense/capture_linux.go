//go:build linux

package sense

import (
	"context"
	"errors"
	"net"
	"sync"

	"golang.org/x/sys/unix"
)

const ethernetAll = 0x0003

type linuxCapture struct {
	fd   int
	once sync.Once
}

func openCapture(interfaceName string) (packetCapture, error) {
	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW|unix.SOCK_CLOEXEC, int(htons(ethernetAll)))
	if err != nil {
		return nil, err
	}
	capture := &linuxCapture{fd: fd}
	_ = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, 4<<20)
	_ = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &unix.Timeval{Sec: 1})
	if interfaceName != "" {
		device, err := net.InterfaceByName(interfaceName)
		if err != nil {
			_ = capture.Close()
			return nil, err
		}
		if err := unix.Bind(fd, &unix.SockaddrLinklayer{Protocol: htons(ethernetAll), Ifindex: device.Index}); err != nil {
			_ = capture.Close()
			return nil, err
		}
	}
	return capture, nil
}

func (c *linuxCapture) Run(ctx context.Context, observe func(Probe)) error {
	buffer := make([]byte, 65536)
	for {
		if ctx.Err() != nil {
			return nil
		}
		read, address, err := unix.Recvfrom(c.fd, buffer, 0)
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EINTR) {
				continue
			}
			if ctx.Err() != nil || errors.Is(err, unix.EBADF) {
				return nil
			}
			return err
		}
		if link, ok := address.(*unix.SockaddrLinklayer); ok && link.Pkttype == unix.PACKET_OUTGOING {
			continue
		}
		if probe, ok := DecodePacket(buffer[:read]); ok {
			observe(probe)
		}
	}
}

func (c *linuxCapture) Close() error {
	var err error
	c.once.Do(func() { err = unix.Close(c.fd) })
	return err
}

func htons(value uint16) uint16 { return value<<8 | value>>8 }
