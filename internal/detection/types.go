package detection

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	PatternContains = "contains"
	PatternRegexp   = "regexp"
)

// Rule is the portable, deliberately restricted rule format shared by Server
// and Agent. It does not execute arbitrary YARA expressions on a node.
type Rule struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Severity    string    `json:"severity"`
	Source      string    `json:"source,omitempty"`
	ExternalID  string    `json:"external_id,omitempty"`
	Revision    int64     `json:"revision"`
	Patterns    []Pattern `json:"patterns"`
}

type Pattern struct {
	ID       string `json:"id"`
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	NoCase   bool   `json:"nocase,omitempty"`
	MinCount int    `json:"min_count,omitempty"`
}

type Event struct {
	EventType string
	Service   string
	RawPacket string
	Payload   map[string]any
}

type Hit struct {
	RuleID      string   `json:"rule_id"`
	RuleKey     string   `json:"rule_key"`
	Name        string   `json:"name"`
	Severity    string   `json:"severity"`
	Description string   `json:"description,omitempty"`
	Source      string   `json:"source,omitempty"`
	ExternalID  string   `json:"external_id,omitempty"`
	Revision    int64    `json:"revision,omitempty"`
	Stage       string   `json:"stage"`
	Evidence    []string `json:"evidence,omitempty"`
	Confirmed   bool     `json:"confirmed,omitempty"`
}

type Matcher struct {
	rules []compiledRule
}

// Count reports the number of compiled portable YARA-compatible rules.
func (m *Matcher) Count() int {
	if m == nil {
		return 0
	}
	return len(m.rules)
}

type compiledRule struct {
	rule     Rule
	patterns []compiledPattern
}

type compiledPattern struct {
	pattern Pattern
	regexp  *regexp.Regexp
}

func Compile(rules []Rule) (*Matcher, error) {
	if len(rules) > 2048 {
		return nil, errors.New("too many detection rules")
	}
	result := &Matcher{rules: make([]compiledRule, 0, len(rules))}
	for _, rule := range rules {
		if err := ValidateRule(rule); err != nil {
			return nil, fmt.Errorf("rule %q: %w", rule.Key, err)
		}
		compiled := compiledRule{rule: rule, patterns: make([]compiledPattern, 0, len(rule.Patterns))}
		for _, pattern := range rule.Patterns {
			item := compiledPattern{pattern: pattern}
			if pattern.Operator == PatternRegexp {
				expression := pattern.Value
				if pattern.NoCase {
					expression = "(?i:" + expression + ")"
				}
				item.regexp = regexp.MustCompile(expression)
			}
			compiled.patterns = append(compiled.patterns, item)
		}
		result.rules = append(result.rules, compiled)
	}
	return result, nil
}

func ValidateRule(rule Rule) error {
	if strings.TrimSpace(rule.Key) == "" || strings.TrimSpace(rule.Name) == "" {
		return errors.New("key and name are required")
	}
	if len(rule.Key) > 128 || len(rule.Name) > 255 || len(rule.Patterns) == 0 || len(rule.Patterns) > 64 {
		return errors.New("rule size is outside allowed limits")
	}
	for _, pattern := range rule.Patterns {
		if strings.TrimSpace(pattern.ID) == "" || len(pattern.ID) > 64 || len(pattern.Value) == 0 || len(pattern.Value) > 4096 {
			return errors.New("pattern id/value is invalid")
		}
		if !validField(pattern.Field) {
			return fmt.Errorf("unsupported field %q", pattern.Field)
		}
		if pattern.Operator != PatternContains && pattern.Operator != PatternRegexp {
			return fmt.Errorf("unsupported operator %q", pattern.Operator)
		}
		if pattern.MinCount < 0 || pattern.MinCount > 100 {
			return errors.New("min_count must be between 0 and 100")
		}
		if pattern.Operator == PatternRegexp {
			expression := pattern.Value
			if pattern.NoCase {
				expression = "(?i:" + expression + ")"
			}
			if _, err := regexp.Compile(expression); err != nil {
				return fmt.Errorf("invalid regular expression in %s: %w", pattern.ID, err)
			}
		}
	}
	return nil
}

func validField(field string) bool {
	return field == "raw" || field == "method" || field == "path" || field == "headers" || field == "body" || field == "event_type" || field == "service" || strings.HasPrefix(field, "payload.")
}

func (m *Matcher) Match(event Event, stage string) []Hit {
	if m == nil {
		return nil
	}
	hits := make([]Hit, 0)
	for _, rule := range m.rules {
		evidence := make([]string, 0, len(rule.patterns))
		matched := true
		for _, pattern := range rule.patterns {
			value := fieldValue(event, pattern.pattern.Field)
			count := patternCount(pattern, value)
			minimum := pattern.pattern.MinCount
			if minimum < 1 {
				minimum = 1
			}
			if count < minimum {
				matched = false
				break
			}
			evidence = append(evidence, fmt.Sprintf("%s 命中 %d 次", pattern.pattern.ID, count))
		}
		if matched {
			hits = append(hits, Hit{RuleID: rule.rule.ID, RuleKey: rule.rule.Key, Name: rule.rule.Name, Severity: normalizeSeverity(rule.rule.Severity), Description: rule.rule.Description, Source: rule.rule.Source, ExternalID: rule.rule.ExternalID, Revision: rule.rule.Revision, Stage: stage, Evidence: evidence})
		}
	}
	return hits
}

func patternCount(pattern compiledPattern, value string) int {
	if value == "" {
		return 0
	}
	if pattern.regexp != nil {
		return len(pattern.regexp.FindAllStringIndex(value, -1))
	}
	needle := pattern.pattern.Value
	if pattern.pattern.NoCase {
		value, needle = strings.ToLower(value), strings.ToLower(needle)
	}
	return strings.Count(value, needle)
}

func fieldValue(event Event, field string) string {
	switch field {
	case "raw":
		if event.RawPacket != "" {
			return event.RawPacket
		}
		return canonicalPayload(event.Payload)
	case "event_type":
		return event.EventType
	case "service":
		return event.Service
	case "method", "path", "body":
		return stringify(event.Payload[field])
	case "headers":
		return canonicalValue(event.Payload["headers"])
	default:
		return stringify(payloadPath(event.Payload, strings.TrimPrefix(field, "payload.")))
	}
}

func canonicalPayload(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(canonicalValue(payload[key]))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func canonicalValue(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return stringify(value)
	}
	return string(data)
}

func stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}

func payloadPath(payload map[string]any, path string) any {
	var current any = payload
	for _, part := range strings.Split(path, ".") {
		values, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = values[part]
	}
	return current
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical", "严重":
		return "critical"
	case "high", "高危":
		return "high"
	case "low", "低危":
		return "low"
	case "info", "提示":
		return "info"
	default:
		return "medium"
	}
}

func SetMinCount(patterns []Pattern, id string, count int) bool {
	for index := range patterns {
		if patterns[index].ID == id {
			patterns[index].MinCount = count
			return true
		}
	}
	return false
}

func ParsePositiveInt(value string) (int, bool) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	return parsed, err == nil && parsed > 0
}
