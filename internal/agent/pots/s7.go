package pots

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxS7Frame = 4096

type S7Service struct {
	listener net.Listener
	once     sync.Once
}

func (s *S7Service) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *S7Service) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *S7Service) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	payload, err := readS7TPKT(conn)
	if err != nil {
		return
	}
	connection, err := parseCOTPConnection(payload)
	if err != nil {
		return
	}
	sink(protocol.NewEvent("s7.connection", src, dst, map[string]any{
		"source_tsap":      fmt.Sprintf("%04x", connection.sourceTSAP),
		"destination_tsap": fmt.Sprintf("%04x", connection.destinationTSAP),
		"source_reference": connection.sourceReference,
	}, "ics", "recon"))
	if err = writeCOTPConfirmation(conn, connection); err != nil {
		return
	}
	for {
		payload, err = readS7TPKT(conn)
		if err != nil {
			return
		}
		request, err := parseS7Request(payload)
		if err != nil {
			return
		}
		eventPayload := map[string]any{
			"function": s7FunctionName(request.function), "function_code": request.function,
			"pdu_reference": request.pduReference, "items": request.items,
		}
		sink(protocol.NewEvent("s7.request", src, dst, eventPayload, "ics", "session"))
		switch request.function {
		case 0xf0:
			if err = writeS7Ack(conn, request.pduReference, request.parameters, nil, 0, 0); err != nil {
				return
			}
		case 0x04:
			sink(protocol.NewEvent("s7.read", src, dst, eventPayload, "ics", "read"))
			parameters, data := s7ReadResponse(request.items)
			if err = writeS7Ack(conn, request.pduReference, parameters, data, 0, 0); err != nil {
				return
			}
		case 0x05:
			eventPayload["data"] = hex.EncodeToString(request.data)
			sink(protocol.NewEvent("s7.write", src, dst, eventPayload, "ics", "write"))
			parameters := []byte{0x05, byte(len(request.items))}
			results := make([]byte, len(request.items))
			for index := range results {
				results[index] = 0xff
			}
			if err = writeS7Ack(conn, request.pduReference, parameters, results, 0, 0); err != nil {
				return
			}
		default:
			if err = writeS7Ack(conn, request.pduReference, nil, nil, 0x81, 0x04); err != nil {
				return
			}
		}
	}
}

type cotpConnection struct {
	sourceReference uint16
	sourceTSAP      uint16
	destinationTSAP uint16
}

func parseCOTPConnection(payload []byte) (cotpConnection, error) {
	if len(payload) < 7 || int(payload[0])+1 > len(payload) || payload[1]&0xf0 != 0xe0 {
		return cotpConnection{}, errors.New("invalid COTP connection request")
	}
	result := cotpConnection{sourceReference: binary.BigEndian.Uint16(payload[4:6])}
	for offset := 7; offset+2 <= int(payload[0])+1; {
		code, length := payload[offset], int(payload[offset+1])
		offset += 2
		if offset+length > len(payload) {
			return cotpConnection{}, errors.New("invalid COTP parameter")
		}
		if length == 2 {
			switch code {
			case 0xc1:
				result.sourceTSAP = binary.BigEndian.Uint16(payload[offset : offset+2])
			case 0xc2:
				result.destinationTSAP = binary.BigEndian.Uint16(payload[offset : offset+2])
			}
		}
		offset += length
	}
	return result, nil
}

func writeCOTPConfirmation(writer io.Writer, request cotpConnection) error {
	payload := []byte{
		0x11, 0xd0, byte(request.sourceReference >> 8), byte(request.sourceReference), 0x00, 0x01, 0x00,
		0xc0, 0x01, 0x0a,
		0xc1, 0x02, byte(request.destinationTSAP >> 8), byte(request.destinationTSAP),
		0xc2, 0x02, byte(request.sourceTSAP >> 8), byte(request.sourceTSAP),
	}
	return writeS7TPKT(writer, payload)
}

type s7Item struct {
	TransportSize byte   `json:"transport_size"`
	Length        uint16 `json:"length"`
	DBNumber      uint16 `json:"db_number"`
	Area          string `json:"area"`
	Address       uint32 `json:"address"`
	BitOffset     byte   `json:"bit_offset"`
}

type s7Request struct {
	pduReference uint16
	function     byte
	parameters   []byte
	data         []byte
	items        []s7Item
}

