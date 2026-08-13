package sense

import (
	"net"
	"sort"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxSessions = 4096

type Probe struct {
	Protocol   string
	SourceIP   string
	SourcePort int
	DestIP     string
	DestPort   int
}

type Detection struct {
	Probe
	FirstSeen     time.Time
	LastSeen      time.Time
	DistinctPorts int
	Attempts      int
	Ports         []int
}

type scanSession struct {
	first, last time.Time
	ports       map[int]struct{}
	attempts    int
	probe       Probe
}

type Detector struct {
	config      protocol.SenseConfig
	excluded    map[int]struct{}
	ignored     []*net.IPNet
	sessions    map[string]*scanSession
	cooldowns   map[string]time.Time
	lastCleanup time.Time
}

func NewDetector(config protocol.SenseConfig) (*Detector, error) {
	normalized, err := NormalizeConfig(config)
	if err != nil {
		return nil, err
	}
	detector := &Detector{config: normalized, excluded: map[int]struct{}{}, sessions: map[string]*scanSession{}, cooldowns: map[string]time.Time{}}
	for _, port := range normalized.ExcludedPorts {
		detector.excluded[port] = struct{}{}
	}
	for _, raw := range normalized.IgnoredCIDRs {
		_, network, _ := net.ParseCIDR(raw)
		detector.ignored = append(detector.ignored, network)
	}
	return detector, nil
}

func (d *Detector) Observe(now time.Time, probe Probe) *Detection {
	sourceIP := net.ParseIP(strings.TrimSpace(probe.SourceIP))
	if probe.DestPort < 1 || probe.DestPort > 65535 || sourceIP == nil {
		return nil
	}
	probe.SourceIP = sourceIP.String()
	if strings.TrimSpace(probe.DestIP) != "" {
		destIP := net.ParseIP(strings.TrimSpace(probe.DestIP))
		if destIP == nil {
			return nil
		}
		probe.DestIP = destIP.String()
	}
	if probe.Protocol == "tcp" && !d.config.TCPEnabled || probe.Protocol == "udp" && !d.config.UDPEnabled {
		return nil
	}
	if probe.Protocol != "tcp" && probe.Protocol != "udp" {
		return nil
	}
	if _, excluded := d.excluded[probe.DestPort]; excluded || d.ignoredSource(probe.SourceIP) {
		return nil
	}
	d.cleanup(now)
	key := probe.Protocol + "|" + probe.SourceIP
	if until := d.cooldowns[key]; now.Before(until) {
		return nil
	}
	delete(d.cooldowns, key)
	session := d.sessions[key]
	window := time.Duration(d.config.WindowSeconds) * time.Second
	if session == nil || now.Sub(session.first) > window || now.Sub(session.last) > window {
		if session == nil && len(d.sessions) >= maxSessions {
			d.dropOldest()
		}
		session = &scanSession{first: now, ports: map[int]struct{}{}, probe: probe}
		d.sessions[key] = session
	}
	session.last = now
	session.attempts++
	session.probe = probe
	session.ports[probe.DestPort] = struct{}{}
	if len(session.ports) < d.config.DistinctPorts {
		return nil
	}
	ports := make([]int, 0, len(session.ports))
	for port := range session.ports {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	detection := &Detection{
		Probe: session.probe, FirstSeen: session.first, LastSeen: session.last,
		DistinctPorts: len(ports), Attempts: session.attempts, Ports: ports,
	}
	delete(d.sessions, key)
	d.cooldowns[key] = now.Add(time.Duration(d.config.CooldownSeconds) * time.Second)
	return detection
}

func (d *Detector) ignoredSource(value string) bool {
	ip := net.ParseIP(value)
	for _, network := range d.ignored {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (d *Detector) cleanup(now time.Time) {
	if !d.lastCleanup.IsZero() && now.Sub(d.lastCleanup) < 30*time.Second {
		return
	}
	d.lastCleanup = now
	maxAge := time.Duration(d.config.WindowSeconds+d.config.CooldownSeconds) * time.Second
	for key, session := range d.sessions {
		if now.Sub(session.last) > maxAge {
			delete(d.sessions, key)
		}
	}
	for key, until := range d.cooldowns {
		if !now.Before(until) {
			delete(d.cooldowns, key)
		}
	}
}

func (d *Detector) dropOldest() {
	var oldestKey string
	var oldest time.Time
	for key, session := range d.sessions {
		if oldestKey == "" || session.last.Before(oldest) {
			oldestKey, oldest = key, session.last
		}
	}
	delete(d.sessions, oldestKey)
}
