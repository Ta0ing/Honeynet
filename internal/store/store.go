package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.DatabaseDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
}

func MigrateAndSeed(db *gorm.DB, cfg config.Config) error {
	if err := migrateLegacySenseColumns(db); err != nil {
		return err
	}
	if err := migrateLegacyAlertFingerprints(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(&User{}, &AuditLog{}, &PlatformOEMSetting{}, &Node{}, &NodeSenseConfig{}, &PotService{}, &PotTemplate{}, &PotInstance{}, &Decoy{}, &AttackEvent{}, &EventReceipt{}, &EventMigrationCheckpoint{}, &Alert{}, &AlertRule{}, &DetectionRule{}, &AIAnalysis{}, &AIHarnessRun{}, &DetectionRuleProposal{}, &DetectionRuleFeedback{}, &AISetting{}, &AlertChannel{}, &AlertDelivery{}, &AgentRelease{}, &AgentBuild{}, &UpgradeRollout{}, &UpgradeTask{}, &IOC{}, &IntelFeed{}); err != nil {
		return err
	}
	if err := migrateNodeAddressCandidates(db); err != nil {
		return err
	}
	var users int64
	if err := db.Model(&User{}).Count(&users).Error; err != nil {
		return err
	}
	if users == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 12)
		if err != nil {
			return err
		}
		user := User{Base: NewBase(), Username: cfg.AdminUsername, DisplayName: "系统管理员", PasswordHash: string(hash), Role: "admin", Enabled: true, TokenVersion: 1}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}
	if err := seedPlatformOEMSetting(db, cfg); err != nil {
		return err
	}
	if err := seedBuiltinNode(db, cfg); err != nil {
		return err
	}
	if err := seedNodeSenseConfigs(db); err != nil {
		return err
	}
	if err := seedPotServices(db); err != nil {
		return err
	}
	if err := seedBuiltinPots(db, cfg); err != nil {
		return err
	}
	return seedRules(db)
}

const PlatformOEMSettingID = "00000000-0000-4000-8200-000000000001"

func seedPlatformOEMSetting(db *gorm.DB, cfg config.Config) error {
	version := strings.TrimSpace(cfg.Version)
	if version == "" {
		version = "dev"
	}
	item := PlatformOEMSetting{
		Base: Base{ID: PlatformOEMSettingID}, SystemName: "Honeynet",
		SystemVersion: version, Copyright: "© Honeynet",
		Revision: 1,
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&item).Error
}

