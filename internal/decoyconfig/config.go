package decoyconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

const (
	MaxConfigBytes  = 128 << 10
	MaxContentBytes = 64 << 10
)

type Spec struct {
	Path            string `json:"path,omitempty"`
	Content         string `json:"content,omitempty"`
	Marker          string `json:"marker,omitempty"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	Mode            string `json:"mode,omitempty"`
	CreateParent    bool   `json:"create_parent,omitempty"`
	MonitorExisting bool   `json:"monitor_existing,omitempty"`
	Token           string `json:"token,omitempty"`
	Description     string `json:"description,omitempty"`
}

func Parse(kind string, raw []byte) (Spec, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || len(trimmed) > MaxConfigBytes || trimmed[0] != '{' {
		return Spec{}, fmt.Errorf("config must be a JSON object no larger than %d bytes", MaxConfigBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var spec Spec
	if err := decoder.Decode(&spec); err != nil {
		return Spec{}, fmt.Errorf("decode decoy config: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Spec{}, errors.New("decoy config must contain exactly one JSON object")
	}
	if err := normalize(kind, &spec); err != nil {
		return Spec{}, err
	}
	return spec, nil
}

func ParseMap(kind string, config map[string]any) (Spec, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return Spec{}, err
	}
	return Parse(kind, raw)
}

func Canonical(spec Spec) []byte {
	raw, _ := json.Marshal(spec)
	return raw
}

func FileMode(spec Spec) (uint32, error) {
	value, err := strconv.ParseUint(spec.Mode, 8, 32)
	if err != nil {
		return 0, errors.New("mode must be an octal file mode")
	}
	return uint32(value), nil
}

func Render(kind, id, name string, spec Spec) []byte {
	if spec.Content != "" {
		return []byte(spec.Content)
	}
	marker := spec.Marker
	if marker == "" {
		marker = "honeynet-decoy-" + id
	}
	if kind == "credential" {
		return []byte(fmt.Sprintf("# %s\n# Managed deception credential; not a production secret.\nDB_USERNAME=%s\nDB_PASSWORD=%s\nHONEYNET_MARKER=%s\n", name, spec.Username, spec.Password, marker))
	}
	return []byte(fmt.Sprintf("%s\nHoneynet managed decoy document.\nmarker=%s\n", name, marker))
}

func normalize(kind string, spec *Spec) error {
	spec.Path = cleanPortablePath(strings.TrimSpace(spec.Path))
	spec.Marker = strings.TrimSpace(spec.Marker)
	spec.Username = strings.TrimSpace(spec.Username)
	spec.Password = strings.TrimSpace(spec.Password)
	spec.Mode = strings.TrimSpace(spec.Mode)
	spec.Token = strings.TrimSpace(spec.Token)
	spec.Description = strings.TrimSpace(spec.Description)
	if len(spec.Content) > MaxContentBytes {
		return fmt.Errorf("content exceeds %d bytes", MaxContentBytes)
	}
	if len(spec.Marker) > 128 || len(spec.Username) > 256 || len(spec.Password) > 256 || len(spec.Description) > 512 {
		return errors.New("one or more decoy fields exceed their maximum length")
	}
	switch kind {
	case "file":
		if err := validateFileSpec(spec); err != nil {
			return err
		}
		if spec.Username != "" || spec.Password != "" || spec.Token != "" || spec.Description != "" {
			return errors.New("file decoys do not accept credential or network fields")
		}
		if spec.Mode == "" {
			spec.Mode = "0644"
		}
	case "credential":
		if err := validateFileSpec(spec); err != nil {
			return err
		}
		if spec.Username == "" || spec.Password == "" {
			return errors.New("credential decoys require username and password")
		}
		if spec.Token != "" || spec.Description != "" {
			return errors.New("credential decoys do not accept network fields")
		}
		if spec.Mode == "" {
			spec.Mode = "0600"
		}
	case "network":
		if spec.Token == "" || len(spec.Token) < 8 || len(spec.Token) > 128 || !safeToken(spec.Token) {
			return errors.New("network decoys require an 8-128 character token containing only letters, digits, '.', '_', ':' or '-'")
		}
		if spec.Path != "" || spec.Content != "" || spec.Marker != "" || spec.Username != "" || spec.Password != "" || spec.Mode != "" || spec.CreateParent || spec.MonitorExisting {
			return errors.New("network decoys only accept token and description")
		}
		spec.Path = ""
	default:
		return errors.New("unsupported decoy type")
	}
	if kind != "network" {
		mode, err := FileMode(*spec)
		if err != nil || mode != 0400 && mode != 0440 && mode != 0600 && mode != 0640 && mode != 0644 {
			return errors.New("mode must be one of 0400, 0440, 0600, 0640 or 0644")
		}
	}
	return nil
}

func validateFileSpec(spec *Spec) error {
	if spec.Path == "." || !portableAbsolute(spec.Path) || portableRoot(spec.Path) || len(spec.Path) > 4096 || strings.ContainsRune(spec.Path, 0) {
		return errors.New("path must be a non-root absolute file path")
	}
	return nil
}

func cleanPortablePath(value string) string {
	if strings.HasPrefix(value, "/") {
		return path.Clean(value)
	}
	if windowsAbsolute(value) {
		return strings.ReplaceAll(value, "/", `\`)
	}
	return value
}

func portableAbsolute(value string) bool {
	return strings.HasPrefix(value, "/") || windowsAbsolute(value)
}

func windowsAbsolute(value string) bool {
	if len(value) >= 3 && ((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) && value[1] == ':' && (value[2] == '\\' || value[2] == '/') {
		return true
	}
	return strings.HasPrefix(value, `\\`) && len(strings.FieldsFunc(value[2:], func(char rune) bool { return char == '\\' || char == '/' })) >= 2
}

func portableRoot(value string) bool {
	if value == "/" {
		return true
	}
	if len(value) == 3 && value[1] == ':' && (value[2] == '\\' || value[2] == '/') {
		return true
	}
	if strings.HasPrefix(value, `\\`) {
		parts := strings.FieldsFunc(value[2:], func(char rune) bool { return char == '\\' || char == '/' })
		return len(parts) <= 2
	}
	return false
}

func safeToken(value string) bool {
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("._:-", char) {
			continue
		}
		return false
	}
	return true
}
