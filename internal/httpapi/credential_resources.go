package httpapi

import (
	"encoding/json"
	"net"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

type credentialResource struct {
	EventID      string    `json:"event_id"`
	NodeID       string    `json:"node_id"`
	EventType    string    `json:"event_type"`
	Timestamp    time.Time `json:"ts"`
	SourceIP     string    `json:"src_ip"`
	Geo          string    `json:"geo"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	AuthResponse string    `json:"auth_response"`
	Mechanism    string    `json:"mechanism"`
	ServiceCode  string    `json:"service_code"`
	ServiceName  string    `json:"service_name"`
}

type credentialTopValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

const maskedCredentialValue = "••••••••"

func maskCredentialValue(value string) string {
	if value == "" {
		return ""
	}
	return maskedCredentialValue
}

func redactCredentialResource(item credentialResource) credentialResource {
	item.Password = maskCredentialValue(item.Password)
	item.AuthResponse = maskCredentialValue(item.AuthResponse)
	return item
}

func redactAnalyticsCredentialCounts(items []analytics.CredentialCount) []analytics.CredentialCount {
	result := make([]analytics.CredentialCount, len(items))
	for index, item := range items {
		result[index] = analytics.CredentialCount{Value: maskCredentialValue(item.Value), Count: item.Count}
	}
	return result
}

func redactCredentialCounts(items []credentialTopValue) []credentialTopValue {
	result := make([]credentialTopValue, len(items))
	for index, item := range items {
		result[index] = credentialTopValue{Value: maskCredentialValue(item.Value), Count: item.Count}
	}
	return result
}

func wantsSensitiveCredentials(c *gin.Context) (requested, permitted bool) {
	wants := c.Query("include_sensitive") == "1" || strings.EqualFold(c.Query("include_sensitive"), "true")
	if !wants {
		return false, true
	}
	role := currentUser(c).Role
	return true, role == "admin" || role == "operator"
}

func (a *API) auditCredentialReveal(c *gin.Context) {
	user := currentUser(c)
	detail, _ := json.Marshal(map[string]any{
		"query_present": strings.TrimSpace(c.Query("q")) != "",
		"service":       strings.TrimSpace(c.Query("service")),
		"page":          c.Query("page"),
		"page_size":     c.Query("page_size"),
	})
	log := store.AuditLog{Base: store.Base{ID: uuid.NewString()}, UserID: user.ID, Username: user.Username, Action: "READ", Object: "/api/v1/credential-resources:sensitive", IP: c.ClientIP(), Detail: datatypes.JSON(detail)}
	_ = a.db.Create(&log).Error
}

var credentialPlaceholders = map[string]struct{}{
	"-": {}, "--": {}, "n/a": {}, "na": {}, "nil": {}, "null": {}, "none": {},
	"undefined": {}, "unknown": {}, "<nil>": {}, "[object object]": {}, "{}": {}, "[]": {},
	"请输入用户名": {}, "请输入密码": {}, "请输入账号": {}, "请输入口令": {},
	"用户名": {}, "密码": {}, "账号": {}, "口令": {}, "******": {}, "********": {},
}

func cleanCredentialValue(value string, maxBytes int) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxBytes || !utf8.ValidString(value) {
		return ""
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return ""
		}
	}
	lower := strings.ToLower(value)
	if _, exists := credentialPlaceholders[lower]; exists {
		return ""
	}
	if (strings.HasPrefix(lower, "{{") && strings.HasSuffix(lower, "}}")) ||
		(strings.HasPrefix(lower, "${") && strings.HasSuffix(lower, "}")) {
		return ""
	}
	if strings.Contains(lower, "<!doctype") || strings.Contains(lower, "<html") || strings.Contains(lower, "<script") {
		return ""
	}
	if (strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[")) && json.Valid([]byte(value)) {
		return ""
	}
	return value
}

func credentialPayloadString(payload map[string]any, maxBytes int, keys ...string) (string, string) {
	for _, key := range keys {
		value, ok := payload[key].(string)
		if !ok {
			continue
		}
		if cleaned := cleanCredentialValue(value, maxBytes); cleaned != "" {
			return cleaned, key
		}
	}
	return "", ""
}

func credentialMechanism(payload map[string]any, responseKind string) string {
	if value, _ := credentialPayloadString(payload, 256, "mechanism"); value != "" {
		return value
	}
	switch responseKind {
	case "auth_response":
		return "认证响应"
	case "digest":
		return "摘要认证"
	case "nt_response":
		return "NTLM"
	case "response":
		return "挑战响应"
	case "community":
		return "SNMP Community"
	default:
		return "未知认证方式"
	}
}

func credentialServiceName(code string, names map[string]string) string {
	if value := strings.TrimSpace(names[code]); value != "" {
		return value
	}
	fallbacks := map[string]string{
		"http": "通用 HTTP", "https": "通用 HTTPS", "ssh": "SSH 服务", "telnet": "Telnet 服务",
		"ftp": "FTP 服务", "smtp": "SMTP 邮件服务", "smtps": "SMTPS 邮件服务",
		"pop3": "POP3 邮件服务", "pop3s": "POP3S 邮件服务", "imap": "IMAP 邮件服务",
		"imaps": "IMAPS 邮件服务", "mysql": "MySQL 数据库", "mssql": "Microsoft SQL Server",
		"postgresql": "PostgreSQL 数据库", "redis": "Redis", "mongodb": "MongoDB",
		"elasticsearch": "Elasticsearch", "ldap": "LDAP 目录服务", "ldaps": "LDAPS 目录服务",
		"smb": "SMB 文件共享", "rdp": "Windows RDP", "vnc": "VNC 远程桌面",
		"mqtt": "MQTT Broker", "snmp": "SNMP 服务", "rtsp-camera": "RTSP 摄像头",
		"web-template": "自定义 Web 模板",
	}
	if value := fallbacks[code]; value != "" {
		return value
	}
	return "未知服务"
}

func buildCredentialResource(event store.AttackEvent, serviceNames map[string]string) (credentialResource, bool) {
	if strings.HasPrefix(event.EventType, "decoy.") || event.Service == "decoy" || net.ParseIP(strings.TrimSpace(event.SrcIP)) == nil {
		return credentialResource{}, false
	}
	payload := map[string]any{}
	if len(event.Payload) == 0 || json.Unmarshal(event.Payload, &payload) != nil {
		return credentialResource{}, false
	}
	username, _ := credentialPayloadString(payload, 512, "username", "user")
	password, _ := credentialPayloadString(payload, 4096, "password")
	authResponse, responseKind := credentialPayloadString(payload, 16<<10, "auth_response", "digest", "nt_response", "response", "community")
	if username == "" && password == "" && authResponse == "" {
		return credentialResource{}, false
	}
	serviceCode := strings.TrimSpace(event.Service)
	if serviceCode == "" {
		serviceCode, _, _ = strings.Cut(event.EventType, ".")
	}
	return credentialResource{
		EventID: event.EventID, NodeID: event.NodeID, EventType: event.EventType, Timestamp: event.Timestamp,
		SourceIP: event.SrcIP, Geo: event.Geo, Username: username, Password: password,
		AuthResponse: authResponse, Mechanism: credentialMechanism(payload, responseKind),
		ServiceCode: serviceCode, ServiceName: credentialServiceName(serviceCode, serviceNames),
	}, true
}

func credentialResourceMatches(item credentialResource, keyword string, includeSensitive bool) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	values := []string{item.SourceIP, item.Geo, item.Username, item.Mechanism, item.ServiceCode, item.ServiceName, item.EventType}
	if includeSensitive {
		values = append(values, item.Password, item.AuthResponse)
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), keyword) {
			return true
		}
	}
	return false
}

func topCredentialValues(items []credentialResource, selectValue func(credentialResource) string) []credentialTopValue {
	counts := map[string]int{}
	for _, item := range items {
		if value := selectValue(item); value != "" {
			counts[value]++
		}
	}
	return topCredentialCounts(counts)
}

func topCredentialCounts(counts map[string]int) []credentialTopValue {
	result := make([]credentialTopValue, 0, len(counts))
	for value, count := range counts {
		result = append(result, credentialTopValue{Value: value, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Value < result[j].Value
		}
		return result[i].Count > result[j].Count
	})
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}

func paginateCredentialResources(items []credentialResource, pageNumber, pageSize int) []credentialResource {
	start := (pageNumber - 1) * pageSize
	if start >= len(items) {
		return []credentialResource{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (a *API) listCredentialResources(c *gin.Context) {
	includeSensitive, allowed := wantsSensitiveCredentials(c)
	if !allowed {
		fail(c, 403, "SENSITIVE_CREDENTIALS_FORBIDDEN", "当前账号无权查看敏感认证数据")
		return
	}
	if includeSensitive {
		a.auditCredentialReveal(c)
	}
	pageNumber, pageSize := page(c)
	keyword := c.Query("q")
	if keyword == "" {
		keyword = c.Query("keyword")
	}
	var serviceRows []struct {
		Code string
		Name string
	}
	if err := a.db.Model(&store.PotService{}).Select("code", "name").Scan(&serviceRows).Error; err != nil {
		fail(c, 500, "QUERY_FAILED", "查询服务目录失败")
		return
	}
	serviceNames := make(map[string]string, len(serviceRows))
	for _, service := range serviceRows {
		serviceNames[service.Code] = service.Name
	}
	if a.analytics != nil {
		result, err := a.analytics.ListCredentials(c.Request.Context(), analytics.CredentialFilter{
			Keyword: keyword, Service: c.Query("service"), Limit: pageSize, Offset: (pageNumber - 1) * pageSize, IncludeSensitive: includeSensitive,
		})
		if err != nil {
			fail(c, 503, "ANALYTICS_UNAVAILABLE", "安全分析服务的账号资源查询暂不可用")
			return
		}
		items := make([]credentialResource, 0, len(result.Items))
		for _, item := range result.Items {
			resource := credentialResource{
				EventID: item.EventID, NodeID: item.NodeID, EventType: item.EventType, Timestamp: item.EventTime,
				SourceIP: item.SourceIP, Geo: item.Geo, Username: item.Username, Password: item.Password,
				AuthResponse: item.AuthResponse, Mechanism: item.Mechanism, ServiceCode: item.Service,
				ServiceName: credentialServiceName(item.Service, serviceNames),
			}
			if !includeSensitive {
				resource = redactCredentialResource(resource)
			}
			items = append(items, resource)
		}
		response := pageResult(items, int64(result.Total), pageNumber, pageSize)
		response["top_usernames"] = result.TopUsernames
		if includeSensitive {
			response["top_passwords"] = result.TopPasswords
		} else {
			response["top_passwords"] = redactAnalyticsCredentialCounts(result.TopPasswords)
		}
		response["sensitive_visible"] = includeSensitive
		ok(c, response)
		return
	}

	query := a.db.Model(&store.AttackEvent{}).
		Where("event_type NOT LIKE ? AND service <> ?", "decoy.%", "decoy").
		Where("(JSON_SEARCH(tags, 'one', ?) IS NOT NULL OR event_type LIKE ? OR event_type LIKE ? OR event_type LIKE ? OR event_type LIKE ?)", "credential", "%.credential", "%.authentication", "%.username", "%.community").
		Order("ts DESC")
	rows, err := query.Rows()
	if err != nil {
		fail(c, 500, "QUERY_FAILED", "查询账号资源失败")
		return
	}
	defer rows.Close()
	start, end := (pageNumber-1)*pageSize, pageNumber*pageSize
	items := make([]credentialResource, 0, pageSize)
	usernameCounts, passwordCounts := map[string]int{}, map[string]int{}
	total := 0
	for rows.Next() {
		var event store.AttackEvent
		a.db.ScanRows(rows, &event)
		item, valid := buildCredentialResource(event, serviceNames)
		if !valid || !credentialResourceMatches(item, keyword, includeSensitive) {
			continue
		}
		if item.Username != "" {
			usernameCounts[item.Username]++
		}
		if item.Password != "" {
			passwordCounts[item.Password]++
		}
		if total >= start && total < end {
			view := item
			if !includeSensitive {
				view = redactCredentialResource(view)
			}
			items = append(items, view)
		}
		total++
	}
	if err := rows.Err(); err != nil {
		fail(c, 500, "QUERY_FAILED", "读取账号资源失败")
		return
	}
	result := pageResult(items, int64(total), pageNumber, pageSize)
	result["top_usernames"] = topCredentialCounts(usernameCounts)
	passwords := topCredentialCounts(passwordCounts)
	if !includeSensitive {
		passwords = redactCredentialCounts(passwords)
	}
	result["top_passwords"] = passwords
	result["sensitive_visible"] = includeSensitive
	ok(c, result)
}
