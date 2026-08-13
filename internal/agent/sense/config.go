package sense

import (
	"errors"
	"net"
	"sort"
	"strings"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

var ErrUnsupported = errors.New("passive scan sensing is not supported on this platform")

func DefaultConfig() protocol.SenseConfig {
	return protocol.SenseConfig{
		TCPEnabled: true, UDPEnabled: true, DistinctPorts: 10,
		WindowSeconds: 10, CooldownSeconds: 60,
	}
}

func NormalizeConfig(value protocol.SenseConfig) (protocol.SenseConfig, error) {
	value.Interface = strings.TrimSpace(value.Interface)
	if len(value.Interface) > 64 {
		return value, errors.New("interface must not exceed 64 bytes")
	}
	if !value.TCPEnabled && !value.UDPEnabled {
		if value.Enabled {
			return value, errors.New("at least one of TCP or UDP sensing must be enabled")
		}
		value.TCPEnabled, value.UDPEnabled = true, true
	}
	if value.DistinctPorts == 0 {
		value.DistinctPorts = 10
	}
	if value.DistinctPorts < 3 || value.DistinctPorts > 1024 {
		return value, errors.New("distinct_ports must be between 3 and 1024")
	}
	if value.WindowSeconds == 0 {
		value.WindowSeconds = 10
	}
	if value.WindowSeconds < 1 || value.WindowSeconds > 300 {
		return value, errors.New("window_seconds must be between 1 and 300")
	}
	if value.CooldownSeconds == 0 {
		value.CooldownSeconds = 60
	}
	if value.CooldownSeconds < 1 || value.CooldownSeconds > 86400 {
		return value, errors.New("cooldown_seconds must be between 1 and 86400")
	}
	originalPorts := append([]int(nil), value.ExcludedPorts...)
	if len(originalPorts) > 4096 {
		return value, errors.New("excluded_ports must not contain more than 4096 entries")
	}
	ports := make(map[int]struct{}, len(originalPorts))
	value.ExcludedPorts = make([]int, 0, len(originalPorts))
	for _, port := range originalPorts {
		if port < 1 || port > 65535 {
			return value, errors.New("excluded_ports contains an invalid port")
		}
		ports[port] = struct{}{}
	}
	for port := range ports {
		value.ExcludedPorts = append(value.ExcludedPorts, port)
	}
	sort.Ints(value.ExcludedPorts)
	originalCIDRs := append([]string(nil), value.IgnoredCIDRs...)
	if len(originalCIDRs) > 256 {
		return value, errors.New("ignored_cidrs must not contain more than 256 entries")
	}
	ignored := make(map[string]struct{}, len(originalCIDRs))
	value.IgnoredCIDRs = make([]string, 0, len(originalCIDRs))
	for _, raw := range originalCIDRs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if len(raw) > 128 {
			return value, errors.New("ignored_cidrs contains an entry longer than 128 bytes")
		}
		_, network, err := net.ParseCIDR(raw)
		if err != nil {
			return value, errors.New("ignored_cidrs contains an invalid CIDR")
		}
		ignored[network.String()] = struct{}{}
	}
	for cidr := range ignored {
		value.IgnoredCIDRs = append(value.IgnoredCIDRs, cidr)
	}
	sort.Strings(value.IgnoredCIDRs)
	return value, nil
}
