package geoip

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	ipdb "github.com/ipipdotnet/ipdb-go"
)

var ErrInvalidAddress = errors.New("invalid IP address")

type cityDatabase interface {
	FindInfo(string, string) (*ipdb.CityInfo, error)
}

// Result is the Server-authoritative location attached to an attack event.
// Geo is deliberately presentation-ready because the current event schema
// stores one compact location string.
type Result struct {
	Geo      string
	Country  string
	Region   string
	City     string
	District string
	ASN      string
	ISP      string
	Internal bool
}

// Resolver keeps the IPIP database in memory. Lookup is guarded so a future
// online Reload can be introduced without racing event ingestion.
type Resolver struct {
	mu        sync.RWMutex
	db        cityDatabase
	language  string
	path      string
	buildTime time.Time
	languages []string
	fields    []string
}

// Open loads an IPIP City .ipdb file. An empty path explicitly disables
// enrichment and returns a nil resolver without error.
func Open(path, language string) (*Resolver, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	language = strings.ToUpper(strings.TrimSpace(language))
	if language == "" {
		language = "CN"
	}
	db, err := ipdb.NewCity(path)
	if err != nil {
		return nil, fmt.Errorf("load IPIP database %q: %w", path, err)
	}
	languages := db.Languages()
	if !containsFold(languages, language) {
		return nil, fmt.Errorf("IPIP database %q does not support language %q (available: %s)", path, language, strings.Join(languages, ", "))
	}
	return &Resolver{
		db: db, language: language, path: path, buildTime: db.BuildTime(),
		languages: append([]string(nil), languages...), fields: append([]string(nil), db.Fields()...),
	}, nil
}

func (r *Resolver) Lookup(address string) (Result, error) {
	ip := net.ParseIP(strings.TrimSpace(address))
	if ip == nil {
		return Result{}, fmt.Errorf("%w: %q", ErrInvalidAddress, address)
	}
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return Result{Geo: "内网地址", Internal: true}, nil
	}
	if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalMulticast() || !ip.IsGlobalUnicast() {
		return Result{Geo: "保留地址", Internal: true}, nil
	}

	r.mu.RLock()
	info, err := r.db.FindInfo(ip.String(), r.language)
	r.mu.RUnlock()
	if err != nil {
		return Result{}, fmt.Errorf("look up %s in IPIP database: %w", ip, err)
	}
	if info == nil {
		return Result{}, fmt.Errorf("look up %s in IPIP database: empty result", ip)
	}
	result := Result{
		Country:  strings.TrimSpace(info.CountryName),
		Region:   strings.TrimSpace(info.RegionName),
		City:     strings.TrimSpace(info.CityName),
		District: strings.TrimSpace(info.DistrictName),
		ASN:      normalizeASN(info.ASN),
		ISP:      strings.TrimSpace(info.IspDomain),
	}
	result.Geo = joinDistinct(result.Country, result.Region, result.City, result.District)
	if result.Geo == "" {
		return Result{}, fmt.Errorf("look up %s in IPIP database: location fields are empty", ip)
	}
	return result, nil
}

func (r *Resolver) Path() string { return r.path }

func (r *Resolver) BuildTime() time.Time { return r.buildTime }

func (r *Resolver) Languages() []string { return append([]string(nil), r.languages...) }

func (r *Resolver) Fields() []string { return append([]string(nil), r.fields...) }

func joinDistinct(values ...string) string {
	parts := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || value == "0" || seen[value] {
			continue
		}
		seen[value] = true
		parts = append(parts, value)
	}
	return strings.Join(parts, " / ")
}

func normalizeASN(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "AS") {
		return ""
	}
	if strings.HasPrefix(strings.ToUpper(value), "AS") {
		return "AS" + strings.TrimSpace(value[2:])
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return value
		}
	}
	return "AS" + value
}

func containsFold(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), want) {
			return true
		}
	}
	return false
}
