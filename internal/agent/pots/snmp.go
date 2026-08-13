package pots

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type SNMPService struct {
	conn net.PacketConn
	once sync.Once
}

func (s *SNMPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	conn, err := net.ListenPacket("udp", listenAddress(target))
	if err != nil {
		return err
	}
	s.conn = conn
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	go s.serve(conn, sink)
	return nil
}

func (s *SNMPService) Stop() error {
	if s.conn == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.conn.Close() })
	return err
}

func (s *SNMPService) serve(conn net.PacketConn, sink Sink) {
	buffer := make([]byte, 64<<10)
	for {
		count, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		request, err := parseSNMPMessage(buffer[:count])
		if err != nil {
			continue
		}
		src, dst := endpoint(remote), endpoint(conn.LocalAddr())
		sink(protocol.NewEvent("snmp.request", src, dst, map[string]any{
			"version": request.version + 1, "community": request.community, "pdu_type": snmpPDUName(request.pduTag),
			"request_id": request.requestID, "oids": request.oids,
		}, "network", "recon"))
		if request.community != "" {
			sink(protocol.NewEvent("snmp.community", src, dst, map[string]any{"community": request.community, "version": request.version + 1}, "credential"))
		}
		_, _ = conn.WriteTo(snmpResponse(request), remote)
	}
}

type snmpRequest struct {
	version   int64
	community string
	pduTag    byte
	requestID int64
	oids      []string
}

func parseSNMPMessage(packet []byte) (snmpRequest, error) {
	tag, message, rest, err := readBERElement(packet)
	if err != nil || tag != 0x30 || len(rest) != 0 {
		return snmpRequest{}, errors.New("invalid SNMP message")
	}
	versionTag, versionRaw, message, err := readBERElement(message)
	if err != nil || versionTag != 0x02 {
		return snmpRequest{}, errors.New("SNMP version is missing")
	}
	version, err := parseBERInteger(versionRaw)
	if err != nil || (version != 0 && version != 1) {
		return snmpRequest{}, errors.New("unsupported SNMP version")
	}
	communityTag, community, message, err := readBERElement(message)
	if err != nil || communityTag != 0x04 {
		return snmpRequest{}, errors.New("SNMP community is missing")
	}
	pduTag, pdu, _, err := readBERElement(message)
	if err != nil || pduTag < 0xa0 || pduTag > 0xa3 {
		return snmpRequest{}, errors.New("unsupported SNMP PDU")
	}
	requestIDTag, requestIDRaw, pdu, err := readBERElement(pdu)
	if err != nil || requestIDTag != 0x02 {
		return snmpRequest{}, errors.New("SNMP request ID is missing")
	}
	requestID, err := parseBERInteger(requestIDRaw)
	if err != nil {
		return snmpRequest{}, err
	}
	for index := 0; index < 2; index++ {
		integerTag, _, remaining, parseErr := readBERElement(pdu)
		if parseErr != nil || integerTag != 0x02 {
			return snmpRequest{}, errors.New("invalid SNMP error fields")
		}
		pdu = remaining
	}
	listTag, list, _, err := readBERElement(pdu)
	if err != nil || listTag != 0x30 {
		return snmpRequest{}, errors.New("SNMP varbind list is missing")
	}
	oids := make([]string, 0, 4)
	for len(list) > 0 {
		varBindTag, varBind, remaining, parseErr := readBERElement(list)
		if parseErr != nil || varBindTag != 0x30 {
			return snmpRequest{}, errors.New("invalid SNMP varbind")
		}
		oidTag, oidRaw, _, parseErr := readBERElement(varBind)
		if parseErr != nil || oidTag != 0x06 {
			return snmpRequest{}, errors.New("invalid SNMP OID")
		}
		oids = append(oids, decodeBEROID(oidRaw))
		list = remaining
	}
	return snmpRequest{version: version, community: string(community), pduTag: pduTag, requestID: requestID, oids: oids}, nil
}

func snmpResponse(request snmpRequest) []byte {
	varBinds := make([]byte, 0, len(request.oids)*32)
	for _, oid := range request.oids {
		value := berElement(0x05, nil)
		if oid == "1.3.6.1.2.1.1.1.0" {
			value = berElement(0x04, []byte("Linux mail-gateway 5.15.0-91-generic x86_64"))
		}
		varBind := append(berElement(0x06, encodeBEROID(oid)), value...)
		varBinds = append(varBinds, berElement(0x30, varBind)...)
	}
	pdu := append(berElement(0x02, berIntegerValue(request.requestID)), berElement(0x02, []byte{0})...)
	pdu = append(pdu, berElement(0x02, []byte{0})...)
	pdu = append(pdu, berElement(0x30, varBinds)...)
	message := append(berElement(0x02, berIntegerValue(request.version)), berElement(0x04, []byte(request.community))...)
	message = append(message, berElement(0xa2, pdu)...)
	return berElement(0x30, message)
}

func decodeBEROID(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	encodedParts := make([]int, 0, 8)
	value := 0
	for _, item := range raw {
		value = value<<7 | int(item&0x7f)
		if item&0x80 == 0 {
			encodedParts = append(encodedParts, value)
			value = 0
		}
	}
	if len(encodedParts) == 0 {
		return ""
	}
	first, second := 2, encodedParts[0]-80
	if encodedParts[0] < 40 {
		first, second = 0, encodedParts[0]
	} else if encodedParts[0] < 80 {
		first, second = 1, encodedParts[0]-40
	}
	parts := append([]int{first, second}, encodedParts[1:]...)
	text := make([]string, len(parts))
	for index, item := range parts {
		text[index] = strconv.Itoa(item)
	}
	return strings.Join(text, ".")
}

func encodeBEROID(value string) []byte {
	partsText := strings.Split(value, ".")
	if len(partsText) < 2 {
		return nil
	}
	parts := make([]int, len(partsText))
	for index, part := range partsText {
		parts[index], _ = strconv.Atoi(part)
	}
	if parts[0] < 0 || parts[0] > 2 || parts[1] < 0 || (parts[0] < 2 && parts[1] > 39) {
		return nil
	}
	encodedParts := append([]int{parts[0]*40 + parts[1]}, parts[2:]...)
	raw := make([]byte, 0, len(encodedParts)+2)
	for _, part := range encodedParts {
		raw = append(raw, encodeBEROIDPart(part)...)
	}
	return raw
}

func encodeBEROIDPart(value int) []byte {
	if value < 0 {
		return nil
	}
	encoded := []byte{byte(value & 0x7f)}
	for value >>= 7; value > 0; value >>= 7 {
		encoded = append([]byte{byte(value&0x7f) | 0x80}, encoded...)
	}
	return encoded
}

func snmpPDUName(tag byte) string {
	switch tag {
	case 0xa0:
		return "get-request"
	case 0xa1:
		return "get-next-request"
	case 0xa2:
		return "get-response"
	case 0xa3:
		return "set-request"
	default:
		return "unknown"
	}
}
