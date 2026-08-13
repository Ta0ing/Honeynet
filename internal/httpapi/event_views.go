package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/store"
	"github.com/honeynet/honeynet/internal/threatintel"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const maskedEventCredentialValue = "••••••••"

// attackEventView keeps the persisted event fields intact and adds display-only
// node context. DstIP remains the address observed by the Agent and must never
// be replaced with a NAT/public address.
type attackEventView struct {
	store.AttackEvent
	NodeName      string `json:"node_name"`
	NodeAddress   string `json:"node_address"`
	NodePublicIP  string `json:"node_public_ip"`
	DisplayDstIP  string `json:"display_dst_ip"`
	ObservedDstIP string `json:"observed_dst_ip"`
	// SensitiveVisible is explicit response metadata. Callers can distinguish a
	// redacted forensic view from either policy-default raw evidence or an
	// operator-authorized reveal without inspecting payload values.
	SensitiveVisible          bool                     `json:"sensitive_visible"`
	EvidenceRedacted          bool                     `json:"evidence_redacted"`
	SensitiveRedactionEnabled bool                     `json:"sensitive_redaction_enabled"`
	SensitiveRevealAudited    bool                     `json:"sensitive_reveal_audited"`
	ThreatIntelligence        *eventThreatIntelligence `json:"threat_intelligence,omitempty"`
}

type eventThreatIntelligence struct {
	Matched   bool     `json:"matched"`
	Level     int      `json:"level"`
	Labels    []string `json:"labels"`
	Behaviors []string `json:"behaviors"`
	Tags      []string `json:"tags"`
}

type eventNodeView struct {
	ID       string
	Name     string
	IP       string
	PublicIP string
}

func presentAttackEvent(event store.AttackEvent, node eventNodeView) attackEventView {
	publicIP := strings.TrimSpace(node.PublicIP)
	selectedIP := strings.TrimSpace(node.IP)
	observedIP := strings.TrimSpace(event.DstIP)
	displayIP := publicIP
	if displayIP == "" {
		displayIP = selectedIP
	}
	if displayIP == "" {
		displayIP = observedIP
	}
	return attackEventView{
		AttackEvent:   event,
		NodeName:      strings.TrimSpace(node.Name),
		NodeAddress:   selectedIP,
		NodePublicIP:  publicIP,
		DisplayDstIP:  displayIP,
		ObservedDstIP: observedIP,
	}
}

// sensitiveEventKey normalizes common Agent/template naming variants before
// comparing keys. Matching is case-insensitive and separator-insensitive so
// nested values such as auth_response, Auth-Response and AUTHRESPONSE share
// the same policy.
func sensitiveEventKey(key string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, strings.TrimSpace(key))
}

func isCredentialEventKey(key string) bool {
	normalized := sensitiveEventKey(key)
	exact := map[string]struct{}{
		"password": {}, "passwd": {}, "passphrase": {}, "pwd": {},
		"secret": {}, "clientsecret": {}, "token": {}, "accesstoken": {},
		"refreshtoken": {}, "idtoken": {}, "apikey": {}, "authorization": {},
		"proxyauthorization": {}, "authresponse": {}, "digest": {},
		"ntresponse": {}, "community": {}, "cookie": {}, "setcookie": {},
		"cookies": {}, "sessioncookie": {}, "session": {}, "sessionid": {},
		"csrftoken": {}, "bearer": {}, "privatekey": {},
	}
	if _, exists := exact[normalized]; exists {
		return true
	}
	// Custom templates frequently prefix secret fields (db_password,
	// bearerToken). A suffix policy errs on the side of masking an ambiguous
	// normalized field rather than exposing a credential.
	for _, suffix := range []string{"password", "passwd", "passphrase", "secret", "token", "authorization", "cookie", "privatekey"} {
		if strings.HasSuffix(normalized, suffix) {
			return true
		}
	}
	return false
}

