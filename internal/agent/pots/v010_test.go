package pots

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestElasticsearchServiceCapturesRequestAndCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &ElasticsearchService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("elasticsearch", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/_cluster/health", port), nil)
	request.SetBasicAuth("elastic", "secret")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("X-Elastic-Product") != "Elasticsearch" {
		t.Fatalf("Elasticsearch response = %d %#v", response.StatusCode, response.Header)
	}
	credential := waitType(t, events, "elasticsearch.credential")
	if credential.Payload["username"] != "elastic" || credential.Payload["password"] != "secret" {
		t.Fatalf("Elasticsearch credential = %#v", credential.Payload)
	}
}

func TestRTSPServiceCapturesBasicCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &RTSPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("rtsp-camera", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	auth := base64.StdEncoding.EncodeToString([]byte("admin:camera123"))
	_, _ = io.WriteString(conn, fmt.Sprintf("DESCRIBE rtsp://127.0.0.1:%d/live RTSP/1.0\r\nCSeq: 1\r\nAuthorization: Basic %s\r\n\r\n", port, auth))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "200 OK") {
		t.Fatalf("RTSP response = %q, %v", line, err)
	}
	credential := waitType(t, events, "rtsp.credential")
	if credential.Payload["username"] != "admin" || credential.Payload["password"] != "camera123" {
		t.Fatalf("RTSP credential = %#v", credential.Payload)
	}
}

func TestMQTTServiceCapturesConnectAndPublish(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &MQTTService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("mqtt", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	payload := appendMQTTText(nil, "MQTT")
	payload = append(payload, 4, 0xc2, 0, 60)
	payload = appendMQTTText(payload, "sensor-01")
	payload = appendMQTTText(payload, "operator")
	payload = appendMQTTText(payload, "secret")
	_, _ = conn.Write(append([]byte{0x10, byte(len(payload))}, payload...))
	connack := make([]byte, 4)
	if _, err := io.ReadFull(conn, connack); err != nil || !bytes.Equal(connack, []byte{0x20, 0x02, 0x00, 0x00}) {
		t.Fatalf("MQTT CONNACK = %x, %v", connack, err)
	}
	credential := waitType(t, events, "mqtt.credential")
	if credential.Payload["username"] != "operator" || credential.Payload["password"] != "secret" {
		t.Fatalf("MQTT credential = %#v", credential.Payload)
	}
	publish := appendMQTTText(nil, "factory/line1")
	publish = append(publish, []byte("temperature=72")...)
	_, _ = conn.Write(append([]byte{0x30, byte(len(publish))}, publish...))
	event := waitType(t, events, "mqtt.publish")
	if event.Payload["topic"] != "factory/line1" || event.Payload["message"] != "temperature=72" {
		t.Fatalf("MQTT publish = %#v", event.Payload)
	}
}

func TestModbusServiceCapturesReadRequest(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &ModbusService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("modbus", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	request := []byte{0, 7, 0, 0, 0, 6, 1, 3, 0, 100, 0, 2}
	_, _ = conn.Write(request)
	response := make([]byte, 13)
	if _, err := io.ReadFull(conn, response); err != nil || response[7] != 3 || response[8] != 4 {
		t.Fatalf("Modbus response = %x, %v", response, err)
	}
	event := waitType(t, events, "modbus.request")
	if event.Payload["function"] != byte(3) || event.Payload["address"] != uint16(100) {
		t.Fatalf("Modbus event = %#v", event.Payload)
	}
}

func TestMongoDBServiceCapturesHello(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &MongoDBService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("mongodb", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	document := encodeBSONDocument(bsonFields{{"hello", int32(1)}, {"$db", "admin"}})
	body := append([]byte{0, 0, 0, 0, 0}, document...)
	header := make([]byte, 16)
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(header)+len(body)))
	binary.LittleEndian.PutUint32(header[4:8], 77)
	binary.LittleEndian.PutUint32(header[12:16], mongoOpMsg)
	_, _ = conn.Write(append(header, body...))
	responseHeader := make([]byte, 16)
	if _, err := io.ReadFull(conn, responseHeader); err != nil || int32(binary.LittleEndian.Uint32(responseHeader[12:16])) != mongoOpMsg {
		t.Fatalf("MongoDB response header = %x, %v", responseHeader, err)
	}
	responseLength := int(binary.LittleEndian.Uint32(responseHeader[0:4]))
	responseBody := make([]byte, responseLength-16)
	if _, err := io.ReadFull(conn, responseBody); err != nil {
		t.Fatal(err)
	}
	event := waitType(t, events, "mongodb.command")
	if event.Payload["command"] != "hello" || event.Payload["database"] != "admin" {
		t.Fatalf("MongoDB event = %#v", event.Payload)
	}
}

func TestMSSQLServiceCapturesCredentialAndQuery(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &MSSQLService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("mssql", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	login := testTDSLogin7("db-client", "sa", "secret", "finance")
	if err := writeTDSPacket(conn, tdsPacketLogin7, login); err != nil {
		t.Fatal(err)
	}
	packetType, _, err := readTDSMessage(conn)
	if err != nil || packetType != tdsPacketResponse {
		t.Fatalf("MSSQL login response type = %d, %v", packetType, err)
	}
	credential := waitType(t, events, "mssql.credential")
	if credential.Payload["username"] != "sa" || credential.Payload["password"] != "secret" || credential.Payload["database"] != "finance" {
		t.Fatalf("MSSQL credential = %#v", credential.Payload)
	}
	if err := writeTDSPacket(conn, tdsPacketSQLBatch, encodeUTF16LE("SELECT @@VERSION")); err != nil {
		t.Fatal(err)
	}
	event := waitType(t, events, "mssql.query")
	if event.Payload["query"] != "SELECT @@VERSION" {
		t.Fatalf("MSSQL query = %#v", event.Payload)
	}
}

func dialTCP(t *testing.T, port int) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	return conn
}

func appendMQTTText(target []byte, value string) []byte {
	target = append(target, byte(len(value)>>8), byte(len(value)))
	return append(target, value...)
}

func testTDSLogin7(hostname, username, password, database string) []byte {
	payload := make([]byte, 94)
	binary.LittleEndian.PutUint32(payload[4:8], 0x74000004)
	binary.LittleEndian.PutUint32(payload[8:12], 4096)
	appendField := func(pairOffset int, value string, encrypted bool) {
		raw := encodeUTF16LE(value)
		if encrypted {
			for index, item := range raw {
				raw[index] = (item<<4 | item>>4) ^ 0xa5
			}
		}
		binary.LittleEndian.PutUint16(payload[pairOffset:pairOffset+2], uint16(len(payload)))
		binary.LittleEndian.PutUint16(payload[pairOffset+2:pairOffset+4], uint16(len(raw)/2))
		payload = append(payload, raw...)
	}
	appendField(36, hostname, false)
	appendField(40, username, false)
	appendField(44, password, true)
	appendField(48, "sqlcmd", false)
	appendField(52, "sql-prod-01", false)
	appendField(60, "ODBC", false)
	appendField(68, database, false)
	binary.LittleEndian.PutUint32(payload[0:4], uint32(len(payload)))
	return payload
}
