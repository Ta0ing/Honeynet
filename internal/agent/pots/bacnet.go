package pots

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"sync"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxBACnetPacket = 64 << 10

type BACnetService struct {
	conn net.PacketConn
	once sync.Once
}

func (s *BACnetService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	conn, err := net.ListenPacket("udp", listenAddress(target))
	if err != nil {
		return err
	}
	s.conn = conn
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	go s.serve(conn, target, sink)
	return nil
}

func (s *BACnetService) Stop() error {
	if s.conn == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.conn.Close() })
	return err
}

func (s *BACnetService) serve(conn net.PacketConn, target protocol.PotTarget, sink Sink) {
	buffer := make([]byte, maxBACnetPacket)
	for {
		count, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		request, err := parseBACnetPacket(buffer[:count])
		if err != nil {
			continue
		}
		src, dst := endpoint(remote), endpoint(conn.LocalAddr())
		payload := map[string]any{
			"bvlc_function": request.bvlcFunction, "service": bacnetServiceName(request.pduType, request.serviceChoice),
			"pdu_type": request.pduType, "service_choice": request.serviceChoice, "invoke_id": request.invokeID,
		}
		if request.hasObject {
			payload["object_type"] = request.objectType
			payload["object_instance"] = request.objectInstance
			payload["property_id"] = request.propertyID
		}
		sink(protocol.NewEvent("bacnet.request", src, dst, payload, "ics", "iot"))
		if request.pduType == 1 && request.serviceChoice == 8 {
			deviceID := configInt(target.Config, "device_id", 12345, 0, 4194303)
			vendorID := configInt(target.Config, "vendor_id", 1200, 0, 65535)
			_, _ = conn.WriteTo(bacnetIAm(uint32(deviceID), uint16(vendorID)), remote)
			sink(protocol.NewEvent("bacnet.discovery", src, dst, map[string]any{
				"request": "who_is", "device_id": deviceID, "vendor_id": vendorID,
			}, "ics", "recon"))
		} else if request.pduType == 0 {
			if request.serviceChoice == 12 {
				sink(protocol.NewEvent("bacnet.read_property", src, dst, payload, "ics", "read"))
			} else if request.serviceChoice == 15 {
				sink(protocol.NewEvent("bacnet.write_property", src, dst, payload, "ics", "write"))
			}
			_, _ = conn.WriteTo(bacnetError(request.invokeID, request.serviceChoice), remote)
		}
	}
}

type bacnetRequest struct {
	bvlcFunction   byte
	pduType        byte
	serviceChoice  byte
	invokeID       byte
	hasObject      bool
	objectType     uint16
	objectInstance uint32
	propertyID     uint32
}

func parseBACnetPacket(packet []byte) (bacnetRequest, error) {
	if len(packet) < 8 || packet[0] != 0x81 || int(binary.BigEndian.Uint16(packet[2:4])) != len(packet) {
		return bacnetRequest{}, errors.New("invalid BACnet/IP BVLC header")
	}
	result := bacnetRequest{bvlcFunction: packet[1]}
	if packet[1] != 0x0a && packet[1] != 0x0b {
		return result, errors.New("unsupported BACnet/IP BVLC function")
	}
	npdu := packet[4:]
	if len(npdu) < 3 || npdu[0] != 1 {
		return result, errors.New("invalid BACnet NPDU")
	}
	control, offset := npdu[1], 2
	if control&0x20 != 0 {
		if offset+3 > len(npdu) {
			return result, errors.New("invalid BACnet destination address")
		}
		destinationLength := int(npdu[offset+2])
		offset += 3 + destinationLength
		if offset >= len(npdu) {
			return result, errors.New("invalid BACnet destination hop count")
		}
		offset++
	}
	if control&0x08 != 0 {
		if offset+3 > len(npdu) {
			return result, errors.New("invalid BACnet source address")
		}
		sourceLength := int(npdu[offset+2])
		offset += 3 + sourceLength
	}
	if control&0x80 != 0 || offset+2 > len(npdu) {
		return result, errors.New("unsupported BACnet network message")
	}
	apdu := npdu[offset:]
	result.pduType = apdu[0] >> 4
	switch result.pduType {
	case 0:
		if len(apdu) < 4 {
			return result, errors.New("invalid BACnet confirmed request")
		}
		result.invokeID, result.serviceChoice = apdu[2], apdu[3]
		if result.serviceChoice == 12 || result.serviceChoice == 15 {
			result.objectType, result.objectInstance, result.propertyID, result.hasObject = parseBACnetPropertyReference(apdu[4:])
		}
	case 1:
		result.serviceChoice = apdu[1]
	default:
		return result, errors.New("unsupported BACnet APDU")
	}
	return result, nil
}

func parseBACnetPropertyReference(payload []byte) (uint16, uint32, uint32, bool) {
	if len(payload) < 7 || payload[0] != 0x0c {
		return 0, 0, 0, false
	}
	objectID := binary.BigEndian.Uint32(payload[1:5])
	header := payload[5]
	if header>>4 != 1 || header&0x08 == 0 {
		return 0, 0, 0, false
	}
	length, offset := int(header&0x07), 6
	if length == 5 {
		if offset >= len(payload) {
			return 0, 0, 0, false
		}
		length = int(payload[offset])
		offset++
	}
	if length < 1 || length > 4 || offset+length > len(payload) {
		return 0, 0, 0, false
	}
	propertyID := uint32(0)
	for _, value := range payload[offset : offset+length] {
		propertyID = propertyID<<8 | uint32(value)
	}
	return uint16(objectID >> 22), objectID & 0x3fffff, propertyID, true
}

func bacnetIAm(deviceID uint32, vendorID uint16) []byte {
	objectID := uint32(8)<<22 | deviceID&0x3fffff
	apdu := []byte{0x10, 0x00, 0xc4, 0, 0, 0, 0, 0x22, 0x05, 0xc4, 0x91, 0x03, 0x22, byte(vendorID >> 8), byte(vendorID)}
	binary.BigEndian.PutUint32(apdu[3:7], objectID)
	return bacnetBVLC(0x0a, append([]byte{0x01, 0x00}, apdu...))
}

func bacnetError(invokeID, serviceChoice byte) []byte {
	apdu := []byte{0x50, invokeID, serviceChoice, 0x91, 0x02, 0x91, 0x20}
	return bacnetBVLC(0x0a, append([]byte{0x01, 0x00}, apdu...))
}

func bacnetBVLC(function byte, npdu []byte) []byte {
	packet := make([]byte, 4+len(npdu))
	packet[0], packet[1] = 0x81, function
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))
	copy(packet[4:], npdu)
	return packet
}

func bacnetServiceName(pduType, choice byte) string {
	if pduType == 1 {
		switch choice {
		case 0:
			return "i_am"
		case 7:
			return "who_has"
		case 8:
			return "who_is"
		}
	}
	if pduType == 0 {
		switch choice {
		case 12:
			return "read_property"
		case 15:
			return "write_property"
		case 14:
			return "read_property_multiple"
		}
	}
	return "unknown"
}

func configInt(config map[string]any, key string, fallback, minimum, maximum int) int {
	value, ok := config[key].(float64)
	if !ok || value < float64(minimum) || value > float64(maximum) {
		return fallback
	}
	return int(value)
}
