package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/agentupdate"
	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openUpgradeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.Node{}, &store.UpgradeTask{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestResolveRolloutNodesCombinesGroupsAndExplicitNodes(t *testing.T) {
	db := openUpgradeTestDB(t)
	nodes := []store.Node{
		{Base: store.NewBase(), Name: "离线节点", GroupName: "办公网", Status: "offline", OS: "linux", Arch: "amd64"},
		{Base: store.NewBase(), Name: "在线节点", GroupName: "办公网", Status: "online", OS: "linux", Arch: "amd64"},
		{Base: store.NewBase(), Name: "生产节点", GroupName: "生产网", Status: "online", OS: "linux", Arch: "arm64"},
	}
	if err := db.Create(&nodes).Error; err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveRolloutNodes(db, []string{nodes[2].ID, nodes[2].ID}, []string{"办公网"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 3 {
		t.Fatalf("resolved %d nodes, want 3", len(resolved))
	}
	if resolved[0].Status != "online" || resolved[1].Status != "online" {
		t.Fatalf("online canaries were not ordered first: %#v", resolved)
	}
}

func TestResolveRolloutNodesRejectsUnknownAndRevokedTargets(t *testing.T) {
	db := openUpgradeTestDB(t)
	revoked := store.Node{Base: store.NewBase(), Name: "已撤销", GroupName: "隔离区", Status: "revoked", OS: "linux", Arch: "amd64"}
	if err := db.Create(&revoked).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := resolveRolloutNodes(db, nil, []string{"不存在"}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("unknown group returned %v", err)
	}
	if _, err := resolveRolloutNodes(db, nil, []string{"隔离区"}); err == nil || !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("revoked target returned %v", err)
	}
}

func TestCreateRolloutRejectsWindowsWithoutStableSupervisor(t *testing.T) {
	db := openUpgradeTestDB(t)
	if err := db.AutoMigrate(&store.AgentRelease{}, &store.AgentBuild{}, &store.UpgradeRollout{}); err != nil {
		t.Fatal(err)
	}
	release := store.AgentRelease{Base: store.NewBase(), Version: "2.0.0", Status: "active"}
	node := store.Node{Base: store.NewBase(), Name: "Windows 节点", GroupName: "办公网", Status: "online", OS: "windows", Arch: "amd64"}
	if err := db.Create(&release).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	build := store.AgentBuild{Base: store.NewBase(), ReleaseID: release.ID, OS: "windows", Arch: "amd64", Filename: "unused.exe", SHA256: strings.Repeat("a", 64), Signature: "signature", Size: 1}
	if err := db.Create(&build).Error; err != nil {
		t.Fatal(err)
	}
	manager := &UpgradeManager{db: db, hub: NewHub(), wake: make(chan struct{}, 1)}
	_, err := manager.CreateRollout(createRolloutInput{Name: "unsafe Windows rollout", ReleaseID: release.ID, NodeIDs: []string{node.ID}}, "admin")
	if err == nil || !strings.Contains(err.Error(), "Linux stable supervisor") {
		t.Fatalf("Windows remote rollout returned %v", err)
	}
}

func TestUpgradeResultPreservesAutomaticRollbackOutcome(t *testing.T) {
	db := openUpgradeTestDB(t)
	task := store.UpgradeTask{
		Base: store.NewBase(), RolloutID: "rollout-1", NodeID: "node-1", BuildID: "build-1",
		TargetVersion: "2.0.0", Status: "restarting",
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	manager := &UpgradeManager{db: db, hub: NewHub(), wake: make(chan struct{}, 1)}
	manager.NodeResult(task.ID, "rolled_back", false, "health check failed")
	if err := db.First(&task, "id = ?", task.ID).Error; err != nil {
		t.Fatal(err)
	}
	if task.Status != "rolled_back" || task.CompletedAt == nil || task.LastError != "health check failed" {
		t.Fatalf("rollback result was not preserved: %#v", task)
	}
	if !isUpgradeFailure(task.Status) || !isUpgradeFailure("rollback_failed") || isUpgradeFailure("succeeded") {
		t.Fatal("upgrade failure classification is incorrect")
	}
}

func TestUpgradeHeartbeatWaitsForMatchingHealthyState(t *testing.T) {
	db := openUpgradeTestDB(t)
	task := store.UpgradeTask{
		Base: store.NewBase(), RolloutID: "rollout-1", NodeID: "node-1", BuildID: "build-1",
		TargetVersion: "2.0.0", Status: "restarting",
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	manager := &UpgradeManager{db: db, hub: NewHub(), wake: make(chan struct{}, 1)}
	manager.NodeHeartbeat("node-1", "2.0.0", &nodeUpgradeReport{TaskID: task.ID, TargetVersion: "2.0.0", Status: "awaiting_health"})
	if err := db.First(&task, "id = ?", task.ID).Error; err != nil {
		t.Fatal(err)
	}
	if task.Status != "restarting" {
		t.Fatalf("pre-health heartbeat completed task: %s", task.Status)
	}
	manager.NodeHeartbeat("node-1", "2.0.0", &nodeUpgradeReport{TaskID: task.ID, TargetVersion: "2.0.0", Status: "healthy"})
	if err := db.First(&task, "id = ?", task.ID).Error; err != nil {
		t.Fatal(err)
	}
	if task.Status != "succeeded" || task.ConfirmedVersion != "2.0.0" {
		t.Fatalf("healthy heartbeat did not complete task: %#v", task)
	}
}

func TestUpgradeTaskViewIncludesNodeIdentity(t *testing.T) {
	task := store.UpgradeTask{Base: store.NewBase(), NodeID: "node-1", Status: "pending", TargetVersion: "2.0.0"}
	node := store.Node{Base: store.Base{ID: "node-1"}, Name: "杭州节点", GroupName: "公网", Status: "online", Version: "1.0.0", OS: "linux", Arch: "amd64"}
	view := upgradeTaskView(task, node)
	nodeView, ok := view["node"].(gin.H)
	if !ok || nodeView["name"] != "杭州节点" || nodeView["group"] != "公网" {
		t.Fatalf("unexpected task node view: %#v", view["node"])
	}
}

func TestScanReleaseSnapshotsImmutableVersionedArtifact(t *testing.T) {
	db := openUpgradeTestDB(t)
	if err := db.AutoMigrate(&store.AgentRelease{}, &store.AgentBuild{}, &store.UpgradeRollout{}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	downloads := filepath.Join(root, "downloads")
	pki := filepath.Join(root, "pki")
	if err := os.MkdirAll(downloads, 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(downloads, "honeynet-agent-linux-amd64")
	if err := os.WriteFile(source, []byte("first Agent binary"), 0755); err != nil {
		t.Fatal(err)
	}
	signer, err := agentupdate.LoadOrCreateSigner(pki)
	if err != nil {
		t.Fatal(err)
	}
	manager := &UpgradeManager{db: db, cfg: config.Config{DownloadsDir: downloads, PKIDir: pki}, signer: signer}
	_, builds, err := manager.ScanRelease("1.0.0", "first")
	if err != nil {
		t.Fatal(err)
	}
	if len(builds) != 1 || !strings.HasPrefix(builds[0].Filename, "honeynet-agent-release-1.0.0-linux-amd64-") {
		t.Fatalf("unexpected versioned builds: %#v", builds)
	}
	artifact := filepath.Join(pki, "agent-releases", builds[0].Filename)
	before, err := os.ReadFile(artifact)
	if err != nil || string(before) != "first Agent binary" {
		t.Fatalf("artifact = %q, %v", before, err)
	}
	if err := os.WriteFile(source, []byte("second Agent binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.ScanRelease("2.0.0", "second"); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(artifact)
	if err != nil || string(after) != "first Agent binary" {
		t.Fatalf("old release artifact changed: %q, %v", after, err)
	}
	cfg := config.Config{DownloadsDir: downloads, PKIDir: pki}
	if _, err := verifyAgentBuildArtifact(cfg, builds[0], true); err != nil {
		t.Fatalf("valid immutable artifact was rejected: %v", err)
	}
	if err := os.WriteFile(artifact, []byte("xxxxx Agent binary"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyAgentBuildArtifact(cfg, builds[0], true); err == nil {
		t.Fatal("corrupted release artifact passed SHA-256 verification")
	}
}
