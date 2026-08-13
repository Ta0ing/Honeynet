package detection

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ImportedRule struct {
	Rule            Rule   `json:"rule"`
	Original        string `json:"original_condition"`
	ValidationError string `json:"validation_error,omitempty"`
}

var (
	ruleStartPattern  = regexp.MustCompile(`^rule\s+([A-Za-z0-9_]+)\s*\{`)
	metaPattern       = regexp.MustCompile(`^([A-Za-z0-9_]+)\s*=\s*(.+?)\s*$`)
	stringPattern     = regexp.MustCompile(`^\$([A-Za-z0-9_]+)\s*=\s*(.+?)\s*$`)
	countPattern      = regexp.MustCompile(`#([A-Za-z0-9_]+)\s*>=\s*([0-9]+)`)
	namePrefixPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]{1,31}_(.+)$`)
)

// ImportRuleDirectory parses the supported Yara rule subset from a directory.
// Response templates and other active-content fixtures are never executed.
func ImportRuleDirectory(directory string) ([]ImportedRule, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	result := make([]ImportedRule, 0)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		items, err := importRuleFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		result = append(result, items...)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Rule.Key < result[j].Rule.Key })
	return result, nil
}

func importRuleFile(filename string) ([]ImportedRule, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []ImportedRule
	var current *ImportedRule
	section := ""
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if match := ruleStartPattern.FindStringSubmatch(line); len(match) == 2 {
			if current != nil {
				finishImportedRule(current)
				result = append(result, *current)
			}
			current = &ImportedRule{Rule: Rule{Key: "builtin:" + match[1], Name: match[1], Severity: "medium", Source: "builtin", ExternalID: match[1]}}
			section = ""
			continue
		}
		if current == nil {
			continue
		}
		switch line {
		case "meta:", "strings:", "condition:":
			section = strings.TrimSuffix(line, ":")
			continue
		case "}":
			finishImportedRule(current)
			result = append(result, *current)
			current = nil
			section = ""
			continue
		}
		switch section {
		case "meta":
			parseRuleMeta(&current.Rule, line)
		case "strings":
			pattern, err := parseRulePattern(line)
			if err != nil {
				current.ValidationError = err.Error()
				continue
			}
			current.Rule.Patterns = append(current.Rule.Patterns, pattern)
		case "condition":
			if current.Original != "" {
				current.Original += " "
			}
			current.Original += line
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if current != nil {
		finishImportedRule(current)
		result = append(result, *current)
	}
	return result, nil
}

func parseRuleMeta(rule *Rule, line string) {
	match := metaPattern.FindStringSubmatch(line)
	if len(match) != 3 {
		return
	}
	value := strings.TrimSpace(match[2])
	if strings.HasPrefix(value, `"`) {
		decoded, err := strconv.Unquote(value)
		if err == nil {
			value = decoded
		}
	}
	switch match[1] {
	case "id":
		rule.ExternalID = value
	case "name":
		if match := namePrefixPattern.FindStringSubmatch(value); len(match) == 2 {
			value = match[1]
		}
		rule.Name = value
	case "description", "feedback":
		if rule.Description == "" || match[1] == "feedback" {
			rule.Description = value
		}
	case "threat_level", "priority":
		rule.Severity = normalizeSeverity(value)
	}
}

func parseRulePattern(line string) (Pattern, error) {
	match := stringPattern.FindStringSubmatch(line)
	if len(match) != 3 {
		return Pattern{}, errors.New("不支持的规则字符串声明")
	}
	id, expression := match[1], strings.TrimSpace(match[2])
	pattern := Pattern{ID: id, Field: importedRuleField(id), Operator: PatternContains, MinCount: 1}
	if strings.HasSuffix(strings.ToLower(expression), " nocase") {
		pattern.NoCase = true
		expression = strings.TrimSpace(expression[:len(expression)-len(" nocase")])
	}
	if strings.HasPrefix(expression, `"`) {
		value, err := strconv.Unquote(expression)
		if err != nil {
			return Pattern{}, fmt.Errorf("invalid quoted value for %s: %w", id, err)
		}
		pattern.Value = value
		return pattern, nil
	}
	if strings.HasPrefix(expression, "/") && strings.HasSuffix(expression, "/") && len(expression) >= 2 {
		pattern.Operator = PatternRegexp
		pattern.Value = expression[1 : len(expression)-1]
		return pattern, nil
	}
	return Pattern{}, fmt.Errorf("unsupported value for %s", id)
}

func importedRuleField(id string) string {
	lower := strings.ToLower(id)
	switch {
	case lower == "method":
		return "method"
	case strings.HasPrefix(lower, "path"):
		return "path"
	case lower == "contenttype" || lower == "type":
		return "headers"
	default:
		return "raw"
	}
}

func finishImportedRule(item *ImportedRule) {
	item.Original = strings.TrimSpace(item.Original)
	condition := strings.Join(strings.Fields(item.Original), " ")
	if condition != "all of them" && !strings.HasPrefix(condition, "all of them and #") {
		item.ValidationError = "暂不支持该规则条件，已导入但不会自动启用"
	}
	for _, match := range countPattern.FindAllStringSubmatch(condition, -1) {
		count, ok := ParsePositiveInt(match[2])
		if !ok || !SetMinCount(item.Rule.Patterns, match[1], count) {
			item.ValidationError = "规则条件的计数约束无效"
		}
	}
	if err := ValidateRule(item.Rule); err != nil {
		item.ValidationError = err.Error()
	}
}
