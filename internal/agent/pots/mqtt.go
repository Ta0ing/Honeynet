package pots

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxMQTTPacket = 1 << 20

type MQTTService struct {
	listener net.Listener
	once     sync.Once
}

func (s *MQTTService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *MQTTService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *MQTTService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(conn)
	header, payload, err := readMQTTPacket(reader)
	if err != nil || header>>4 != 1 {
		return
	}
	connect, err := parseMQTTConnect(payload)
	if err != nil {
		return
	}
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	sink(protocol.NewEvent("mqtt.connect", src, dst, map[string]any{
		"client_id": connect.clientID, "protocol": connect.protocol, "version": connect.version,
		"keep_alive": connect.keepAlive, "clean_start": connect.flags&0x02 != 0,
	}, "iot", "session"))
	if connect.username != "" || connect.password != "" {
		sink(protocol.NewEvent("mqtt.credential", src, dst, map[string]any{
			"username": connect.username, "password": connect.password, "client_id": connect.clientID,
		}, "credential"))
	}
	if connect.version == 5 {
		_, _ = conn.Write([]byte{0x20, 0x03, 0x00, 0x00, 0x00})
	} else {
		_, _ = conn.Write([]byte{0x20, 0x02, 0x00, 0x00})
	}
	for {
		header, payload, err = readMQTTPacket(reader)
		if err != nil {
			return
		}
		switch header >> 4 {
		case 3:
			handleMQTTPublish(conn, header, payload, src, dst, sink)
		case 8:
			handleMQTTSubscribe(conn, connect.version, payload, src, dst, sink)
		case 12:
			_, _ = conn.Write([]byte{0xd0, 0x00})
		case 14:
			return
		}
	}
}

type mqttConnect struct {
	protocol           string
	version            byte
	flags              byte
	keepAlive          uint16
	clientID           string
	username, password string
}

func parseMQTTConnect(payload []byte) (mqttConnect, error) {
	cursor := mqttCursor{data: payload}
	protocolName, ok := cursor.text()
	if !ok || cursor.remaining() < 4 {
		return mqttConnect{}, errors.New("invalid MQTT CONNECT")
	}
	result := mqttConnect{protocol: protocolName, version: cursor.byte(), flags: cursor.byte(), keepAlive: cursor.uint16()}
	if result.protocol != "MQTT" && result.protocol != "MQIsdp" {
		return mqttConnect{}, errors.New("unknown MQTT protocol")
	}
	if result.version == 5 && !cursor.skipVariableBytes() {
		return mqttConnect{}, errors.New("invalid MQTT properties")
	}
	if result.clientID, ok = cursor.text(); !ok {
		return mqttConnect{}, errors.New("invalid MQTT client ID")
	}
	if result.flags&0x04 != 0 {
		if result.version == 5 && !cursor.skipVariableBytes() {
			return mqttConnect{}, errors.New("invalid MQTT will properties")
		}
		if _, ok = cursor.text(); !ok {
			return mqttConnect{}, errors.New("invalid MQTT will topic")
		}
		if _, ok = cursor.binary(); !ok {
			return mqttConnect{}, errors.New("invalid MQTT will payload")
		}
	}
	if result.flags&0x80 != 0 {
		result.username, ok = cursor.text()
		if !ok {
			return mqttConnect{}, errors.New("invalid MQTT username")
		}
	}
	if result.flags&0x40 != 0 {
		password, valid := cursor.binary()
		if !valid {
			return mqttConnect{}, errors.New("invalid MQTT password")
		}
		result.password = string(password)
	}
	return result, nil
}

func readMQTTPacket(reader io.Reader) (byte, []byte, error) {
	var first [1]byte
	if _, err := io.ReadFull(reader, first[:]); err != nil {
		return 0, nil, err
	}
	remaining, _, err := readMQTTVariableInteger(reader)
	if err != nil || remaining < 0 || remaining > maxMQTTPacket {
		return 0, nil, errors.New("invalid MQTT remaining length")
	}
	payload := make([]byte, remaining)
	_, err = io.ReadFull(reader, payload)
	return first[0], payload, err
}

func readMQTTVariableInteger(reader io.Reader) (int, int, error) {
	value, multiplier := 0, 1
	for index := 0; index < 4; index++ {
		var raw [1]byte
		if _, err := io.ReadFull(reader, raw[:]); err != nil {
			return 0, index, err
		}
		value += int(raw[0]&0x7f) * multiplier
		if raw[0]&0x80 == 0 {
			return value, index + 1, nil
		}
		multiplier *= 128
	}
	return 0, 4, errors.New("MQTT variable integer is too long")
}

func handleMQTTPublish(conn net.Conn, header byte, payload []byte, src, dst protocol.Endpoint, sink Sink) {
	cursor := mqttCursor{data: payload}
	topic, ok := cursor.text()
	if !ok {
		return
	}
	qos := int((header >> 1) & 0x03)
	packetID := uint16(0)
	if qos > 0 {
		if cursor.remaining() < 2 {
			return
		}
		packetID = cursor.uint16()
	}
	message := string(cursor.data[cursor.offset:])
	sink(protocol.NewEvent("mqtt.publish", src, dst, map[string]any{
		"topic": topic, "message": message, "qos": qos, "retain": header&0x01 != 0,
	}, "iot", "session"))
	if qos == 1 {
		_, _ = conn.Write([]byte{0x40, 0x02, byte(packetID >> 8), byte(packetID)})
	}
}

func handleMQTTSubscribe(conn net.Conn, version byte, payload []byte, src, dst protocol.Endpoint, sink Sink) {
	cursor := mqttCursor{data: payload}
	if cursor.remaining() < 2 {
		return
	}
	packetID := cursor.uint16()
	if version == 5 && !cursor.skipVariableBytes() {
		return
	}
	topics := make([]string, 0, 4)
	for cursor.remaining() > 0 {
		topic, ok := cursor.text()
		if !ok || cursor.remaining() < 1 {
			return
		}
		topics = append(topics, topic)
		_ = cursor.byte()
	}
	sink(protocol.NewEvent("mqtt.subscribe", src, dst, map[string]any{"topics": topics}, "iot", "session"))
	response := []byte{0x90, 0, byte(packetID >> 8), byte(packetID)}
	if version == 5 {
		response = append(response, 0)
	}
	response = append(response, make([]byte, len(topics))...)
	response[1] = byte(len(response) - 2)
	_, _ = conn.Write(response)
}

type mqttCursor struct {
	data   []byte
	offset int
}

func (c *mqttCursor) remaining() int { return len(c.data) - c.offset }
func (c *mqttCursor) byte() byte {
	value := c.data[c.offset]
	c.offset++
	return value
}
func (c *mqttCursor) uint16() uint16 {
	value := binary.BigEndian.Uint16(c.data[c.offset : c.offset+2])
	c.offset += 2
	return value
}
func (c *mqttCursor) binary() ([]byte, bool) {
	if c.remaining() < 2 {
		return nil, false
	}
	length := int(c.uint16())
	if length > c.remaining() {
		return nil, false
	}
	value := c.data[c.offset : c.offset+length]
	c.offset += length
	return value, true
}
func (c *mqttCursor) text() (string, bool) {
	value, ok := c.binary()
	return string(value), ok
}
func (c *mqttCursor) skipVariableBytes() bool {
	length, used, err := readMQTTVariableInteger(strings.NewReader(string(c.data[c.offset:])))
	if err != nil || used+length > c.remaining() {
		return false
	}
	c.offset += used + length
	return true
}
