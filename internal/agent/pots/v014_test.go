package pots

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestS7ServiceNegotiatesAndCapturesReadVar(t *testing.T) {
	events := make(chan protocol.Event, 16)
	port := freePort(t)
	service := &S7Service{}
	if err := service.Start(context.Background(), target("s7comm", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	connectionRequest := []byte{0x11, 0xe0, 0, 0, 0, 1, 0, 0xc0, 1, 0x0a, 0xc1, 2, 0x01, 0x00, 0xc2, 2, 0x01, 0x02}
	if err = writeS7TPKT(conn, connectionRequest); err != nil {
		t.Fatal(err)
	}
	confirmation, err := readS7TPKT(conn)
	if err != nil || len(confirmation) < 2 || confirmation[1] != 0xd0 {
		t.Fatalf("COTP confirmation = %x, %v", confirmation, err)
	}
	connection := waitType(t, events, "s7.connection")
	if connection.Payload["source_tsap"] != "0100" || connection.Payload["destination_tsap"] != "0102" {
		t.Fatalf("S7 connection event = %#v", connection.Payload)
	}

	setupParameters := []byte{0xf0, 0, 0, 1, 0, 1, 0x01, 0xe0}
	if err = writeS7Job(conn, 1, setupParameters, nil); err != nil {
		t.Fatal(err)
	}
	setupResponse, err := readS7TPKT(conn)
	if err != nil || len(setupResponse) < 16 || setupResponse[4] != 0x03 || setupResponse[15] != 0xf0 {
		t.Fatalf("S7 setup response = %x, %v", setupResponse, err)
	}

	readParameters := []byte{0x04, 0x01, 0x12, 0x0a, 0x10, 0x02, 0, 1, 0, 1, 0x84, 0, 0, 0}
	if err = writeS7Job(conn, 2, readParameters, nil); err != nil {
		t.Fatal(err)
	}
	readResponse, err := readS7TPKT(conn)
	if err != nil || len(readResponse) < 21 || readResponse[15] != 0x04 || readResponse[17] != 0xff {
		t.Fatalf("S7 read response = %x, %v", readResponse, err)
	}
	event := waitType(t, events, "s7.read")
	items, ok := event.Payload["items"].([]s7Item)
	if !ok || len(items) != 1 || items[0].Area != "data_blocks" || items[0].DBNumber != 1 || items[0].Address != 0 {
		t.Fatalf("S7 read event = %#v", event.Payload)
	}
}

func writeS7Job(writer io.Writer, reference uint16, parameters, data []byte) error {
	header := make([]byte, 10)
	header[0], header[1] = 0x32, 0x01
	binary.BigEndian.PutUint16(header[4:6], reference)
	binary.BigEndian.PutUint16(header[6:8], uint16(len(parameters)))
	binary.BigEndian.PutUint16(header[8:10], uint16(len(data)))
	payload := append([]byte{0x02, 0xf0, 0x80}, header...)
	payload = append(payload, parameters...)
	payload = append(payload, data...)
	return writeS7TPKT(writer, payload)
}

func TestCoAPServiceAnswersConfirmableGet(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freeUDPPort(t)
	service := &CoAPService{}
	if err := service.Start(context.Background(), target("coap", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	request := []byte{0x42, 0x01, 0x12, 0x34, 0xaa, 0xbb, 0xb6}
	request = append(request, []byte("status")...)
	if _, err = conn.Write(request); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 1024)
	count, err := conn.Read(response)
	if err != nil || count < 7 || response[0] != 0x62 || response[1] != 69 || !bytes.Equal(response[4:6], []byte{0xaa, 0xbb}) {
		t.Fatalf("CoAP response = %x, %v", response[:count], err)
	}
	event := waitType(t, events, "coap.request")
	if event.Payload["method"] != "GET" || event.Payload["path"] != "/status" || event.Payload["token"] != "aabb" {
		t.Fatalf("CoAP event = %#v", event.Payload)
	}
}

func TestBACnetServiceAnswersWhoIsWithIAm(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freeUDPPort(t)
	service := &BACnetService{}
	target := target("bacnet", port)
	target.Config["device_id"] = float64(4242)
	target.Config["vendor_id"] = float64(1200)
	if err := service.Start(context.Background(), target, func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	whoIs := []byte{0x81, 0x0b, 0, 12, 0x01, 0x20, 0xff, 0xff, 0, 0xff, 0x10, 0x08}
	if _, err = conn.Write(whoIs); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 1024)
	count, err := conn.Read(response)
	if err != nil || count < 15 || response[0] != 0x81 || response[1] != 0x0a || response[6] != 0x10 || response[7] != 0 || response[8] != 0xc4 {
		t.Fatalf("BACnet I-Am = %x, %v", response[:count], err)
	}
	objectID := binary.BigEndian.Uint32(response[9:13])
	if objectID>>22 != 8 || objectID&0x3fffff != 4242 {
		t.Fatalf("BACnet device object = %#x", objectID)
	}
	event := waitType(t, events, "bacnet.discovery")
	if event.Payload["request"] != "who_is" || event.Payload["device_id"] != 4242 || event.Payload["vendor_id"] != 1200 {
		t.Fatalf("BACnet discovery event = %#v", event.Payload)
	}

	objectReference := make([]byte, 4)
	binary.BigEndian.PutUint32(objectReference, uint32(8)<<22|4242)
	readPropertyAPDU := []byte{0x00, 0x05, 0x07, 0x0c, 0x0c}
	readPropertyAPDU = append(readPropertyAPDU, objectReference...)
	readPropertyAPDU = append(readPropertyAPDU, 0x19, 77)
	readProperty := bacnetBVLC(0x0a, append([]byte{0x01, 0x00}, readPropertyAPDU...))
	if _, err = conn.Write(readProperty); err != nil {
		t.Fatal(err)
	}
	count, err = conn.Read(response)
	if err != nil || count < 13 || response[6] != 0x50 || response[7] != 7 || response[8] != 12 {
		t.Fatalf("BACnet ReadProperty error = %x, %v", response[:count], err)
	}
	readEvent := waitType(t, events, "bacnet.read_property")
	if readEvent.Payload["object_type"] != uint16(8) || readEvent.Payload["object_instance"] != uint32(4242) || readEvent.Payload["property_id"] != uint32(77) {
		t.Fatalf("BACnet ReadProperty event = %#v", readEvent.Payload)
	}
}
