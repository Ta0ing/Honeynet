package pots

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"
)

// httpRequestSnapshot is a defender-friendly HTTP/1-style representation of
// the request received by the honeypot. net/http has already parsed the wire,
// so this is a normalized packet snapshot rather than a PCAP replacement.
func httpRequestSnapshot(request *http.Request, body []byte, truncated bool) (string, string) {
	var builder strings.Builder
	protocol := request.Proto
	if protocol == "" {
		protocol = "HTTP/1.1"
	}
	fmt.Fprintf(&builder, "%s %s %s\r\n", request.Method, request.URL.RequestURI(), protocol)
	if request.Host != "" {
		fmt.Fprintf(&builder, "Host: %s\r\n", request.Host)
	}
	keys := make([]string, 0, len(request.Header))
	for key := range request.Header {
		if !strings.EqualFold(key, "Host") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range request.Header.Values(key) {
			fmt.Fprintf(&builder, "%s: %s\r\n", key, value)
		}
	}
	builder.WriteString("\r\n")
	bodyBase64 := ""
	if utf8.Valid(body) {
		builder.Write(body)
	} else if len(body) > 0 {
		bodyBase64 = base64.StdEncoding.EncodeToString(body)
		builder.WriteString("[binary body retained as payload.body_base64]")
	}
	if truncated {
		builder.WriteString("\r\n[request body truncated at capture limit]")
	}
	return builder.String(), bodyBase64
}