func parseS7Request(payload []byte) (s7Request, error) {
	if len(payload) < 13 || payload[0] != 0x02 || payload[1] != 0xf0 || payload[2] != 0x80 || payload[3] != 0x32 {
		return s7Request{}, errors.New("invalid S7 request")
	}
	s7 := payload[3:]
	if s7[1] != 0x01 {
		return s7Request{}, errors.New("unsupported S7 ROSCTR")
	}
	parameterLength := int(binary.BigEndian.Uint16(s7[6:8]))
	dataLength := int(binary.BigEndian.Uint16(s7[8:10]))
	if parameterLength < 1 || 10+parameterLength+dataLength > len(s7) {
		return s7Request{}, errors.New("invalid S7 payload lengths")
	}
	parameters := append([]byte(nil), s7[10:10+parameterLength]...)
	result := s7Request{
		pduReference: binary.BigEndian.Uint16(s7[4:6]), function: parameters[0], parameters: parameters,
		data: append([]byte(nil), s7[10+parameterLength:10+parameterLength+dataLength]...),
	}
	if result.function == 0x04 || result.function == 0x05 {
		if len(parameters) < 2 || int(parameters[1]) > 32 || len(parameters) < 2+int(parameters[1])*12 {
			return s7Request{}, errors.New("invalid S7 variable specification")
		}
		for index := 0; index < int(parameters[1]); index++ {
			raw := parameters[2+index*12 : 2+(index+1)*12]
			if raw[0] != 0x12 || raw[1] != 0x0a || raw[2] != 0x10 {
				return s7Request{}, errors.New("unsupported S7 variable specification")
			}
			address := uint32(raw[9])<<16 | uint32(raw[10])<<8 | uint32(raw[11])
			result.items = append(result.items, s7Item{
				TransportSize: raw[3], Length: binary.BigEndian.Uint16(raw[4:6]), DBNumber: binary.BigEndian.Uint16(raw[6:8]),
				Area: s7AreaName(raw[8]), Address: address / 8, BitOffset: byte(address % 8),
			})
		}
	}
	return result, nil
}

func s7ReadResponse(items []s7Item) ([]byte, []byte) {
	parameters := []byte{0x04, byte(len(items))}
	data := make([]byte, 0, len(items)*5)
	for _, item := range items {
		length := int(item.Length)
		if length < 1 {
			length = 1
		}
		if length > 64 {
			length = 64
		}
		bitLength := length * 8
		data = append(data, 0xff, 0x04, byte(bitLength>>8), byte(bitLength))
		data = append(data, make([]byte, length)...)
		if length%2 != 0 {
			data = append(data, 0)
		}
	}
	return parameters, data
}

func writeS7Ack(writer io.Writer, pduReference uint16, parameters, data []byte, errorClass, errorCode byte) error {
	header := make([]byte, 12)
	header[0], header[1] = 0x32, 0x03
	binary.BigEndian.PutUint16(header[4:6], pduReference)
	binary.BigEndian.PutUint16(header[6:8], uint16(len(parameters)))
	binary.BigEndian.PutUint16(header[8:10], uint16(len(data)))
	header[10], header[11] = errorClass, errorCode
	payload := append([]byte{0x02, 0xf0, 0x80}, header...)
	payload = append(payload, parameters...)
	payload = append(payload, data...)
	return writeS7TPKT(writer, payload)
}

func readS7TPKT(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, err
	}
	length := int(binary.BigEndian.Uint16(header[2:4]))
	if header[0] != 3 || header[1] != 0 || length < 5 || length > maxS7Frame {
		return nil, errors.New("invalid TPKT frame")
	}
	payload := make([]byte, length-4)
	_, err := io.ReadFull(reader, payload)
	return payload, err
}

func writeS7TPKT(writer io.Writer, payload []byte) error {
	if len(payload)+4 > maxS7Frame {
		return errors.New("TPKT response is too large")
	}
	header := []byte{3, 0, 0, 0}
	binary.BigEndian.PutUint16(header[2:4], uint16(len(payload)+4))
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func s7FunctionName(function byte) string {
	switch function {
	case 0xf0:
		return "setup_communication"
	case 0x04:
		return "read_var"
	case 0x05:
		return "write_var"
	default:
		return "unknown"
	}
}

func s7AreaName(area byte) string {
	switch area {
	case 0x81:
		return "inputs"
	case 0x82:
		return "outputs"
	case 0x83:
		return "markers"
	case 0x84:
		return "data_blocks"
	case 0x1d:
		return "counters"
	case 0x1c:
		return "timers"
	default:
		return fmt.Sprintf("0x%02x", area)
	}
}
