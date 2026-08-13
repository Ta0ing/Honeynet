package pots

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxModbusFrame = 260

type ModbusService struct {
	listener net.Listener
	once     sync.Once
}

func (s *ModbusService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *ModbusService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *ModbusService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		transactionID, unitID, pdu, err := readModbusFrame(conn)
		if err != nil {
			return
		}
		function := pdu[0]
		payload := modbusEventPayload(unitID, pdu)
		sink(protocol.NewEvent("modbus.request", src, dst, payload, "ics", "session"))
		if function == 5 || function == 6 || function == 15 || function == 16 {
			sink(protocol.NewEvent("modbus.write", src, dst, payload, "ics", "write"))
		}
		response := modbusResponsePDU(pdu)
		if err := writeModbusFrame(conn, transactionID, unitID, response); err != nil {
			return
		}
	}
}

func readModbusFrame(reader io.Reader) (uint16, byte, []byte, error) {
	var header [7]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return 0, 0, nil, err
	}
	if binary.BigEndian.Uint16(header[2:4]) != 0 {
		return 0, 0, nil, errors.New("invalid Modbus protocol identifier")
	}
	length := int(binary.BigEndian.Uint16(header[4:6]))
	if length < 2 || length > maxModbusFrame {
		return 0, 0, nil, errors.New("invalid Modbus frame length")
	}
	pdu := make([]byte, length-1)
	if _, err := io.ReadFull(reader, pdu); err != nil {
		return 0, 0, nil, err
	}
	return binary.BigEndian.Uint16(header[0:2]), header[6], pdu, nil
}

func writeModbusFrame(writer io.Writer, transactionID uint16, unitID byte, pdu []byte) error {
	header := make([]byte, 7)
	binary.BigEndian.PutUint16(header[0:2], transactionID)
	binary.BigEndian.PutUint16(header[4:6], uint16(len(pdu)+1))
	header[6] = unitID
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(pdu)
	return err
}

func modbusEventPayload(unitID byte, pdu []byte) map[string]any {
	payload := map[string]any{"unit_id": unitID, "function": pdu[0], "data": hex.EncodeToString(pdu[1:])}
	if len(pdu) >= 5 {
		payload["address"] = binary.BigEndian.Uint16(pdu[1:3])
		if pdu[0] == 5 || pdu[0] == 6 {
			payload["value"] = binary.BigEndian.Uint16(pdu[3:5])
		} else {
			payload["quantity"] = binary.BigEndian.Uint16(pdu[3:5])
		}
	}
	return payload
}

func modbusResponsePDU(request []byte) []byte {
	function := request[0]
	if len(request) < 5 {
		return []byte{function | 0x80, 0x03}
	}
	quantity := int(binary.BigEndian.Uint16(request[3:5]))
	switch function {
	case 1, 2:
		if quantity < 1 || quantity > 2000 {
			return []byte{function | 0x80, 0x03}
		}
		byteCount := (quantity + 7) / 8
		return append([]byte{function, byte(byteCount)}, make([]byte, byteCount)...)
	case 3, 4:
		if quantity < 1 || quantity > 125 {
			return []byte{function | 0x80, 0x03}
		}
		return append([]byte{function, byte(quantity * 2)}, make([]byte, quantity*2)...)
	case 5, 6:
		return append([]byte(nil), request[:5]...)
	case 15, 16:
		return append([]byte{function}, request[1:5]...)
	default:
		return []byte{function | 0x80, 0x01}
	}
}
