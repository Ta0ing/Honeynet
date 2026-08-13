package httpapi

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

const maxNodeCapabilities = 128

func normalizeCapabilities(values []string) datatypes.JSON {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" || len(value) > 128 || !validCapability(value) {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
		if len(normalized) == maxNodeCapabilities {
			break
		}
	}
	sort.Strings(normalized)
	raw, _ := json.Marshal(normalized)
	return datatypes.JSON(raw)
}

func validCapability(value string) bool {
	for _, char := range value {
		if unicode.IsLower(char) || unicode.IsDigit(char) || char == '.' || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func nodeSupportsService(node store.Node, serviceCode string) (known, supported bool) {
	if len(node.Capabilities) == 0 {
		return false, false
	}
	var values []string
	if json.Unmarshal(node.Capabilities, &values) != nil {
		return false, false
	}
	wanted := "pot." + strings.ToLower(strings.TrimSpace(serviceCode))
	for _, value := range values {
		if value == wanted {
			return true, true
		}
	}
	return true, false
}