func isRawEvidenceEventKey(key string) bool {
	normalized := sensitiveEventKey(key)
	_, exists := map[string]struct{}{
		"raw": {}, "rawdata": {}, "rawpacket": {}, "packet": {}, "packetdata": {},
		"rawrequest": {}, "rawresponse": {}, "body": {}, "bodybase64": {},
		"requestbody": {}, "responsebody": {}, "payloadbase64": {},
		"content": {}, "filecontent": {}, "contentbase64": {},
		"headers": {}, "requestheaders": {}, "responseheaders": {},
	}[normalized]
	return exists || strings.HasPrefix(normalized, "raw") || strings.HasSuffix(normalized, "body") ||
		(strings.HasSuffix(normalized, "base64") && (strings.Contains(normalized, "body") || strings.Contains(normalized, "content") || strings.Contains(normalized, "payload") || strings.Contains(normalized, "packet")))
}

func evidenceBytes(value any) []byte {
	if text, ok := value.(string); ok {
		return []byte(text)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return encoded
}

func evidenceKind(value any) string {
	switch value.(type) {
	case string:
		return "text"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case json.Number, float64, float32, int, int64, uint64:
		return "number"
	case bool:
		return "boolean"
	default:
		return "unknown"
	}
}

// safeEvidenceSummary preserves useful chain-of-custody metadata without
// returning the original packet/body. Persisted MySQL/ClickHouse data is not
// changed; this value exists only in the response copy.
func safeEvidenceSummary(value any) any {
	if value == nil || value == "" {
		return value
	}
	raw := evidenceBytes(value)
	digest := sha256.Sum256(raw)
	return map[string]any{
		"redacted":    true,
		"kind":        evidenceKind(value),
		"byte_length": len(raw),
		"sha256":      hex.EncodeToString(digest[:]),
	}
}

func safeRawPacketSummary(raw string) string {
	if raw == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("[敏感原文已隐藏；原始长度=%d 字节；SHA-256=%s]", len([]byte(raw)), hex.EncodeToString(digest[:]))
}

func redactEventJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			switch {
			case isCredentialEventKey(key):
				if item == nil || item == "" {
					result[key] = item
				} else {
					result[key] = maskedEventCredentialValue
				}
			case isRawEvidenceEventKey(key):
				result[key] = safeEvidenceSummary(item)
			default:
				result[key] = redactEventJSONValue(item)
			}
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = redactEventJSONValue(item)
		}
		return result
	default:
		return value
	}
}

// redactAttackEvent returns a deep response copy. It never changes the
// persisted event or its backing JSON byte slices.
func redactAttackEvent(event store.AttackEvent) store.AttackEvent {
	redacted := event
	redacted.RawPacket = safeRawPacketSummary(event.RawPacket)
	redacted.Tags = append(datatypes.JSON(nil), event.Tags...)
	redacted.Detections = append(datatypes.JSON(nil), event.Detections...)
	if len(event.Payload) == 0 {
		redacted.Payload = datatypes.JSON(`{}`)
		return redacted
	}
	decoder := json.NewDecoder(bytes.NewReader(event.Payload))
	decoder.UseNumber()
	var document any
	if !json.Valid(event.Payload) || decoder.Decode(&document) != nil {
		// Invalid historical payloads must not become a default-view bypass.
		document = map[string]any{"invalid_payload": safeEvidenceSummary(string(event.Payload))}
	}
	encoded, err := json.Marshal(redactEventJSONValue(document))
	if err != nil {
		encoded = []byte(`{"redaction_error":true}`)
	}
	redacted.Payload = datatypes.JSON(encoded)
	return redacted
}

type eventDisclosure struct {
	includeSensitive bool
	redactionEnabled bool
	explicitReveal   bool
}

// eventDisclosureForRequest resolves the response policy before either MySQL
// or ClickHouse is queried. When redaction is disabled, raw evidence is the
// ordinary response for every authenticated role: include_sensitive is not
// required, does not elevate access and does not create a misleading reveal
// audit. When enabled, the established admin/operator explicit reveal flow is
// retained.
func (a *API) eventDisclosureForRequest(c *gin.Context) (eventDisclosure, bool) {
	if !a.cfg.RedactSensitiveEvents {
		return eventDisclosure{includeSensitive: true}, true
	}
	wants := c.Query("include_sensitive") == "1" || strings.EqualFold(c.Query("include_sensitive"), "true")
	if !wants {
		return eventDisclosure{redactionEnabled: true}, true
	}
	role := currentUser(c).Role
	permitted := role == "admin" || role == "operator"
	return eventDisclosure{includeSensitive: permitted, redactionEnabled: true, explicitReveal: permitted}, permitted
}

