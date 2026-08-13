package pots

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const (
	mongoOpReply    = 1
	mongoOpQuery    = 2004
	mongoOpMsg      = 2013
	maxMongoMessage = 16 << 20
)

var mongoResponseID atomic.Int32

type MongoDBService struct {
	listener net.Listener
	once     sync.Once
}

func (s *MongoDBService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *MongoDBService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *MongoDBService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		request, err := readMongoRequest(conn)
		if err != nil {
			return
		}
		command, document, database, err := parseMongoCommand(request)
		if err != nil {
			return
		}
		payload := map[string]any{"command": command, "database": database, "document": document, "opcode": request.opcode}
		sink(protocol.NewEvent("mongodb.command", src, dst, payload, "database", "session"))
		if strings.EqualFold(command, "saslStart") || strings.EqualFold(command, "saslContinue") {
			username := mongoSCRAMUsername(document["payload"])
			sink(protocol.NewEvent("mongodb.authentication", src, dst, map[string]any{
				"username": username, "mechanism": document["mechanism"], "database": database,
			}, "credential"))
		}
		response := mongoCommandResponse(command)
		if err := writeMongoResponse(conn, request, response); err != nil {
			return
		}
	}
}

type mongoRequest struct {
	requestID int32
	opcode    int32
	payload   []byte
}

func readMongoRequest(reader io.Reader) (mongoRequest, error) {
	var header [16]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return mongoRequest{}, err
	}
	length := int(int32(binary.LittleEndian.Uint32(header[0:4])))
	if length < 16 || length > maxMongoMessage {
		return mongoRequest{}, errors.New("invalid MongoDB message length")
	}
	payload := make([]byte, length-16)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return mongoRequest{}, err
	}
	return mongoRequest{
		requestID: int32(binary.LittleEndian.Uint32(header[4:8])),
		opcode:    int32(binary.LittleEndian.Uint32(header[12:16])),
		payload:   payload,
	}, nil
}

func parseMongoCommand(request mongoRequest) (string, map[string]any, string, error) {
	var documentBytes []byte
	database := "admin"
	switch request.opcode {
	case mongoOpQuery:
		if len(request.payload) < 4 {
			return "", nil, "", errors.New("invalid MongoDB OP_QUERY")
		}
		collectionEnd := bytes.IndexByte(request.payload[4:], 0)
		if collectionEnd < 0 {
			return "", nil, "", errors.New("invalid MongoDB collection")
		}
		collectionEnd += 4
		collection := string(request.payload[4:collectionEnd])
		if value, _, ok := strings.Cut(collection, "."); ok && value != "" {
			database = value
		}
		documentOffset := collectionEnd + 1 + 8
		if documentOffset >= len(request.payload) {
			return "", nil, "", errors.New("missing MongoDB query document")
		}
		documentBytes = request.payload[documentOffset:]
	case mongoOpMsg:
		if len(request.payload) < 6 || request.payload[4] != 0 {
			return "", nil, "", errors.New("unsupported MongoDB OP_MSG section")
		}
		documentBytes = request.payload[5:]
	default:
		return "", nil, "", errors.New("unsupported MongoDB opcode")
	}
	document, command, err := parseBSONDocument(documentBytes)
	if err != nil {
		return "", nil, "", err
	}
	if value, ok := document["$db"].(string); ok && value != "" {
		database = value
	}
	return command, document, database, nil
}

func writeMongoResponse(writer io.Writer, request mongoRequest, document []byte) error {
	responseID := mongoResponseID.Add(1)
	var body []byte
	opcode := int32(mongoOpMsg)
	if request.opcode == mongoOpQuery {
		opcode = mongoOpReply
		body = make([]byte, 20)
		binary.LittleEndian.PutUint32(body[16:20], 1)
		body = append(body, document...)
	} else {
		body = make([]byte, 5)
		body[4] = 0
		body = append(body, document...)
	}
	header := make([]byte, 16)
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(header)+len(body)))
	binary.LittleEndian.PutUint32(header[4:8], uint32(responseID))
	binary.LittleEndian.PutUint32(header[8:12], uint32(request.requestID))
	binary.LittleEndian.PutUint32(header[12:16], uint32(opcode))
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(body)
	return err
}

func mongoCommandResponse(command string) []byte {
	switch strings.ToLower(command) {
	case "ismaster", "hello":
		return encodeBSONDocument(bsonFields{
			{"ok", float64(1)}, {"ismaster", true}, {"isWritablePrimary", true},
			{"maxWireVersion", int32(17)}, {"minWireVersion", int32(0)},
			{"maxBsonObjectSize", int32(16777216)}, {"maxMessageSizeBytes", int32(48000000)},
			{"maxWriteBatchSize", int32(100000)}, {"logicalSessionTimeoutMinutes", int32(30)},
			{"connectionId", int32(42)}, {"localTime", bsonDateTime(time.Now().UnixMilli())}, {"readOnly", false},
		})
	case "buildinfo":
		return encodeBSONDocument(bsonFields{{"ok", float64(1)}, {"version", "6.0.14"}, {"gitVersion", "embedded-honeypot"}})
	case "saslstart", "saslcontinue":
		return encodeBSONDocument(bsonFields{{"ok", float64(0)}, {"errmsg", "Authentication failed."}, {"code", int32(18)}, {"codeName", "AuthenticationFailed"}})
	default:
		return encodeBSONDocument(bsonFields{{"ok", float64(1)}})
	}
}

