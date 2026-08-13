package pots

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type RTSPService struct {
	listener net.Listener
	once     sync.Once
}

func (s *RTSPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *RTSPService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *RTSPService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReaderSize(conn, 16<<10)
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		requestLine = strings.TrimSpace(requestLine)
		parts := strings.Fields(requestLine)
		if len(parts) != 3 || !strings.HasPrefix(parts[2], "RTSP/") {
			return
		}
		headers, err := readRTSPHeaders(reader)
		if err != nil {
			return
		}
		method, requestURI := strings.ToUpper(parts[0]), parts[1]
		sink(protocol.NewEvent("rtsp.request", src, dst, map[string]any{
			"method": method, "uri": requestURI, "user_agent": headers["user-agent"],
			"transport": headers["transport"],
		}, "camera", "session"))
		username, password, authenticated := parseRTSPBasicAuth(headers["authorization"])
		if authenticated {
			sink(protocol.NewEvent("rtsp.credential", src, dst, map[string]any{
				"username": username, "password": password, "uri": requestURI,
			}, "credential"))
		}
		cseq := headers["cseq"]
		if cseq == "" {
			cseq = "1"
		}
		if method != "OPTIONS" && !authenticated {
			writeRTSPResponse(conn, 401, "Unauthorized", cseq, map[string]string{"WWW-Authenticate": `Basic realm="IP Camera"`}, "")
			continue
		}
		switch method {
		case "OPTIONS":
			writeRTSPResponse(conn, 200, "OK", cseq, map[string]string{"Public": "OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, TEARDOWN, GET_PARAMETER"}, "")
		case "DESCRIBE":
			body := "v=0\r\no=- 0 0 IN IP4 192.0.2.20\r\ns=Network Camera\r\nt=0 0\r\na=control:*\r\nm=video 0 RTP/AVP 96\r\na=rtpmap:96 H264/90000\r\na=control:trackID=1\r\n"
			writeRTSPResponse(conn, 200, "OK", cseq, map[string]string{"Content-Type": "application/sdp", "Content-Base": requestURI + "/"}, body)
		case "SETUP":
			transport := headers["transport"]
			if transport == "" {
				transport = "RTP/AVP/TCP;unicast;interleaved=0-1"
			}
			writeRTSPResponse(conn, 200, "OK", cseq, map[string]string{"Session": "84510231;timeout=60", "Transport": transport + ";server_port=5000-5001"}, "")
		case "PLAY", "PAUSE", "GET_PARAMETER":
			writeRTSPResponse(conn, 200, "OK", cseq, map[string]string{"Session": "84510231;timeout=60"}, "")
		case "TEARDOWN":
			writeRTSPResponse(conn, 200, "OK", cseq, map[string]string{"Session": "84510231"}, "")
			return
		default:
			writeRTSPResponse(conn, 405, "Method Not Allowed", cseq, nil, "")
		}
	}
}

func readRTSPHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := map[string]string{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return headers, nil
		}
		key, value, ok := strings.Cut(line, ":")
		if ok {
			headers[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		}
		if len(headers) > 100 {
			return nil, fmt.Errorf("too many RTSP headers")
		}
	}
}

func parseRTSPBasicAuth(value string) (string, string, bool) {
	if !strings.HasPrefix(strings.ToLower(value), "basic ") {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value[6:]))
	if err != nil {
		return "", "", false
	}
	username, password, ok := strings.Cut(string(raw), ":")
	return username, password, ok
}

func writeRTSPResponse(writer io.Writer, status int, reason, cseq string, headers map[string]string, body string) {
	_, _ = fmt.Fprintf(writer, "RTSP/1.0 %d %s\r\nCSeq: %s\r\nServer: Embedded RTSP Server/1.0\r\n", status, reason, cseq)
	for key, value := range headers {
		_, _ = fmt.Fprintf(writer, "%s: %s\r\n", key, value)
	}
	if body != "" {
		_, _ = fmt.Fprintf(writer, "Content-Length: %s\r\n", strconv.Itoa(len(body)))
	}
	_, _ = io.WriteString(writer, "\r\n"+body)
}
