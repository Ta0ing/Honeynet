package analytics

import (
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"
)

var credentialPlaceholders = map[string]struct{}{
	"-": {}, "--": {}, "n/a": {}, "na": {}, "nil": {}, "null": {}, "none": {},
	"undefined": {}, "unknown": {}, "<nil>": {}, "[object object]": {}, "{}": {}, "[]": {},
	"请输入用户名": {}, "请输入密码": {}, "请输入账号": {}, "请输入口令": {},
	"用户名": {}, "密码": {}, "账号": {}, "口令": {}, "******": {}, "********": {},
}

func ExtractCredential(event Event) *Credential {
	if event.Credential != nil {
		credential := normalizeCredential(*event.Credential)
		if credential.Username != "" || credential.Password != "" || credential.AuthResponse != "" {
			return &credential
		}
	}
	if strings.HasPrefix(event.EventType, "decoy.") || event.Service == "decoy" || !ValidIP(event.SourceIP) || len(event.Payload) == 0 {
		return nil
	}
	payload := map[string]any{}
	if json.Unmarshal(event.Payload, &payload) != nil {
		return nil
	}
	username, _ := payloadString(payload, 512, "username", "user")
	password, _ := payloadString(payload, 4096, "password")
	authResponse, kind := payloadString(payload, 16<<10, "auth_response", "digest", "nt_response", "response", "community")
	if username == "" && password == "" && authResponse == "" {
		return nil
	}
	mechanism, _ := payloadString(payload, 256, "mechanism")
	if mechanism == "" {
		switch kind {
		case "auth_response":
			mechanism = "认证响应"
		case "digest":
			mechanism = "摘要认证"
		case "nt_response":
			mechanism = "NTLM"
		case "response":
			mechanism = "挑战响应"
		case "community":
			mechanism = "SNMP Community"
		}
	}
	return &Credential{Username: username, Password: password, AuthResponse: authResponse, Mechanism: mechanism}
}

func normalizeCredential(credential Credential) Credential {
	credential.Username = cleanCredentialValue(credential.Username, 512)
	credential.Password = cleanCredentialValue(credential.Password, 4096)
	credential.AuthResponse = cleanCredentialValue(credential.AuthResponse, 16<<10)
	credential.Mechanism = cleanCredentialValue(credential.Mechanism, 256)
	return credential
}

func payloadString(payload map[string]any, maximum int, keys ...string) (string, string) {
	for _, key := range keys {
		value, ok := payload[key].(string)
		if ok {
			if cleaned := cleanCredentialValue(value, maximum); cleaned != "" {
				return cleaned, key
			}
		}
	}
	return "", ""
}

func cleanCredentialValue(value string, maximum int) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maximum || !utf8.ValidString(value) {
		return ""
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return ""
		}
	}
	lower := strings.ToLower(value)
	if _, found := credentialPlaceholders[lower]; found {
		return ""
	}
	if (strings.HasPrefix(lower, "{{") && strings.HasSuffix(lower, "}}")) ||
		(strings.HasPrefix(lower, "${") && strings.HasSuffix(lower, "}")) {
		return ""
	}
	if strings.Contains(lower, "<!doctype") || strings.Contains(lower, "<html") || strings.Contains(lower, "<script") {
		return ""
	}
	if (strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[")) && json.Valid([]byte(value)) {
		return ""
	}
	return value
}