func mongoSCRAMUsername(value any) string {
	payload, ok := value.([]byte)
	if !ok {
		return ""
	}
	text := string(payload)
	index := strings.Index(text, "n=")
	if index < 0 {
		return ""
	}
	username := text[index+2:]
	if end := strings.IndexByte(username, ','); end >= 0 {
		username = username[:end]
	}
	return strings.ReplaceAll(username, "=2C", ",")
}

func parseBSONDocument(raw []byte) (map[string]any, string, error) {
	if len(raw) < 5 {
		return nil, "", errors.New("invalid BSON document")
	}
	length := int(int32(binary.LittleEndian.Uint32(raw[:4])))
	if length < 5 || length > len(raw) || raw[length-1] != 0 {
		return nil, "", errors.New("invalid BSON document length")
	}
	result := map[string]any{}
	firstKey := ""
	for offset := 4; offset < length-1; {
		typeByte := raw[offset]
		offset++
		keyEnd := bytes.IndexByte(raw[offset:length], 0)
		if keyEnd < 0 {
			return nil, "", errors.New("invalid BSON key")
		}
		key := string(raw[offset : offset+keyEnd])
		offset += keyEnd + 1
		if firstKey == "" {
			firstKey = key
		}
		value, used, err := parseBSONValue(typeByte, raw[offset:length])
		if err != nil {
			return nil, "", err
		}
		result[key] = value
		offset += used
	}
	return result, firstKey, nil
}

func parseBSONValue(typeByte byte, raw []byte) (any, int, error) {
	switch typeByte {
	case 0x01:
		if len(raw) < 8 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(raw[:8])), 8, nil
	case 0x02:
		if len(raw) < 4 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		length := int(int32(binary.LittleEndian.Uint32(raw[:4])))
		if length < 1 || 4+length > len(raw) {
			return nil, 0, errors.New("invalid BSON string")
		}
		return string(raw[4 : 4+length-1]), 4 + length, nil
	case 0x03, 0x04:
		if len(raw) < 4 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		length := int(int32(binary.LittleEndian.Uint32(raw[:4])))
		if length < 5 || length > len(raw) {
			return nil, 0, errors.New("invalid nested BSON document")
		}
		value, _, err := parseBSONDocument(raw[:length])
		return value, length, err
	case 0x05:
		if len(raw) < 5 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		length := int(int32(binary.LittleEndian.Uint32(raw[:4])))
		if length < 0 || 5+length > len(raw) {
			return nil, 0, errors.New("invalid BSON binary")
		}
		return append([]byte(nil), raw[5:5+length]...), 5 + length, nil
	case 0x08:
		if len(raw) < 1 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		return raw[0] != 0, 1, nil
	case 0x09, 0x11, 0x12:
		if len(raw) < 8 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		return int64(binary.LittleEndian.Uint64(raw[:8])), 8, nil
	case 0x0a:
		return nil, 0, nil
	case 0x10:
		if len(raw) < 4 {
			return nil, 0, io.ErrUnexpectedEOF
		}
		return int32(binary.LittleEndian.Uint32(raw[:4])), 4, nil
	default:
		return nil, 0, errors.New("unsupported BSON value type")
	}
}

type bsonField struct {
	key   string
	value any
}
type bsonFields []bsonField
type bsonDateTime int64

func encodeBSONDocument(fields bsonFields) []byte {
	document := make([]byte, 4, 256)
	for _, field := range fields {
		document = appendBSONElement(document, field)
	}
	document = append(document, 0)
	binary.LittleEndian.PutUint32(document[:4], uint32(len(document)))
	return document
}

func appendBSONElement(document []byte, field bsonField) []byte {
	appendKey := func(typeByte byte) {
		document = append(document, typeByte)
		document = append(document, field.key...)
		document = append(document, 0)
	}
	switch value := field.value.(type) {
	case string:
		appendKey(0x02)
		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(value)+1))
		document = append(document, length...)
		document = append(document, value...)
		document = append(document, 0)
	case bool:
		appendKey(0x08)
		if value {
			document = append(document, 1)
		} else {
			document = append(document, 0)
		}
	case int32:
		appendKey(0x10)
		var raw [4]byte
		binary.LittleEndian.PutUint32(raw[:], uint32(value))
		document = append(document, raw[:]...)
	case int64:
		appendKey(0x12)
		var raw [8]byte
		binary.LittleEndian.PutUint64(raw[:], uint64(value))
		document = append(document, raw[:]...)
	case bsonDateTime:
		appendKey(0x09)
		var raw [8]byte
		binary.LittleEndian.PutUint64(raw[:], uint64(value))
		document = append(document, raw[:]...)
	case float64:
		appendKey(0x01)
		document = append(document, float64Bytes(value)...)
	}
	return document
}

func float64Bytes(value float64) []byte {
	var raw [8]byte
	binary.LittleEndian.PutUint64(raw[:], math.Float64bits(value))
	return raw[:]
}
