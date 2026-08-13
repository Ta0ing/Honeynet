package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/honeynet/honeynet/internal/agent/protocol"
	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AgentEnvelope struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	TS      int64           `json:"ts"`
	Payload json.RawMessage `json:"payload"`
}

type agentConnection struct {
	nodeID     string
	generation uint64
	conn       *websocket.Conn
	writeMu    sync.Mutex
}

func (c *agentConnection) write(messageType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	envelope := AgentEnvelope{ID: uuid.NewString(), Type: messageType, TS: time.Now().Unix(), Payload: data}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(envelope)
}

type AgentGateway struct {
	mu              sync.RWMutex
	connections     map[string]*agentConnection
	generations     map[string]uint64
	nextGeneration  uint64
	disconnectGrace time.Duration
	db              *gorm.DB
	console         *Hub
	upgrades        *UpgradeManager
	updatePublicKey string
	updateKeyID     string
}

var errNodeOffline = errors.New("node is offline")

func NewAgentGateway(db *gorm.DB, console *Hub) *AgentGateway {
	gateway := &AgentGateway{
		connections:     map[string]*agentConnection{},
		generations:     map[string]uint64{},
		disconnectGrace: 45 * time.Second,
		db:              db,
		console:         console,
	}
	// Presence is process-local. A clean or unclean Server restart cannot
	// preserve any Agent WebSocket, so do not expose stale "online" rows while
	// Agents reconnect to the new process.
	if db != nil {
		db.Model(&store.Node{}).
			Where("status IN ?", []string{"online", "degraded"}).
			Update("status", "offline")
	}
	return gateway
}

func (g *AgentGateway) SetUpgradeManager(manager *UpgradeManager) {
	g.upgrades = manager
}

func (g *AgentGateway) SetUpdateTrust(publicKey, keyID string) {
	g.updatePublicKey = publicKey
	g.updateKeyID = keyID
}

func (g *AgentGateway) Serve(c *gin.Context) {
	node := c.MustGet("agent.node").(store.Node)
	observedIP := requestRemoteIP(c.Request)
	upgrader := websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &agentConnection{nodeID: node.ID, conn: conn}
	g.mu.Lock()
	g.nextGeneration++
	client.generation = g.nextGeneration
	if old := g.connections[node.ID]; old != nil {
		_ = old.conn.Close()
	}
	g.connections[node.ID] = client
	g.generations[node.ID] = client.generation
	g.mu.Unlock()
	// Register cleanup before the first write. A peer can disappear immediately
	// after the HTTP upgrade; without this defer, a failed hello.ack would leave
	// a stale connection in the registry and keep the node falsely online.
	defer func() {
		_ = conn.Close()
		g.connectionClosed(client)
	}()

	now := time.Now()
	g.db.Model(&node).Updates(map[string]any{"status": "online", "last_heartbeat_at": now})
	addressNode, _ := updateNodeAddressReport(g.db, node.ID, observedIP, nil)
	g.console.Broadcast("node.status", gin.H{"id": node.ID, "status": "online", "last_heartbeat_at": now, "ip": addressNode.IP, "public_ip": addressNode.PublicIP, "public_ips": addressNode.PublicIPs, "private_ips": addressNode.PrivateIPs, "address_mode": addressNode.AddressMode})
	if err := client.write("hello.ack", gin.H{"node_id": node.ID, "heartbeat_interval": 30, "server_time": now.Unix(), "update_public_key": g.updatePublicKey, "update_key_id": g.updateKeyID}); err != nil {
		return
	}
	_ = g.SendDetectionRules(node.ID)
	_ = g.SendPotApply(node.ID)
	_ = g.SendDecoyApply(node.ID)
	_ = g.SendSenseApply(node.ID)

	conn.SetReadLimit(2 << 20)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		var message AgentEnvelope
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		switch message.Type {
		case "heartbeat":
			g.handleHeartbeat(node.ID, message)
			_ = client.write("heartbeat.ack", gin.H{"ref_id": message.ID, "server_time": time.Now().Unix()})
		case "hello":
			g.handleHello(node.ID, message)
		case "result":
			g.handleResult(node.ID, message)
		}
	}
}