func migrateNodeAddressCandidates(db *gorm.DB) error {
	var nodes []Node
	if err := db.Unscoped().Select("id", "public_ip", "public_ips", "private_ips").Find(&nodes).Error; err != nil {
		return err
	}
	for _, node := range nodes {
		updates := map[string]any{}
		if len(node.PublicIPs) == 0 || string(node.PublicIPs) == "null" {
			values := []string{}
			if value := strings.TrimSpace(node.PublicIP); value != "" {
				values = append(values, value)
			}
			raw, _ := json.Marshal(values)
			updates["public_ips"] = datatypes.JSON(raw)
		}
		if len(node.PrivateIPs) == 0 || string(node.PrivateIPs) == "null" {
			updates["private_ips"] = datatypes.JSON(`[]`)
		}
		if len(updates) > 0 {
			if err := db.Unscoped().Model(&Node{}).Where("id = ?", node.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateLegacyAlertFingerprints(db *gorm.DB) error {
	if !db.Migrator().HasTable(&Alert{}) || !db.Migrator().HasColumn(&Alert{}, "fingerprint") {
		return nil
	}
	// Historical releases indexed (but did not uniquely constrain)
	// fingerprint. Before AutoMigrate enforces event-level idempotency, give
	// duplicate legacy rows deterministic unique values while preserving the
	// first row's fingerprint for audit continuity.
	var duplicates []string
	if err := db.Model(&Alert{}).Select("fingerprint").Where("fingerprint <> ''").Group("fingerprint").Having("COUNT(*) > 1").Pluck("fingerprint", &duplicates).Error; err != nil {
		return err
	}
	for _, fingerprint := range duplicates {
		var alerts []Alert
		if err := db.Unscoped().Where("fingerprint = ?", fingerprint).Order("created_at ASC").Order("id ASC").Find(&alerts).Error; err != nil {
			return err
		}
		for index := 1; index < len(alerts); index++ {
			sum := sha256.Sum256([]byte(fingerprint + "|legacy|" + alerts[index].ID))
			if err := db.Unscoped().Model(&Alert{}).Where("id = ?", alerts[index].ID).UpdateColumn("fingerprint", hex.EncodeToString(sum[:])).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateLegacySenseColumns(db *gorm.DB) error {
	const table = "node_sense_configs"
	if !db.Migrator().HasTable(table) {
		return nil
	}
	hasLegacy := db.Migrator().HasColumn(table, "ignored_c_id_rs")
	if !hasLegacy {
		return nil
	}
	if !db.Migrator().HasColumn(table, "ignored_cidrs") {
		return db.Exec("ALTER TABLE node_sense_configs RENAME COLUMN ignored_c_id_rs TO ignored_cidrs").Error
	}
	if err := db.Exec("UPDATE node_sense_configs SET ignored_cidrs = ignored_c_id_rs").Error; err != nil {
		return err
	}
	return db.Exec("ALTER TABLE node_sense_configs DROP COLUMN ignored_c_id_rs").Error
}

const BuiltinNodeID = "00000000-0000-4000-8000-000000000001"

func seedBuiltinNode(db *gorm.DB, cfg config.Config) error {
	if cfg.BuiltinAgentToken == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(cfg.BuiltinAgentToken))
	hash := hex.EncodeToString(sum[:])
	expires := time.Now().Add(24 * time.Hour)
	node := Node{Base: Base{ID: BuiltinNodeID}, Name: "管理端内置节点", GroupName: "内置节点", Status: "offline", OS: "linux", Arch: "native", Labels: datatypes.JSON(`{"builtin":true}`), PublicIPs: datatypes.JSON(`[]`), PrivateIPs: datatypes.JSON(`[]`), RegistrationTokenHash: hash, TokenExpiresAt: &expires}
	var existing Node
	err := db.Unscoped().Where("id = ?", BuiltinNodeID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&node).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		node = existing
		if node.DeletedAt.Valid {
			return nil
		}
	}
	updates := map[string]any{"agent_token_hash": ""}
	if node.CertificateSerial == "" {
		updates["registration_token_hash"] = hash
		updates["token_expires_at"] = expires
	}
	return db.Model(&node).Updates(updates).Error
}

func seedBuiltinPots(db *gorm.DB, cfg config.Config) error {
	if cfg.BuiltinAgentToken == "" {
		return nil
	}
	items := []PotInstance{
		{Base: Base{ID: "00000000-0000-4000-8100-000000000001"}, NodeID: BuiltinNodeID, ServiceCode: "http", Name: "内置 HTTP 蜜罐", Port: 80, Config: datatypes.JSON(`{"title":"Internal OA Portal"}`), DesiredStatus: "running", ActualStatus: "pending"},
		{Base: Base{ID: "00000000-0000-4000-8100-000000000002"}, NodeID: BuiltinNodeID, ServiceCode: "ssh", Name: "内置 SSH 蜜罐", Port: 22, Config: datatypes.JSON(`{}`), DesiredStatus: "running", ActualStatus: "pending"},
		{Base: Base{ID: "00000000-0000-4000-8100-000000000003"}, NodeID: BuiltinNodeID, ServiceCode: "telnet", Name: "内置 Telnet 蜜罐", Port: 23, Config: datatypes.JSON(`{}`), DesiredStatus: "running", ActualStatus: "pending"},
		{Base: Base{ID: "00000000-0000-4000-8100-000000000004"}, NodeID: BuiltinNodeID, ServiceCode: "redis", Name: "内置 Redis 蜜罐", Port: 6379, Config: datatypes.JSON(`{}`), DesiredStatus: "running", ActualStatus: "pending"},
		{Base: Base{ID: "00000000-0000-4000-8100-000000000005"}, NodeID: BuiltinNodeID, ServiceCode: "mysql", Name: "内置 MySQL 蜜罐", Port: 3306, Config: datatypes.JSON(`{}`), DesiredStatus: "running", ActualStatus: "pending"},
	}
	for i := range items {
		var existing PotInstance
		err := db.Unscoped().Where("id = ?", items[i].ID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = db.Create(&items[i]).Error
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func NewBase() Base { return Base{ID: uuid.NewString()} }

func NewNodeSenseConfig(nodeID string) NodeSenseConfig {
	return NodeSenseConfig{
		Base: NewBase(), NodeID: nodeID, TCPEnabled: true, UDPEnabled: true,
		DistinctPorts: 10, WindowSeconds: 10, CooldownSeconds: 60,
		ExcludedPorts: datatypes.JSON("[]"), IgnoredCIDRs: datatypes.JSON("[]"),
		ActualStatus: "disabled",
	}
}

func EnsureNodeSenseConfig(db *gorm.DB, nodeID string) (NodeSenseConfig, error) {
	var item NodeSenseConfig
	if err := db.Where("node_id = ?", nodeID).First(&item).Error; err == nil {
		return item, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return item, err
	}
	item = NewNodeSenseConfig(nodeID)
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "node_id"}}, DoNothing: true}).Create(&item).Error; err != nil {
		return item, err
	}
	err := db.Where("node_id = ?", nodeID).First(&item).Error
	return item, err
}

func seedNodeSenseConfigs(db *gorm.DB) error {
	var nodeIDs []string
	if err := db.Model(&Node{}).Pluck("id", &nodeIDs).Error; err != nil {
		return err
	}
	for _, nodeID := range nodeIDs {
		if _, err := EnsureNodeSenseConfig(db, nodeID); err != nil {
			return err
		}
	}
	return nil
}

type serviceSeed struct {
	code, name, category, protocol string
	port                           int
	depth                          string
}

func potServiceSeeds() []serviceSeed {
	seeds := []serviceSeed{
		{"ssh", "SSH 服务", "基础网络", "tcp", 22, "中交互"}, {"telnet", "Telnet 服务", "基础网络", "tcp", 23, "中交互"}, {"ftp", "FTP 服务", "基础网络", "tcp", 21, "中交互"}, {"tftp", "TFTP 服务", "基础网络", "udp", 69, "低交互"}, {"smb", "SMB 文件共享", "基础网络", "tcp", 445, "低交互"}, {"rdp", "Windows RDP", "基础网络", "tcp", 3389, "低交互"}, {"vnc", "VNC 远程桌面", "基础网络", "tcp", 5900, "低交互"}, {"ldap", "LDAP 目录服务", "基础网络", "tcp", 389, "低交互"}, {"ldaps", "LDAPS 目录服务", "基础网络", "tcp", 636, "中交互"}, {"dns", "DNS 服务", "基础网络", "udp", 53, "中交互"}, {"snmp", "SNMP 服务", "基础网络", "udp", 161, "中交互"},
		{"http", "通用 HTTP", "Web服务器", "tcp", 80, "中交互"}, {"web-template", "自定义 Web 模板", "Web服务器", "tcp", 8080, "低交互"}, {"https", "通用 HTTPS", "Web服务器", "tcp", 443, "中交互"},
		{"mysql", "MySQL 数据库", "数据库/中间件", "tcp", 3306, "中交互"}, {"postgresql", "PostgreSQL 数据库", "数据库/中间件", "tcp", 5432, "中交互"}, {"mssql", "Microsoft SQL Server", "数据库/中间件", "tcp", 1433, "中交互"}, {"oracle", "Oracle 数据库", "数据库/中间件", "tcp", 1521, "低交互"}, {"redis", "Redis", "数据库/中间件", "tcp", 6379, "中交互"}, {"mongodb", "MongoDB", "数据库/中间件", "tcp", 27017, "中交互"}, {"elasticsearch", "Elasticsearch", "数据库/中间件", "tcp", 9200, "中交互"}, {"memcached", "Memcached", "数据库/中间件", "tcp", 11211, "中交互"}, {"zookeeper", "ZooKeeper", "数据库/中间件", "tcp", 2181, "中交互"}, {"kafka", "Kafka", "数据库/中间件", "tcp", 9092, "低交互"},
		{"cisco-router", "Cisco 路由器", "网络设备", "tcp", 23, "中交互"}, {"huawei-switch", "华为交换机", "网络设备", "tcp", 23, "中交互"}, {"h3c-switch", "H3C 交换机", "网络设备", "tcp", 23, "中交互"}, {"mikrotik", "MikroTik RouterOS", "网络设备", "tcp", 8291, "中交互"},
		{"smtp", "SMTP 邮件服务", "邮件系统", "tcp", 25, "中交互"}, {"smtps", "SMTPS 邮件服务", "邮件系统", "tcp", 465, "中交互"}, {"pop3", "POP3 邮件服务", "邮件系统", "tcp", 110, "中交互"}, {"pop3s", "POP3S 邮件服务", "邮件系统", "tcp", 995, "中交互"}, {"imap", "IMAP 邮件服务", "邮件系统", "tcp", 143, "中交互"}, {"imaps", "IMAPS 邮件服务", "邮件系统", "tcp", 993, "中交互"},
		{"rtsp-camera", "RTSP 摄像头", "IoT设备", "tcp", 554, "中交互"}, {"modbus", "Modbus TCP", "IoT设备", "tcp", 502, "中交互"}, {"s7comm", "Siemens S7", "IoT设备", "tcp", 102, "中交互"}, {"mqtt", "MQTT Broker", "IoT设备", "tcp", 1883, "中交互"}, {"coap", "CoAP", "IoT设备", "udp", 5683, "中交互"}, {"bacnet", "BACnet", "IoT设备", "udp", 47808, "低交互"}, {"upnp", "UPnP", "IoT设备", "udp", 1900, "中交互"},
		{"dubbo", "Dubbo", "数据库/中间件", "tcp", 20880, "低交互"}, {"rocketmq", "RocketMQ", "数据库/中间件", "tcp", 9876, "低交互"}, {"cassandra", "Cassandra", "数据库/中间件", "tcp", 9042, "低交互"},
	}
	return append(seeds, webTemplateServiceSeeds...)
}

var webTemplateServiceSeeds = []serviceSeed{
	{"ac-sangfor", "深信服应用交付控制台", "安全产品", "tcp", 9222, "低交互"},
	{"baota", "宝塔面板", "运维平台", "tcp", 9224, "低交互"},
	{"canal", "Canal 管理台", "数据库/中间件", "tcp", 9219, "低交互"},
	{"cisco-vpn", "Cisco SSL VPN", "安全产品", "tcp", 9299, "低交互"},
	{"cloudreve", "Cloudreve 网盘", "NAS存储", "tcp", 9203, "低交互"},
	{"confluence", "Atlassian Confluence", "OA系统", "tcp", 9293, "低交互"},
	{"coremail", "Coremail 邮件系统", "邮件系统", "tcp", 9094, "低交互"},
	{"cpanel", "cPanel", "运维平台", "tcp", 9220, "低交互"},
	{"edr-sangfor", "深信服 EDR", "安全产品", "tcp", 9223, "低交互"},
	{"electric", "电力监控平台", "IoT设备", "tcp", 9303, "低交互"},
	{"esxi", "VMware ESXi", "运维平台", "tcp", 9190, "低交互"},
	{"exchange", "Microsoft Exchange", "邮件系统", "tcp", 9095, "低交互"},
	{"filebrowser", "File Browser", "NAS存储", "tcp", 9204, "低交互"},
	{"fw-360", "360 防火墙", "安全产品", "tcp", 9218, "低交互"},
	{"fw-haofeng", "皓峰防火墙", "安全产品", "tcp", 9217, "低交互"},
	{"fw-nsfocus", "绿盟防火墙", "安全产品", "tcp", 9084, "低交互"},
	{"fw-topsec", "天融信防火墙", "安全产品", "tcp", 9081, "低交互"},
	{"fw-zkww", "中科网威防火墙", "安全产品", "tcp", 9214, "低交互"},
	{"gitlab", "GitLab", "运维平台", "tcp", 9093, "低交互"},
	{"gophish", "Gophish", "安全产品", "tcp", 9205, "低交互"},
	{"huorong-zd", "火绒终端安全", "安全产品", "tcp", 9295, "低交互"},
	{"iis", "Microsoft IIS", "Web服务器", "tcp", 9199, "低交互"},
	{"intel-am", "Intel Active Management", "运维平台", "tcp", 9085, "低交互"},
	{"iot-hikcam", "海康威视管理页", "IoT设备", "tcp", 9082, "低交互"},
	{"isport", "iSport 平台", "OA系统", "tcp", 9302, "低交互"},
	{"jenkins", "Jenkins", "运维平台", "tcp", 8080, "低交互"},
	{"jira", "Atlassian Jira", "OA系统", "tcp", 9292, "低交互"},
	{"joomla", "Joomla", "Web服务器", "tcp", 8080, "低交互"},
	{"jspspy", "JSPSpy", "安全产品", "tcp", 9294, "低交互"},
	{"jumpserver", "JumpServer", "运维平台", "tcp", 9304, "低交互"},
	{"kelai-qll", "科来网络分析系统", "安全产品", "tcp", 9296, "低交互"},
	{"kibana", "Kibana", "运维平台", "tcp", 9221, "低交互"},
	{"mailu", "Mailu", "邮件系统", "tcp", 9206, "低交互"},
	{"nagios", "Nagios", "运维平台", "tcp", 9191, "低交互"},
	{"nas-qnap", "QNAP QTS", "NAS存储", "tcp", 9211, "低交互"},
	{"nginx", "Nginx 默认站点", "Web服务器", "tcp", 9291, "低交互"},
	{"oa", "OA 办公系统", "OA系统", "tcp", 9096, "低交互"},
	{"oa-gov", "政务 OA", "OA系统", "tcp", 9098, "低交互"},
	{"oa-tongda", "通达 OA", "OA系统", "tcp", 9099, "低交互"},
	{"oa-yy", "用友 OA", "OA系统", "tcp", 9215, "低交互"},
	{"phpadmin", "phpMyAdmin", "Web服务器", "tcp", 9301, "低交互"},
	{"portainer", "Portainer", "运维平台", "tcp", 9207, "低交互"},
	{"poste", "Poste.io", "邮件系统", "tcp", 9208, "低交互"},
	{"printer-dell", "Dell 打印机", "IoT设备", "tcp", 9084, "低交互"},
	{"qzsec", "启明星辰安全产品", "安全产品", "tcp", 9193, "低交互"},
	{"router-aruba", "Aruba 网络设备", "网络设备", "tcp", 9101, "低交互"},
	{"router-h3c", "H3C 路由器", "网络设备", "tcp", 9092, "低交互"},
	{"router-ikuai", "爱快路由器", "网络设备", "tcp", 9202, "低交互"},
	{"router-openwrt", "OpenWrt 路由器", "网络设备", "tcp", 9201, "低交互"},
	{"router-ruijie", "锐捷路由器", "网络设备", "tcp", 9196, "低交互"},
	{"router-tplink", "TP-Link 路由器", "网络设备", "tcp", 9197, "低交互"},
	{"routos", "RouterOS", "网络设备", "tcp", 8291, "低交互"},
	{"ruoyi", "若依管理系统", "CRM/ERP", "tcp", 9213, "低交互"},
	{"sangfor-fcg", "深信服下一代防火墙", "安全产品", "tcp", 9298, "低交互"},
	{"sangfor-vpn", "深信服 SSL VPN", "安全产品", "tcp", 9091, "低交互"},
	{"synology-nas", "群晖 DiskStation", "NAS存储", "tcp", 9194, "低交互"},
	{"tdp", "TDP 运维平台", "运维平台", "tcp", 9209, "低交互"},
	{"thinkphp", "ThinkPHP", "Web服务器", "tcp", 9226, "低交互"},
	{"tomcat", "Apache Tomcat", "Web服务器", "tcp", 9198, "低交互"},
	{"uniaccess-lr", "UniAccess 访问控制", "安全产品", "tcp", 9225, "低交互"},
	{"weblogic", "Oracle WebLogic", "Web服务器", "tcp", 7001, "低交互"},
	{"webmin", "Webmin", "运维平台", "tcp", 10000, "低交互"},
	{"websphere", "IBM WebSphere", "Web服务器", "tcp", 9080, "低交互"},
	{"wordpress", "WordPress", "Web服务器", "tcp", 9000, "低交互"},
	{"zabbix", "Zabbix", "运维平台", "tcp", 9192, "低交互"},
	{"zhongke-kongzhi", "中科控制平台", "IoT设备", "tcp", 9297, "低交互"},
	{"zimbra", "Zimbra 邮件系统", "邮件系统", "tcp", 9210, "低交互"},
}

var webTemplateCatalogCodes = func() map[string]bool {
	result := make(map[string]bool, len(webTemplateServiceSeeds))
	for _, seed := range webTemplateServiceSeeds {
		result[seed.code] = true
	}
	return result
}()

func CurrentPotServiceCodes() []string {
	seeds := potServiceSeeds()
	codes := make([]string, 0, len(seeds))
	for _, seed := range seeds {
		codes = append(codes, seed.code)
	}
	return codes
}

func IsCurrentPotService(code string) bool {
	for _, seed := range potServiceSeeds() {
		if seed.code == code {
			return true
		}
	}
	return false
}

func seedPotServices(db *gorm.DB) error {
	seeds := potServiceSeeds()
	schema, _ := json.Marshal(map[string]any{"type": "object", "properties": map[string]any{"banner": map[string]string{"type": "string"}}})
	staticSchema, _ := json.Marshal(map[string]any{"type": "object", "properties": map[string]any{}})
	for _, seed := range seeds {
		itemSchema := schema
		description := fmt.Sprintf("模拟 %s 的低交互蜜罐服务", seed.name)
		if webTemplateCatalogCodes[seed.code] {
			itemSchema = staticSchema
			description = fmt.Sprintf("直接使用 honeypot-templates-server 原始资源模拟 %s", seed.name)
		}
		item := PotService{Code: seed.code, Name: seed.name, Category: seed.category, Protocol: seed.protocol, DefaultPort: seed.port, Depth: seed.depth, Description: description, ConfigSchema: datatypes.JSON(itemSchema)}
		if err := db.Where("code = ?", item.Code).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedRules(db *gorm.DB) error {
	var count int64
	if err := db.Model(&AlertRule{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	rules := []AlertRule{
		{Base: NewBase(), Name: "蜜饵命中", Enabled: true, EventType: "decoy.*", Level: "critical", Threshold: 1, WindowMinute: 1, SilenceMinute: 0, ChannelIDs: datatypes.JSON("[]")},
		{Base: NewBase(), Name: "凭据捕获", Enabled: true, EventType: "*.credential", Level: "high", Threshold: 1, WindowMinute: 1, SilenceMinute: 5, ChannelIDs: datatypes.JSON("[]")},
		{Base: NewBase(), Name: "端口扫描", Enabled: true, EventType: "port.scan", Level: "low", Threshold: 1, WindowMinute: 5, SilenceMinute: 30, ChannelIDs: datatypes.JSON("[]")},
	}
	return db.Create(&rules).Error
}

func LikePattern(value string) string { return "%" + strings.ReplaceAll(value, "%", "\\%") + "%" }

func NowPtr() *time.Time { now := time.Now(); return &now }
