//go:build !linux

package sense

func openCapture(string) (packetCapture, error) { return nil, ErrUnsupported }