// connectionClosed gives a transient network break a short reconnect grace
// period. The generation check is the important part: an old socket, or an
// earlier disconnect timer, can never mark a newer connection offline.
func (g *AgentGateway) connectionClosed(client *agentConnection) {
	g.mu.Lock()
	if g.connections[client.nodeID] != client {
		g.mu.Unlock()
		return
	}
	delete(g.connections, client.nodeID)
	generation := client.generation
	grace := g.disconnectGrace
	g.mu.Unlock()

	markOffline := func() {
		// Keep the connectivity decision, persisted status and notification in
		// the same critical section. Otherwise a replacement connection could be
		// registered after the check but before the old timer writes "offline".
		g.mu.Lock()
		stillDisconnected := g.connections[client.nodeID] == nil && g.generations[client.nodeID] == generation
		if !stillDisconnected || g.db == nil {
			g.mu.Unlock()
			return
		}
		result := g.db.Model(&store.Node{}).
			Where("id = ? AND status IN ?", client.nodeID, []string{"online", "degraded"}).
			Update("status", "offline")
		if result.Error == nil && result.RowsAffected == 1 && g.console != nil {
			g.console.Broadcast("node.status", gin.H{"id": client.nodeID, "status": "offline"})
		}
		g.mu.Unlock()
	}
	if grace <= 0 {
		markOffline()
		return
	}
	time.AfterFunc(grace, markOffline)
}

func (g *AgentGateway) SendDetectionRules(nodeID string) error {
	var items []store.DetectionRule
	if err := g.db.Where("enabled = ? AND agent_enabled = ? AND validation_error = ''", true, true).Order("rule_key ASC").Find(&items).Error; err != nil {
		return err
	}
	rules := make([]detection.Rule, 0, len(items))
	var revision int64
	for _, item := range items {
		rule, err := detectionProtocolRule(item)
		if err != nil {
			return err
		}
		rules = append(rules, rule)
		if item.Revision > revision {
			revision = item.Revision
		}
	}
	g.mu.RLock()
	client := g.connections[nodeID]
	g.mu.RUnlock()
	if client == nil {
		return errNodeOffline
	}
	if err := client.write("cmd.rules.apply", gin.H{"revision": revision, "rules": rules}); err != nil {
		return err
	}
	g.db.Model(&store.Node{}).Where("id = ?", nodeID).Updates(map[string]any{"detection_rule_status": "pending", "detection_rule_error": ""})
	return nil
}

func (g *AgentGateway) BroadcastDetectionRules() {
	g.mu.RLock()
	nodeIDs := make([]string, 0, len(g.connections))
	for nodeID := range g.connections {
		nodeIDs = append(nodeIDs, nodeID)
	}
	g.mu.RUnlock()
	for _, nodeID := range nodeIDs {
		_ = g.SendDetectionRules(nodeID)
	}
}

func (g *AgentGateway) handleHello(nodeID string, message AgentEnvelope) {
	var payload struct {
		Version      string   `json:"version"`
		OS           string   `json:"os"`
		Arch         string   `json:"arch"`
		Capabilities []string `json:"capabilities"`
		RuleRevision int64    `json:"rule_revision"`
		RuleCount    int      `json:"rule_count"`
	}
	if json.Unmarshal(message.Payload, &payload) != nil {
		return
	}
	updates := map[string]any{
		"version": payload.Version, "os": payload.OS, "arch": payload.Arch,
		"capabilities":            normalizeCapabilities(payload.Capabilities),
		"detection_rule_revision": payload.RuleRevision, "detection_rule_count": payload.RuleCount,
	}
	serverRevision, _, _ := detectionRuleSetStatus(g.db, true)
	if payload.RuleRevision == serverRevision {
		updates["detection_rule_status"] = "synced"
	} else {
		updates["detection_rule_status"] = "stale"
	}
	if g.db.Model(&store.Node{}).Where("id = ?", nodeID).Updates(updates).Error != nil {
		return
	}
	g.console.Broadcast("node.capabilities", gin.H{"id": nodeID, "capabilities": json.RawMessage(updates["capabilities"].(datatypes.JSON))})
}

