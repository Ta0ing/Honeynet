package pots

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type Sink func(protocol.Event)

type Service interface {
	Start(context.Context, protocol.PotTarget, Sink) error
	Stop() error
}

type TLSProvider interface {
	TLSConfig() *tls.Config
}

// TemplateRootProvider is implemented by the Agent runtime provider. Keeping
// it separate from TLSProvider lets the native protocol services remain
// independent from the on-disk Web resource template pack.
type TemplateRootProvider interface {
	TemplateRoot() string
}

var ErrUnsupportedService = errors.New("service is not implemented by this agent")

var factories = map[string]func(TLSProvider) Service{
	"bacnet":        func(TLSProvider) Service { return &BACnetService{} },
	"coap":          func(TLSProvider) Service { return &CoAPService{} },
	"dns":           func(TLSProvider) Service { return &DNSService{} },
	"elasticsearch": func(TLSProvider) Service { return &ElasticsearchService{} },
	"ftp":           func(TLSProvider) Service { return &FTPService{} },
	"http":          func(TLSProvider) Service { return &HTTPService{} },
	"https":         func(provider TLSProvider) Service { return &HTTPService{secure: true, tlsProvider: provider} },
	"imap":          func(TLSProvider) Service { return &IMAPService{} },
	"imaps":         func(provider TLSProvider) Service { return &IMAPService{secure: true, tlsProvider: provider} },
	"kafka":         func(TLSProvider) Service { return &KafkaService{} },
	"ldap":          func(TLSProvider) Service { return &LDAPService{} },
	"ldaps":         func(provider TLSProvider) Service { return &LDAPService{secure: true, tlsProvider: provider} },
	"memcached":     func(TLSProvider) Service { return &MemcachedService{} },
	"modbus":        func(TLSProvider) Service { return &ModbusService{} },
	"mongodb":       func(TLSProvider) Service { return &MongoDBService{} },
	"mqtt":          func(TLSProvider) Service { return &MQTTService{} },
	"mssql":         func(TLSProvider) Service { return &MSSQLService{} },
	"mysql":         func(TLSProvider) Service { return &MySQLService{} },
	"oracle":        func(TLSProvider) Service { return &OracleService{} },
	"pop3":          func(TLSProvider) Service { return &POP3Service{} },
	"pop3s":         func(provider TLSProvider) Service { return &POP3Service{secure: true, tlsProvider: provider} },
	"postgresql":    func(TLSProvider) Service { return &PostgreSQLService{} },
	"rdp":           func(TLSProvider) Service { return &RDPService{} },
	"redis":         func(TLSProvider) Service { return &RedisService{} },
	"rtsp-camera":   func(TLSProvider) Service { return &RTSPService{} },
	"s7comm":        func(TLSProvider) Service { return &S7Service{} },
	"smb":           func(TLSProvider) Service { return &SMBService{} },
	"smtp":          func(TLSProvider) Service { return &SMTPService{} },
	"smtps":         func(provider TLSProvider) Service { return &SMTPService{secure: true, tlsProvider: provider} },
	"snmp":          func(TLSProvider) Service { return &SNMPService{} },
	"ssh":           func(TLSProvider) Service { return &SSHService{} },
	"telnet":        func(TLSProvider) Service { return &TelnetService{} },
	"tftp":          func(TLSProvider) Service { return &TFTPService{} },
	"vnc":           func(TLSProvider) Service { return &VNCService{} },
	"web-template":  func(TLSProvider) Service { return &TemplateHTTPService{} },
	"zookeeper":     func(TLSProvider) Service { return &ZooKeeperService{} },
}

func New(code string, providers ...TLSProvider) (Service, error) {
	factory := factories[code]
	if factory == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedService, code)
	}
	var provider TLSProvider
	if len(providers) > 0 {
		provider = providers[0]
	}
	return factory(provider), nil
}

func wrapTLSListener(listener net.Listener, provider TLSProvider) (net.Listener, error) {
	if provider == nil {
		_ = listener.Close()
		return nil, errors.New("node TLS certificate manager is unavailable")
	}
	config := provider.TLSConfig()
	if config == nil {
		_ = listener.Close()
		return nil, errors.New("node TLS certificate manager is unavailable")
	}
	return tls.NewListener(listener, config), nil
}

func SupportedCodes() []string {
	codes := make([]string, 0, len(factories))
	for code := range factories {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// SupportedCodesAt reports only services that can actually start with the
// supplied template pack. Native protocol services are always returned;
// Web resource Web services are returned only when their configured root
// and index files are present.
func SupportedCodesAt(templateRoot string) []string {
	available := availableWebTemplates(templateRoot)
	codes := make([]string, 0, len(factories))
	for code := range factories {
		if _, isTemplate := webTemplateCodes[code]; isTemplate && !available[code] {
			continue
		}
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

func listenAddress(target protocol.PotTarget) string {
	bind := "0.0.0.0"
	if value, ok := target.Config["bind"].(string); ok {
		if normalized, valid := normalizeBindIP(value); valid {
			bind = normalized
		}
	}
	return net.JoinHostPort(bind, strconv.Itoa(target.Port))
}

func normalizeBindIP(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		value = strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	}
	address := value
	if index := strings.LastIndex(address, "%"); index > 0 {
		address = address[:index]
		if net.ParseIP(address) == nil || !strings.Contains(address, ":") || strings.TrimSpace(value[index+1:]) == "" {
			return "", false
		}
		return value, true
	}
	if ip := net.ParseIP(address); ip != nil {
		return ip.String(), true
	}
	return "", false
}

func endpoint(address net.Addr) protocol.Endpoint {
	if address == nil {
		return protocol.Endpoint{}
	}
	host, portText, err := net.SplitHostPort(address.String())
	if err != nil {
		return protocol.Endpoint{IP: address.String()}
	}
	port, _ := strconv.Atoi(portText)
	return protocol.Endpoint{IP: canonicalEndpointIP(host), Port: port}
}

func canonicalEndpointIP(host string) string {
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if index := strings.LastIndex(host, "%"); index > 0 && strings.Contains(host[:index], ":") {
		host = host[:index]
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return host
}

func configString(config map[string]any, key, fallback string) string {
	if value, ok := config[key].(string); ok && value != "" {
		return value
	}
	return fallback
}
