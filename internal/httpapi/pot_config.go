package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"strings"
)

const maxPotConfigBytes = 32 << 10

func normalizePotConfig(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage("{}")
	}
	if len(raw) > maxPotConfigBytes {
		return nil, errors.New("pot config exceeds 32 KiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var config map[string]any
	if err := decoder.Decode(&config); err != nil || config == nil {
		return nil, errors.New("pot config must be a JSON object")
	}
	if value, exists := config["bind"]; exists {
		bind, ok := value.(string)
		if !ok {
			return nil, errors.New("pot bind must be an IP address")
		}
		bind = strings.TrimSpace(bind)
		address := strings.Trim(bind, "[]")
		if zone := strings.LastIndex(address, "%"); zone > 0 {
			address = address[:zone]
		}
		ip := net.ParseIP(address)
		if ip == nil {
			return nil, errors.New("pot bind must be a valid IPv4 or IPv6 address")
		}
		config["bind"] = ip.String()
	}
	return json.Marshal(config)
}