func (g *AgentGateway) handleHeartbeat(nodeID string, message AgentEnvelope) {
	var payload struct {
		Version      string                `json:"version"`
		IP           string                `json:"ip"`
		IPs          []string              `json:"ips"`
		OS           string                `json:"os"`
		Arch         string                `json:"arch"`
		Capabilities []string              `json:"capabilities"`
		Healthy      bool                  `json:"healthy"`
		QueuedEvents int                   `json:"queued_events"`
		RuleRevision int64                 `json:"rule_revision"`
		RuleCount    int                   `json:"rule_count"`
		Upgrade      *nodeUpgradeReport    `json:"upgrade"`
		Sense        *protocol.SenseStatus `json:"sense"`
		Pots         []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"pots"`
		Decoys []protocol.DecoyStatus `json:"decoys"`
	}
	if json.Unmarshal(message.Payload, &payload) != nil {
		return
	}
	status := "online"
	if !payload.Healthy {
		status = "degraded"
	}
	now := time.Now()
	updates := map[string]any{"version": payload.Version, "os": payload.OS, "arch": payload.Arch, "status": status, "last_heartbeat_at": now, "queued_events": payload.QueuedEvents, "detection_rule_revision": payload.RuleRevision, "detection_rule_count": payload.RuleCount}
	serverRevision, _, _ := detectionRuleSetStatus(g.db, true)
	if payload.RuleRevision == serverRevision {
		updates["detection_rule_status"] = "synced"
		updates["detection_rule_error"] = ""
	} else {
		updates["detection_rule_status"] = "stale"
	}
	if payload.Capabilities != nil {
		updates["capabilities"] = normalizeCapabilities(payload.Capabilities)
	}
	g.db.Model(&store.Node{}).Where("id = ?", nodeID).Updates(updates)
	reportedIPs := payload.IPs
	if payload.IP != "" {
		reportedIPs = append(append([]string(nil), reportedIPs...), payload.IP)
	}
	addressNode, _ := updateNodeAddressReport(g.db, nodeID, "", reportedIPs)
	for _, pot := range payload.Pots {
		if pot.ID != "" && pot.Status != "" {
			g.db.Model(&store.PotInstance{}).Where("id = ? AND node_id = ?", pot.ID, nodeID).Update("actual_status", pot.Status)
		}
	}
	for _, decoy := range payload.Decoys {
		g.updateDecoyStatus(nodeID, decoy)
	}
	if payload.Sense != nil {
		g.updateSenseStatus(nodeID, *payload.Sense)
	}
	g.console.Broadcast("node.status", gin.H{"id": nodeID, "status": status, "last_heartbeat_at": now, "ip": addressNode.IP, "public_ip": addressNode.PublicIP, "public_ips": addressNode.PublicIPs, "private_ips": addressNode.PrivateIPs, "address_mode": addressNode.AddressMode})
	if g.upgrades != nil {
		g.upgrades.NodeHeartbeat(nodeID, payload.Version, payload.Upgrade)
	}
}

func (g *AgentGateway) handleResult(nodeID string, message AgentEnvelope) {
	var payload struct {
		Operation   string                `json:"operation"`
		TaskID      string                `json:"task_id"`
		PotID       string                `json:"pot_id"`
		DecoyID     string                `json:"decoy_id"`
		Status      string                `json:"status"`
		Success     bool                  `json:"success"`
		Error       string                `json:"error"`
		ManagedPath string                `json:"managed_path"`
		Revision    int64                 `json:"revision"`
		RuleCount   int                   `json:"rule_count"`
		Sense       *protocol.SenseStatus `json:"sense"`
	}
	if json.Unmarshal(message.Payload, &payload) != nil {
		return
	}
	if payload.Operation == "upgrade" {
		if g.upgrades != nil {
			g.upgrades.NodeResult(payload.TaskID, payload.Status, payload.Success, payload.Error)
		}
		return
	}
	if payload.Operation == "rules" {
		now := time.Now()
		status := "error"
		if payload.Success {
			serverRevision, _, _ := detectionRuleSetStatus(g.db, true)
			if payload.Revision == serverRevision {
				status = "synced"
			} else {
				status = "stale"
			}
		}
		g.db.Model(&store.Node{}).Where("id = ?", nodeID).Updates(map[string]any{
			"detection_rule_revision": payload.Revision, "detection_rule_count": payload.RuleCount,
			"detection_rule_status": status, "detection_rule_synced_at": now, "detection_rule_error": payload.Error,
		})
		g.console.Broadcast("node.rules", gin.H{"id": nodeID, "revision": payload.Revision, "revision_text": strconv.FormatInt(payload.Revision, 10), "rule_count": payload.RuleCount, "status": status, "error": payload.Error})
		return
	}
	if payload.Operation == "sense" {
		status := protocol.SenseStatus{ActualStatus: payload.Status, LastError: payload.Error}
		if payload.Sense != nil {
			status = *payload.Sense
		}
		if !payload.Success {
			if status.ActualStatus != "unsupported" {
				status.ActualStatus = "error"
			}
			if status.LastError == "" {
				status.LastError = payload.Error
			}
		}
		g.updateSenseStatus(nodeID, status)
		return
	}
	if payload.Operation == "decoy" {
		status := payload.Status
		if !payload.Success {
			status = "error"
		}
		g.updateDecoyStatus(nodeID, protocol.DecoyStatus{ID: payload.DecoyID, Status: status, ManagedPath: payload.ManagedPath, LastError: payload.Error})
		return
	}
	if payload.PotID == "" {
		return
	}
	status := payload.Status
	if !payload.Success && status != "unsupported" {
		status = "error"
	}
	if status != "" {
		g.db.Model(&store.PotInstance{}).Where("id = ? AND node_id = ?", payload.PotID, nodeID).Update("actual_status", status)
	}
	g.console.Broadcast("pot.status", gin.H{"id": payload.PotID, "node_id": nodeID, "status": status, "error": payload.Error})
}

func (g *AgentGateway) SendUpgrade(nodeID string, payload any) error {
	g.mu.RLock()
	client := g.connections[nodeID]
	g.mu.RUnlock()
	if client == nil {
		return errNodeOffline
	}
	return client.write("cmd.agent.upgrade", payload)
}

func (g *AgentGateway) SendSenseApply(nodeID string) error {
	item, err := ensureNodeSenseConfig(g.db, nodeID)
	if err != nil {
		return err
	}
	config, err := nodeSenseProtocolConfig(item)
	if err != nil {
		return err
	}
	g.mu.RLock()
	client := g.connections[nodeID]
	g.mu.RUnlock()
	if client == nil {
		return errNodeOffline
	}
	return client.write("cmd.sense.apply", gin.H{"revision": time.Now().UnixNano(), "config": config})
}

func (g *AgentGateway) updateSenseStatus(nodeID string, status protocol.SenseStatus) {
	if status.ActualStatus == "" {
		return
	}
	item, err := ensureNodeSenseConfig(g.db, nodeID)
	if err != nil {
		return
	}
	updates := map[string]any{
		"actual_status": status.ActualStatus, "observed_packets": status.ObservedPackets,
		"detections": status.Detections, "started_at": status.StartedAt,
		"last_detection_at": status.LastDetectionAt, "last_error": status.LastError,
	}
	if g.db.Model(&item).Updates(updates).Error != nil {
		return
	}
	g.console.Broadcast("sense.status", gin.H{
		"node_id": nodeID, "actual_status": status.ActualStatus,
		"observed_packets": status.ObservedPackets, "detections": status.Detections,
		"started_at": status.StartedAt, "last_detection_at": status.LastDetectionAt,
		"last_error": status.LastError,
	})
}

func (g *AgentGateway) updateDecoyStatus(nodeID string, status protocol.DecoyStatus) {
	if status.ID == "" || status.Status == "" {
		return
	}
	updates := map[string]any{"actual_status": status.Status, "managed_path": status.ManagedPath, "last_error": status.LastError}
	if status.Status == "monitoring" || status.Status == "passive" {
		updates["deployed_at"] = gorm.Expr("COALESCE(deployed_at, ?)", time.Now())
	} else if status.Status == "stopped" {
		updates["deployed_at"] = nil
		updates["managed_path"] = ""
	}
	result := g.db.Model(&store.Decoy{}).Where("id = ? AND node_id = ?", status.ID, nodeID).Updates(updates)
	if result.Error != nil || result.RowsAffected == 0 {
		return
	}
	g.console.Broadcast("decoy.status", gin.H{"id": status.ID, "node_id": nodeID, "status": status.Status, "managed_path": updates["managed_path"], "error": status.LastError})
}

func (g *AgentGateway) SendPotApply(nodeID string) error {
	var pots []store.PotInstance
	if err := g.db.Preload("Template").Where("node_id = ?", nodeID).Order("created_at").Find(&pots).Error; err != nil {
		return err
	}
	targets := make([]protocol.PotTarget, 0, len(pots))
	for _, pot := range pots {
		targets = append(targets, potProtocolTarget(pot))
	}
	g.mu.RLock()
	client := g.connections[nodeID]
	g.mu.RUnlock()
	if client == nil {
		return nil
	}
	return client.write("cmd.pot.apply", gin.H{"revision": time.Now().UnixNano(), "pots": targets})
}

func potProtocolTarget(pot store.PotInstance) protocol.PotTarget {
	config := map[string]any{}
	if len(pot.Config) > 0 {
		_ = json.Unmarshal(pot.Config, &config)
	}
	target := protocol.PotTarget{ID: pot.ID, Service: pot.ServiceCode, Name: pot.Name, Port: pot.Port, Config: config, DesiredStatus: pot.DesiredStatus}
	if pot.Template != nil {
		target.Template = &protocol.WebTemplate{ID: pot.Template.ID, Name: pot.Template.Name, Version: pot.Template.Version, YAML: pot.Template.YAML}
	}
	return target
}

func (g *AgentGateway) SendDecoyApply(nodeID string) error {
	var decoys []store.Decoy
	if err := g.db.Where("node_id = ?", nodeID).Order("created_at").Find(&decoys).Error; err != nil {
		return err
	}
	targets := make([]protocol.DecoyTarget, 0, len(decoys))
	for _, decoy := range decoys {
		targets = append(targets, decoyProtocolTarget(decoy))
	}
	g.mu.RLock()
	client := g.connections[nodeID]
	g.mu.RUnlock()
	if client == nil {
		return nil
	}
	return client.write("cmd.decoy.apply", gin.H{"revision": time.Now().UnixNano(), "decoys": targets})
}

func decoyProtocolTarget(decoy store.Decoy) protocol.DecoyTarget {
	config := map[string]any{}
	if len(decoy.Config) > 0 {
		_ = json.Unmarshal(decoy.Config, &config)
	}
	return protocol.DecoyTarget{ID: decoy.ID, Name: decoy.Name, Type: decoy.Type, Config: config, Status: decoy.Status}
}

func (g *AgentGateway) Online(nodeID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.connections[nodeID] != nil
}

func (g *AgentGateway) Revoke(nodeID string) {
	g.mu.Lock()
	client := g.connections[nodeID]
	delete(g.connections, nodeID)
	// Invalidate a pending disconnect timer from the revoked connection.
	g.nextGeneration++
	g.generations[nodeID] = g.nextGeneration
	g.mu.Unlock()
	if client != nil {
		_ = client.write("cmd.revoke", map[string]string{"reason": "node credential rotated or revoked"})
		_ = client.conn.Close()
	}
}
