package httpapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func nodeStabilityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.Node{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestIssueNodeTokenKeepsHealthyCertificateAndConnection(t *testing.T) {
	db := nodeStabilityTestDB(t)
	now := time.Now()
	expires := now.Add(400 * 24 * time.Hour)
	issued := now.Add(-time.Hour)
	node := store.Node{
		Base:                 store.NewBase(),
		Name:                 "online-node",
		Status:               "online",
		CertificateSerial:    "existing-serial",
		CertificateIssuedAt:  &issued,
		CertificateExpiresAt: &expires,
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	gateway := NewAgentGateway(nil, nil)
	gateway.connections[node.ID] = &agentConnection{nodeID: node.ID}
	api := &API{db: db, agents: gateway}

	token, err := api.issueNodeToken(&node)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || !gateway.Online(node.ID) {
		t.Fatalf("token=%q online=%v; issuing an installer must not disconnect a healthy node", token, gateway.Online(node.ID))
	}
	var stored store.Node
	if err := db.First(&stored, "id = ?", node.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != "online" || stored.CertificateSerial != "existing-serial" || stored.CertificateExpiresAt == nil || !stored.CertificateExpiresAt.Equal(expires) {
		t.Fatalf("healthy identity changed while issuing installer: %#v", stored)
	}
	if stored.RegistrationTokenHash == "" || stored.TokenExpiresAt == nil || !stored.TokenExpiresAt.After(now) {
		t.Fatalf("one-time registration token was not staged: %#v", stored)
	}
}

func TestIssueNodeTokenMarksOnlyUnenrolledNodeRegistering(t *testing.T) {
	db := nodeStabilityTestDB(t)
	node := store.Node{Base: store.NewBase(), Name: "new-node", Status: "offline"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	api := &API{db: db}
	if _, err := api.issueNodeToken(&node); err != nil {
		t.Fatal(err)
	}
	if node.Status != "registering" || node.CertificateSerial != "" {
		t.Fatalf("unenrolled node state = %#v", node)
	}
}

func TestDisconnectGraceCannotOverrideReplacementConnection(t *testing.T) {
	db := nodeStabilityTestDB(t)
	node := store.Node{Base: store.NewBase(), Name: "reconnecting-node", Status: "online"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	gateway := NewAgentGateway(db, nil)
	gateway.disconnectGrace = 15 * time.Millisecond
	oldConnection := &agentConnection{nodeID: node.ID, generation: 1}
	newConnection := &agentConnection{nodeID: node.ID, generation: 2}
	gateway.connections[node.ID] = oldConnection
	gateway.generations[node.ID] = oldConnection.generation
	gateway.connectionClosed(oldConnection)
	gateway.mu.Lock()
	gateway.connections[node.ID] = newConnection
	gateway.generations[node.ID] = newConnection.generation
	gateway.mu.Unlock()
	if err := db.Model(&node).Update("status", "online").Error; err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := db.First(&node, "id = ?", node.ID).Error; err != nil {
		t.Fatal(err)
	}
	if node.Status != "online" || !gateway.Online(node.ID) {
		t.Fatalf("old disconnect timer overrode replacement connection: status=%s online=%v", node.Status, gateway.Online(node.ID))
	}
}
