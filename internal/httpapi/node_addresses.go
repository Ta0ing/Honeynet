package httpapi

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	nodeAddressAuto    = "auto"
	nodeAddressPublic  = "public"
	nodeAddressPrivate = "private"
	nodeAddressCustom  = "custom"
)

var carrierGradeNAT = func() *net.IPNet {
	_, block, _ := net.ParseCIDR("100.64.0.0/10")
	return block
}()

func normalizeNodeAddressMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case nodeAddressPublic:
		return nodeAddressPublic
	case nodeAddressPrivate:
		return nodeAddressPrivate
	case nodeAddressCustom:
		return nodeAddressCustom
	default:
		return nodeAddressAuto
	}
}

func canonicalNodeIP(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return ""
	}
	return ip.String()
}

func privateNodeIP(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil && ip.IsGlobalUnicast() && (ip.IsPrivate() || carrierGradeNAT.Contains(ip))
}

func publicNodeIP(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !carrierGradeNAT.Contains(ip)
}

func uniqueNodeIPs(values []string, accept func(string) bool) []string {
	items := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		ip := canonicalNodeIP(value)
		if ip == "" || seen[ip] || !accept(ip) {
			continue
		}
		seen[ip] = true
		items = append(items, ip)
	}
	return items
}

func decodeNodeIPsBy(raw datatypes.JSON, accept func(string) bool) []string {
	var values []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &values)
	}
	return uniqueNodeIPs(values, accept)
}

// decodeNodeIPs retains the historical helper name for private address
// candidates. Public candidates use the explicit companion below.
func decodeNodeIPs(raw datatypes.JSON) []string {
	return decodeNodeIPsBy(raw, privateNodeIP)
}

func decodePublicNodeIPs(raw datatypes.JSON) []string {
	return decodeNodeIPsBy(raw, publicNodeIP)
}

func encodeNodeIPs(values []string) datatypes.JSON {
	raw, _ := json.Marshal(values)
	return datatypes.JSON(raw)
}

func appendUniqueNodeIP(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append(values, value)
}

func prependUniqueNodeIP(values []string, value string) []string {
	if value == "" {
		return values
	}
	result := []string{value}
	for _, item := range values {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

func requestRemoteIP(request *http.Request) string {
	if request == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(request.RemoteAddr))
	if err != nil {
		host = strings.Trim(strings.TrimSpace(request.RemoteAddr), "[]")
	}
	return canonicalNodeIP(host)
}

func mergeNodeAddressReport(node *store.Node, observed string, reported []string) {
	mode := normalizeNodeAddressMode(node.AddressMode)
	privateIPs := decodeNodeIPs(node.PrivateIPs)
	publicIPs := decodePublicNodeIPs(node.PublicIPs)
	legacyPublicIP := canonicalNodeIP(node.PublicIP)
	if publicNodeIP(legacyPublicIP) {
		publicIPs = appendUniqueNodeIP(publicIPs, legacyPublicIP)
	}
	if reported != nil {
		privateIPs = uniqueNodeIPs(reported, privateNodeIP)
		publicIPs = uniqueNodeIPs(reported, publicNodeIP)
		// The observed NAT address may not be present on a local interface.
		// Retain it until the next connection supplies a newer observed peer.
		if publicNodeIP(legacyPublicIP) {
			publicIPs = appendUniqueNodeIP(publicIPs, legacyPublicIP)
		}
	}
	observed = canonicalNodeIP(observed)
	if privateNodeIP(observed) {
		privateIPs = appendUniqueNodeIP(privateIPs, observed)
	}
	if publicNodeIP(observed) {
		publicIPs = prependUniqueNodeIP(publicIPs, observed)
	}

	selected := canonicalNodeIP(node.IP)
	publicIP := legacyPublicIP
	switch mode {
	case nodeAddressAuto:
		if publicNodeIP(observed) {
			publicIP = observed
		} else if !publicNodeIP(publicIP) && len(publicIPs) > 0 {
			publicIP = publicIPs[0]
		}
		if publicIP != "" {
			publicIPs = prependUniqueNodeIP(publicIPs, publicIP)
			selected = publicIP
		} else if len(privateIPs) > 0 {
			selected = privateIPs[0]
		}
	case nodeAddressPublic:
		if publicNodeIP(selected) {
			publicIPs = prependUniqueNodeIP(publicIPs, selected)
			publicIP = selected
		} else if len(publicIPs) > 0 {
			selected, publicIP = publicIPs[0], publicIPs[0]
		}
	case nodeAddressPrivate:
		if !privateNodeIP(selected) && len(privateIPs) > 0 {
			selected = privateIPs[0]
		}
	case nodeAddressCustom:
		// A custom selection is intentionally never overwritten by a heartbeat.
	}

	node.AddressMode = mode
	node.PublicIP = publicIP
	node.PublicIPs = encodeNodeIPs(publicIPs)
	node.PrivateIPs = encodeNodeIPs(privateIPs)
	node.IP = selected
}

func updateNodeAddressReport(db *gorm.DB, nodeID, observed string, reported []string) (store.Node, error) {
	var node store.Node
	if err := db.First(&node, "id = ?", nodeID).Error; err != nil {
		return node, err
	}
	mergeNodeAddressReport(&node, observed, reported)
	err := db.Model(&store.Node{}).Where("id = ?", nodeID).Updates(map[string]any{
		"address_mode": node.AddressMode,
		"public_ip":    node.PublicIP,
		"public_ips":   node.PublicIPs,
		"private_ips":  node.PrivateIPs,
		"ip":           node.IP,
	}).Error
	return node, err
}

func configureNodeAddress(node *store.Node, mode, selected string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = nodeAddressAuto
	}
	if mode != nodeAddressAuto && mode != nodeAddressPublic && mode != nodeAddressPrivate && mode != nodeAddressCustom {
		return errors.New("节点地址策略无效")
	}
	selected = canonicalNodeIP(selected)
	privateIPs := decodeNodeIPs(node.PrivateIPs)
	publicIPs := decodePublicNodeIPs(node.PublicIPs)
	if publicNodeIP(node.PublicIP) {
		publicIPs = appendUniqueNodeIP(publicIPs, canonicalNodeIP(node.PublicIP))
	}
	switch mode {
	case nodeAddressAuto:
		node.AddressMode = mode
		mergeNodeAddressReport(node, "", nil)
		return nil
	case nodeAddressPublic:
		if selected == "" {
			selected = canonicalNodeIP(node.PublicIP)
			if selected == "" && len(publicIPs) > 0 {
				selected = publicIPs[0]
			}
		}
		if !publicNodeIP(selected) {
			return errors.New("请选择有效的公网 IP 地址")
		}
		publicIPs = prependUniqueNodeIP(publicIPs, selected)
		node.PublicIP = selected
		node.PublicIPs = encodeNodeIPs(publicIPs)
	case nodeAddressPrivate:
		if selected == "" && len(privateIPs) > 0 {
			selected = privateIPs[0]
		}
		if !privateNodeIP(selected) {
			return errors.New("请选择有效的私网 IP 地址")
		}
		privateIPs = appendUniqueNodeIP(privateIPs, selected)
		node.PrivateIPs = encodeNodeIPs(privateIPs)
	case nodeAddressCustom:
		if selected == "" {
			return errors.New("请输入有效的 IP 地址")
		}
	}
	node.AddressMode = mode
	node.IP = selected
	return nil
}