func (a *API) auditEventReveal(c *gin.Context, eventID, scope string) error {
	user := currentUser(c)
	detail, _ := json.Marshal(map[string]any{
		"scope":      scope,
		"event_id":   eventID,
		"request_id": c.GetString("request_id"),
	})
	item := store.AuditLog{
		Base: store.NewBase(), UserID: user.ID, Username: user.Username,
		Action: "READ", Object: "/api/v1/events:sensitive", IP: c.ClientIP(),
		Detail: datatypes.JSON(detail),
	}
	return a.db.Create(&item).Error
}

func loadEventViewsWithDisclosure(db *gorm.DB, events []store.AttackEvent, disclosure eventDisclosure) []attackEventView {
	views := make([]attackEventView, 0, len(events))
	if len(events) == 0 {
		return views
	}
	nodeIDs := make([]string, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	for _, event := range events {
		if event.NodeID == "" {
			continue
		}
		if _, exists := seen[event.NodeID]; exists {
			continue
		}
		seen[event.NodeID] = struct{}{}
		nodeIDs = append(nodeIDs, event.NodeID)
	}
	nodes := make([]eventNodeView, 0, len(nodeIDs))
	if len(nodeIDs) > 0 {
		// Deleted nodes are still useful historical context for retained events.
		_ = db.Unscoped().Model(&store.Node{}).Select("id", "name", "ip", "public_ip").Where("id IN ?", nodeIDs).Scan(&nodes).Error
	}
	byID := make(map[string]eventNodeView, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	for _, event := range events {
		if !disclosure.includeSensitive {
			event = redactAttackEvent(event)
		}
		view := presentAttackEvent(event, byID[event.NodeID])
		view.SensitiveVisible = disclosure.includeSensitive
		view.EvidenceRedacted = !disclosure.includeSensitive
		view.SensitiveRedactionEnabled = disclosure.redactionEnabled
		view.SensitiveRevealAudited = disclosure.explicitReveal
		views = append(views, view)
	}
	return views
}

func (a *API) enrichEventThreatIntelligence(views []attackEventView) []attackEventView {
	if a.threatIntel == nil || !a.threatIntel.Status().Loaded {
		return views
	}
	for index := range views {
		result, found := a.threatIntel.Lookup(views[index].SrcIP)
		if !found {
			continue
		}
		tags := threatintel.EventTags(result)
		views[index].ThreatIntelligence = &eventThreatIntelligence{Matched: true, Level: result.Level, Labels: result.Labels, Behaviors: result.Behaviors, Tags: tags}
		var existing []string
		_ = json.Unmarshal(views[index].Tags, &existing)
		existing = appendUniqueStrings(existing, tags...)
		views[index].Tags, _ = json.Marshal(existing)
	}
	return views
}

// realtimeEventView applies the configured default disclosure policy to
// console WebSocket messages. A live push is never classified as an explicit
// reveal: with redaction enabled it is masked, and with redaction disabled it
// contains the same default raw evidence as the list/detail endpoints.
func (a *API) realtimeEventView(event store.AttackEvent) attackEventView {
	disclosure := eventDisclosure{includeSensitive: !a.cfg.RedactSensitiveEvents, redactionEnabled: a.cfg.RedactSensitiveEvents}
	views := a.enrichEventThreatIntelligence(loadEventViewsWithDisclosure(a.db, []store.AttackEvent{event}, disclosure))
	if len(views) != 0 {
		return views[0]
	}
	if disclosure.includeSensitive {
		view := presentAttackEvent(event, eventNodeView{})
		view.SensitiveVisible = true
		return view
	}
	view := presentAttackEvent(redactAttackEvent(event), eventNodeView{})
	view.EvidenceRedacted = true
	view.SensitiveRedactionEnabled = true
	return view
}
