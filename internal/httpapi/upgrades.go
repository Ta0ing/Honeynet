package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/agentupdate"
	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/gorm"
)

type UpgradeManager struct {
	db     *gorm.DB
	cfg    config.Config
	signer *agentupdate.Signer
	agents *AgentGateway
	hub    *Hub
	wake   chan struct{}
	start  sync.Once
}

var supportedAgentBuilds = []struct{ OS, Arch string }{
	{"linux", "386"}, {"linux", "amd64"}, {"linux", "arm"}, {"linux", "arm64"}, {"linux", "loong64"},
	{"windows", "386"}, {"windows", "amd64"},
}

func NewUpgradeManager(db *gorm.DB, cfg config.Config, signer *agentupdate.Signer, agents *AgentGateway, hub *Hub) *UpgradeManager {
	return &UpgradeManager{db: db, cfg: cfg, signer: signer, agents: agents, hub: hub, wake: make(chan struct{}, 1)}
}

func (m *UpgradeManager) Start(ctx context.Context) {
	m.start.Do(func() {
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				case <-m.wake:
				}
				m.reconcile()
			}
		}()
	})
}

func (m *UpgradeManager) notify() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *UpgradeManager) ScanRelease(version, notes string) (store.AgentRelease, []store.AgentBuild, error) {
	version = strings.TrimSpace(version)
	notes = strings.TrimSpace(notes)
	if version == "" {
		version = strings.TrimSpace(m.cfg.Version)
	}
	if version == "" {
		return store.AgentRelease{}, nil, errors.New("release version is required")
	}
	if len([]rune(notes)) > 1024 {
		return store.AgentRelease{}, nil, errors.New("release notes exceed 1024 characters")
	}
	base, err := filepath.Abs(m.cfg.DownloadsDir)
	if err != nil {
		return store.AgentRelease{}, nil, err
	}
	artifactBase, err := filepath.Abs(filepath.Join(m.cfg.PKIDir, "agent-releases"))
	if err != nil {
		return store.AgentRelease{}, nil, err
	}
	if err := os.MkdirAll(artifactBase, 0700); err != nil {
		return store.AgentRelease{}, nil, fmt.Errorf("create Agent release artifact directory: %w", err)
	}
	release := store.AgentRelease{Base: store.NewBase(), Version: version, Notes: notes, Status: "active", KeyID: m.signer.KeyID()}
	var existingRelease store.AgentRelease
	if queryErr := m.db.Where("version = ?", version).First(&existingRelease).Error; queryErr == nil {
		var rolloutCount int64
		if countErr := m.db.Model(&store.UpgradeRollout{}).Where("release_id = ?", existingRelease.ID).Count(&rolloutCount).Error; countErr != nil {
			return store.AgentRelease{}, nil, countErr
		}
		if rolloutCount > 0 {
			return store.AgentRelease{}, nil, errors.New("release has rollout history and is immutable; publish a new version instead")
		}
	} else if !errors.Is(queryErr, gorm.ErrRecordNotFound) {
		return store.AgentRelease{}, nil, queryErr
	}
	builds := make([]store.AgentBuild, 0, len(supportedAgentBuilds))
	for _, platform := range supportedAgentBuilds {
		filename := "honeynet-agent-" + platform.OS + "-" + platform.Arch
		if platform.OS == "windows" {
			filename += ".exe"
		}
		path := filepath.Join(base, filename)
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return store.AgentRelease{}, nil, err
		}
		hash := sha256.New()
		size, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return store.AgentRelease{}, nil, copyErr
		}
		if closeErr != nil {
			return store.AgentRelease{}, nil, closeErr
		}
		digest := hex.EncodeToString(hash.Sum(nil))
		descriptor := agentupdate.Descriptor{Version: version, OS: platform.OS, Arch: platform.Arch, SHA256: digest, Size: size}
		signature, err := m.signer.Sign(descriptor)
		if err != nil {
			return store.AgentRelease{}, nil, err
		}
		artifactName := releaseArtifactFilename(version, platform.OS, platform.Arch, digest, platform.OS == "windows")
		if err := snapshotReleaseArtifact(path, filepath.Join(artifactBase, artifactName), digest, size); err != nil {
			return store.AgentRelease{}, nil, fmt.Errorf("snapshot Agent build %s/%s: %w", platform.OS, platform.Arch, err)
		}
		builds = append(builds, store.AgentBuild{Base: store.NewBase(), ReleaseID: release.ID, OS: platform.OS, Arch: platform.Arch, Filename: artifactName, SHA256: digest, Signature: signature, Size: size})
	}
	if len(builds) == 0 {
		return store.AgentRelease{}, nil, errors.New("no Agent binaries were found in downloads_dir")
	}
	err = m.db.Transaction(func(tx *gorm.DB) error {
		var existing store.AgentRelease
		if err := tx.Where("version = ?", version).First(&existing).Error; err == nil {
			var rolloutCount int64
			if err := tx.Model(&store.UpgradeRollout{}).Where("release_id = ?", existing.ID).Count(&rolloutCount).Error; err != nil {
				return err
			}
			if rolloutCount > 0 {
				return errors.New("release has rollout history and is immutable; publish a new version instead")
			}
			release.ID = existing.ID
			release.CreatedAt = existing.CreatedAt
			if err := tx.Model(&existing).Updates(map[string]any{"notes": release.Notes, "status": "active", "key_id": release.KeyID}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("release_id = ?", existing.ID).Delete(&store.AgentBuild{}).Error; err != nil {
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		} else if err := tx.Create(&release).Error; err != nil {
			return err
		}
		for i := range builds {
			builds[i].ReleaseID = release.ID
			if err := tx.Create(&builds[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return release, builds, err
}

func releaseArtifactFilename(version, osName, arch, digest string, windows bool) string {
	safe := func(value string) string {
		return strings.Map(func(character rune) rune {
			if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._+-", character) {
				return character
			}
			return '-'
		}, value)
	}
	filename := fmt.Sprintf("honeynet-agent-release-%s-%s-%s-%s", safe(version), safe(osName), safe(arch), digest[:12])
	if windows {
		filename += ".exe"
	}
	return filename
}

func snapshotReleaseArtifact(source, target, expectedDigest string, expectedSize int64) error {
	if info, err := os.Stat(target); err == nil && info.Size() == expectedSize {
		file, openErr := os.Open(target)
		if openErr == nil {
			hash := sha256.New()
			_, copyErr := io.Copy(hash, file)
			closeErr := file.Close()
			if copyErr == nil && closeErr == nil && strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expectedDigest) {
				return nil
			}
		}
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	temporary, err := os.CreateTemp(filepath.Dir(target), ".agent-release-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(temporary, hash), io.LimitReader(input, expectedSize+1))
	syncErr := temporary.Sync()
	closeErr := temporary.Close()
	if copyErr != nil {
		return copyErr
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written != expectedSize || !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expectedDigest) {
		return errors.New("Agent build changed while it was being snapshotted")
	}
	if err := os.Chmod(temporaryPath, 0600); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

type createRolloutInput struct {
	Name         string   `json:"name"`
	ReleaseID    string   `json:"release_id"`
	NodeIDs      []string `json:"node_ids"`
	GroupNames   []string `json:"group_names"`
	Strategy     string   `json:"strategy"`
	CanaryCount  int      `json:"canary_count"`
	BatchSize    int      `json:"batch_size"`
	PauseSeconds int      `json:"pause_seconds"`
}

func (m *UpgradeManager) CreateRollout(input createRolloutInput, createdBy string) (store.UpgradeRollout, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.ReleaseID == "" || len(input.NodeIDs)+len(input.GroupNames) == 0 {
		return store.UpgradeRollout{}, errors.New("rollout name, release and at least one node or node group are required")
	}
	if len([]rune(input.Name)) > 128 {
		return store.UpgradeRollout{}, errors.New("rollout name exceeds 128 characters")
	}
	var release store.AgentRelease
	if err := m.db.First(&release, "id = ? AND status = ?", input.ReleaseID, "active").Error; err != nil {
		return store.UpgradeRollout{}, errors.New("release does not exist or is not active")
	}
	var builds []store.AgentBuild
	if err := m.db.Where("release_id = ?", release.ID).Find(&builds).Error; err != nil {
		return store.UpgradeRollout{}, err
	}
	buildByPlatform := map[string]store.AgentBuild{}
	for _, build := range builds {
		buildByPlatform[build.OS+"/"+build.Arch] = build
	}
	nodes, err := resolveRolloutNodes(m.db, input.NodeIDs, input.GroupNames)
	if err != nil {
		return store.UpgradeRollout{}, err
	}
	for _, node := range nodes {
		if node.OS != "linux" {
			return store.UpgradeRollout{}, fmt.Errorf("node %s uses %s: remote Agent rollout in this release requires the Linux stable supervisor; use the signed offline package for Windows", node.Name, node.OS)
		}
	}
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	var occupied int64
	if err := m.db.Table("upgrade_tasks AS task").
		Joins("JOIN upgrade_rollouts AS rollout ON rollout.id = task.rollout_id AND rollout.deleted_at IS NULL").
		Where("task.deleted_at IS NULL AND task.node_id IN ? AND rollout.status IN ?", ids, []string{"running", "paused"}).
		Count(&occupied).Error; err != nil {
		return store.UpgradeRollout{}, err
	}
	if occupied > 0 {
		return store.UpgradeRollout{}, errors.New("one or more nodes already belong to an active rollout")
	}
	if input.CanaryCount < 0 {
		input.CanaryCount = 0
	}
	if input.CanaryCount > len(nodes) {
		input.CanaryCount = len(nodes)
	}
	if input.BatchSize <= 0 {
		input.BatchSize = len(nodes)
	}
	if input.PauseSeconds < 0 || input.PauseSeconds > 86400 {
		return store.UpgradeRollout{}, errors.New("pause_seconds must be between 0 and 86400")
	}
	if input.Strategy == "" {
		input.Strategy = "canary"
	}
	if input.Strategy != "canary" {
		return store.UpgradeRollout{}, errors.New("only the canary rollout strategy is supported")
	}
	now := time.Now()
	rollout := store.UpgradeRollout{
		Base: store.NewBase(), Name: input.Name, ReleaseID: release.ID, Version: release.Version,
		Strategy: input.Strategy, CanaryCount: input.CanaryCount, BatchSize: input.BatchSize, PauseSeconds: input.PauseSeconds,
		Status: "running", CurrentWave: 0, CreatedBy: createdBy, StartedAt: &now,
	}
	tasks := make([]store.UpgradeTask, 0, len(nodes))
	verifiedBuilds := make(map[string]bool, len(builds))
	for index, node := range nodes {
		build, ok := buildByPlatform[node.OS+"/"+node.Arch]
		if !ok {
			return store.UpgradeRollout{}, fmt.Errorf("release has no build for node %s (%s/%s)", node.Name, node.OS, node.Arch)
		}
		if !verifiedBuilds[build.ID] {
			if _, err := verifyAgentBuildArtifact(m.cfg, build, true); err != nil {
				return store.UpgradeRollout{}, fmt.Errorf("release build %s/%s is unavailable or corrupted: %w", build.OS, build.Arch, err)
			}
			verifiedBuilds[build.ID] = true
		}
		wave := rolloutWave(index, input.CanaryCount, input.BatchSize)
		tasks = append(tasks, store.UpgradeTask{Base: store.NewBase(), RolloutID: rollout.ID, NodeID: node.ID, BuildID: build.ID, Wave: wave, FromVersion: node.Version, TargetVersion: release.Version, Status: "pending"})
	}
	if err := m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rollout).Error; err != nil {
			return err
		}
		return tx.Create(&tasks).Error
	}); err != nil {
		return store.UpgradeRollout{}, err
	}
	m.notify()
	return rollout, nil
}

func resolveRolloutNodes(db *gorm.DB, nodeIDs, groupNames []string) ([]store.Node, error) {
	ids := uniqueStrings(nodeIDs)
	groups := uniqueStrings(groupNames)
	if len(ids)+len(groups) == 0 {
		return nil, errors.New("at least one node or node group is required")
	}
	query := db.Model(&store.Node{})
	switch {
	case len(ids) > 0 && len(groups) > 0:
		query = query.Where("id IN ? OR group_name IN ?", ids, groups)
	case len(ids) > 0:
		query = query.Where("id IN ?", ids)
	default:
		query = query.Where("group_name IN ?", groups)
	}
	var nodes []store.Node
	if err := query.Find(&nodes).Error; err != nil {
		return nil, err
	}
	byID := make(map[string]bool, len(nodes))
	byGroup := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = true
		byGroup[node.GroupName] = true
		if node.Status == "revoked" {
			return nil, fmt.Errorf("node %s is revoked and cannot be upgraded", node.Name)
		}
	}
	for _, id := range ids {
		if !byID[id] {
			return nil, fmt.Errorf("node %s does not exist", id)
		}
	}
	for _, group := range groups {
		if !byGroup[group] {
			return nil, fmt.Errorf("node group %s does not exist or is empty", group)
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		leftOnline := nodes[i].Status == "online" || nodes[i].Status == "degraded"
		rightOnline := nodes[j].Status == "online" || nodes[j].Status == "degraded"
		if leftOnline != rightOnline {
			return leftOnline
		}
		if nodes[i].GroupName != nodes[j].GroupName {
			return nodes[i].GroupName < nodes[j].GroupName
		}
		if nodes[i].Name != nodes[j].Name {
			return nodes[i].Name < nodes[j].Name
		}
		return nodes[i].ID < nodes[j].ID
	})
	return nodes, nil
}

func rolloutWave(index, canaryCount, batchSize int) int {
	if canaryCount <= 0 {
		return index / batchSize
	}
	if index < canaryCount {
		return 0
	}
	return 1 + (index-canaryCount)/batchSize
}

type nodeUpgradeReport struct {
	TaskID        string `json:"task_id"`
	TargetVersion string `json:"target_version"`
	Status        string `json:"status"`
	LastError     string `json:"last_error"`
}

func (m *UpgradeManager) NodeHeartbeat(nodeID, version string, report *nodeUpgradeReport) {
	if nodeID == "" || version == "" {
		return
	}
	now := time.Now()
	var tasks []store.UpgradeTask
	m.db.Where("node_id = ? AND target_version = ? AND status IN ?", nodeID, version, []string{"pending", "deploying", "restarting"}).Find(&tasks)
	for _, task := range tasks {
		if report != nil && report.TaskID == task.ID {
			switch report.Status {
			case "rolled_back", "rollback_failed":
				m.NodeResult(task.ID, report.Status, false, report.LastError)
				continue
			case "awaiting_health", "rolling_back":
				continue
			case "healthy":
			default:
				continue
			}
		} else if task.Status != "pending" {
			// A remotely installed version is only successful after the Agent
			// reports its matching update task as healthy. This prevents the
			// first pre-hello heartbeat from racing the rollback watchdog.
			continue
		}
		m.db.Model(&task).Updates(map[string]any{"status": "succeeded", "confirmed_version": version, "completed_at": now, "last_error": ""})
		m.hub.Broadcast("upgrade.status", gin.H{"rollout_id": task.RolloutID, "task_id": task.ID, "node_id": nodeID, "status": "succeeded"})
	}
	if len(tasks) > 0 {
		m.notify()
	}
}

func (m *UpgradeManager) NodeResult(taskID, reportedStatus string, success bool, message string) {
	if taskID == "" {
		return
	}
	updates := map[string]any{"status": "restarting", "last_error": ""}
	if !success {
		status := "failed"
		if reportedStatus == "rolled_back" || reportedStatus == "rollback_failed" {
			status = reportedStatus
		}
		updates["status"] = status
		updates["last_error"] = sanitizeUpgradeError(message)
		now := time.Now()
		updates["completed_at"] = now
	}
	allowedStatuses := []string{"deploying", "restarting", "pending"}
	if !success && (reportedStatus == "rolled_back" || reportedStatus == "rollback_failed") {
		allowedStatuses = append(allowedStatuses, "failed")
	}
	result := m.db.Model(&store.UpgradeTask{}).Where("id = ? AND status IN ?", taskID, allowedStatuses).Updates(updates)
	if result.RowsAffected > 0 {
		status := updates["status"]
		m.hub.Broadcast("upgrade.status", gin.H{"task_id": taskID, "status": status, "error": updates["last_error"]})
		m.notify()
	}
}

func (m *UpgradeManager) reconcile() {
	var rollouts []store.UpgradeRollout
	if err := m.db.Where("status = ?", "running").Order("created_at").Find(&rollouts).Error; err != nil {
		return
	}
	for i := range rollouts {
		m.reconcileRollout(&rollouts[i])
	}
}

func (m *UpgradeManager) reconcileRollout(rollout *store.UpgradeRollout) {
	var tasks []store.UpgradeTask
	if m.db.Where("rollout_id = ? AND wave = ?", rollout.ID, rollout.CurrentWave).Order("created_at").Find(&tasks).Error != nil {
		return
	}
	if len(tasks) == 0 {
		m.completeRollout(rollout)
		return
	}
	allSucceeded := true
	latestCompletion := time.Time{}
	for i := range tasks {
		task := &tasks[i]
		if isUpgradeFailure(task.Status) {
			m.db.Model(rollout).Update("status", "paused")
			m.hub.Broadcast("upgrade.status", gin.H{"rollout_id": rollout.ID, "status": "paused"})
			return
		}
		if task.Status != "succeeded" {
			allSucceeded = false
		}
		if task.CompletedAt != nil && task.CompletedAt.After(latestCompletion) {
			latestCompletion = *task.CompletedAt
		}
	}
	if allSucceeded {
		if !latestCompletion.IsZero() && time.Since(latestCompletion) < time.Duration(rollout.PauseSeconds)*time.Second {
			return
		}
		var remaining int64
		m.db.Model(&store.UpgradeTask{}).Where("rollout_id = ? AND wave > ?", rollout.ID, rollout.CurrentWave).Count(&remaining)
		if remaining == 0 {
			m.completeRollout(rollout)
			return
		}
		m.db.Model(rollout).Update("current_wave", rollout.CurrentWave+1)
		m.hub.Broadcast("upgrade.status", gin.H{"rollout_id": rollout.ID, "status": "running", "current_wave": rollout.CurrentWave + 1})
		m.notify()
		return
	}
	for i := range tasks {
		task := &tasks[i]
		if task.Status == "deploying" || task.Status == "restarting" {
			if task.StartedAt != nil && time.Since(*task.StartedAt) > 10*time.Minute {
				now := time.Now()
				m.db.Model(task).Updates(map[string]any{"status": "failed", "last_error": "升级后 10 分钟内未确认目标版本", "completed_at": now})
				m.notify()
			}
			continue
		}
		if task.Status != "pending" {
			continue
		}
		var node store.Node
		if m.db.First(&node, "id = ?", task.NodeID).Error != nil {
			continue
		}
		if node.Version == task.TargetVersion {
			now := time.Now()
			m.db.Model(task).Updates(map[string]any{"status": "succeeded", "confirmed_version": node.Version, "completed_at": now})
			m.notify()
			continue
		}
		if !m.agents.Online(node.ID) {
			continue
		}
		var build store.AgentBuild
		if m.db.First(&build, "id = ?", task.BuildID).Error != nil {
			continue
		}
		now := time.Now()
		result := m.db.Model(task).Where("status = ?", "pending").Updates(map[string]any{"status": "deploying", "attempt": gorm.Expr("attempt + 1"), "started_at": now})
		if result.RowsAffected != 1 {
			continue
		}
		m.hub.Broadcast("upgrade.status", gin.H{"rollout_id": rollout.ID, "task_id": task.ID, "node_id": node.ID, "status": "deploying"})
		payload := agentupdate.Command{
			TaskID: task.ID, RolloutID: rollout.ID, Version: task.TargetVersion, OS: build.OS, Arch: build.Arch,
			URL: "/agent/v1/updates/" + build.ID, SHA256: build.SHA256, Signature: build.Signature, Size: build.Size,
		}
		if err := m.agents.SendUpgrade(node.ID, payload); err != nil {
			m.db.Model(task).Updates(map[string]any{"status": "pending", "last_error": sanitizeUpgradeError(err.Error()), "started_at": nil})
		}
	}
}

func isUpgradeFailure(status string) bool {
	return status == "failed" || status == "rolled_back" || status == "rollback_failed"
}

func (m *UpgradeManager) completeRollout(rollout *store.UpgradeRollout) {
	now := time.Now()
	m.db.Model(rollout).Updates(map[string]any{"status": "completed", "completed_at": now})
	m.hub.Broadcast("upgrade.status", gin.H{"rollout_id": rollout.ID, "status": "completed"})
}

func (a *API) listAgentReleases(c *gin.Context) {
	var releases []store.AgentRelease
	a.db.Order("created_at DESC").Find(&releases)
	items := make([]gin.H, 0, len(releases))
	for _, release := range releases {
		var builds []store.AgentBuild
		a.db.Where("release_id = ?", release.ID).Order("os,arch").Find(&builds)
		availableBuilds := 0
		for _, build := range builds {
			if _, err := verifyAgentBuildArtifact(a.cfg, build, false); err == nil {
				availableBuilds++
			}
		}
		items = append(items, gin.H{"release": release, "builds": builds, "available": len(builds) > 0 && availableBuilds == len(builds), "available_builds": availableBuilds})
	}
	ok(c, items)
}

func (a *API) scanAgentRelease(c *gin.Context) {
	var req struct {
		Version string `json:"version"`
		Notes   string `json:"notes"`
	}
	if c.ShouldBindJSON(&req) != nil && c.Request.ContentLength > 0 {
		fail(c, 400, "INVALID_ARGUMENT", "发布信息格式错误")
		return
	}
	release, builds, err := a.upgrades.ScanRelease(req.Version, req.Notes)
	if err != nil {
		fail(c, 400, "RELEASE_SCAN_FAILED", err.Error())
		return
	}
	created(c, gin.H{"release": release, "builds": builds, "public_key": a.updateSigner.PublicKey()})
}

func (a *API) listUpgradeRollouts(c *gin.Context) {
	var rollouts []store.UpgradeRollout
	a.db.Order("created_at DESC").Find(&rollouts)
	items := make([]gin.H, 0, len(rollouts))
	for _, rollout := range rollouts {
		var tasks []store.UpgradeTask
		a.db.Where("rollout_id = ?", rollout.ID).Order("wave,created_at").Find(&tasks)
		nodeIDs := make([]string, 0, len(tasks))
		for _, task := range tasks {
			nodeIDs = append(nodeIDs, task.NodeID)
		}
		var nodes []store.Node
		if len(nodeIDs) > 0 {
			a.db.Select("id", "name", "group_name", "status", "version", "ip", "public_ip", "os", "arch").Where("id IN ?", nodeIDs).Find(&nodes)
		}
		nodeByID := make(map[string]store.Node, len(nodes))
		for _, node := range nodes {
			nodeByID[node.ID] = node
		}
		taskViews := make([]gin.H, 0, len(tasks))
		counts := map[string]int{}
		for _, task := range tasks {
			counts[task.Status]++
			taskViews = append(taskViews, upgradeTaskView(task, nodeByID[task.NodeID]))
		}
		finished := counts["succeeded"] + counts["failed"] + counts["rolled_back"] + counts["rollback_failed"] + counts["cancelled"]
		percent := 0
		if len(tasks) > 0 {
			percent = finished * 100 / len(tasks)
		}
		items = append(items, gin.H{"rollout": rollout, "tasks": taskViews, "counts": counts, "progress": gin.H{"total": len(tasks), "finished": finished, "percent": percent}})
	}
	ok(c, items)
}

func upgradeTaskView(task store.UpgradeTask, node store.Node) gin.H {
	data, _ := json.Marshal(task)
	view := gin.H{}
	_ = json.Unmarshal(data, &view)
	view["node"] = gin.H{
		"id": node.ID, "name": node.Name, "group": node.GroupName, "status": node.Status,
		"version": node.Version, "ip": node.IP, "public_ip": node.PublicIP, "os": node.OS, "arch": node.Arch,
	}
	return view
}

func (a *API) createUpgradeRollout(c *gin.Context) {
	var req createRolloutInput
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "灰度发布参数格式错误")
		return
	}
	rollout, err := a.upgrades.CreateRollout(req, currentUser(c).ID)
	if err != nil {
		fail(c, 400, "ROLLOUT_CREATE_FAILED", err.Error())
		return
	}
	created(c, rollout)
}

func (a *API) updateUpgradeRolloutStatus(status string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !map[string]bool{"running": true, "paused": true, "cancelled": true}[status] {
			fail(c, 400, "INVALID_STATUS", "升级状态无效")
			return
		}
		updates := map[string]any{"status": status}
		if status == "running" {
			a.db.Model(&store.UpgradeTask{}).Where("rollout_id = ? AND status IN ?", c.Param("id"), []string{"failed", "rolled_back", "rollback_failed"}).Updates(map[string]any{
				"status": "pending", "started_at": nil, "completed_at": nil, "last_error": "",
			})
		}
		if status == "cancelled" {
			now := time.Now()
			updates["completed_at"] = now
			a.db.Model(&store.UpgradeTask{}).Where("rollout_id = ? AND status = ?", c.Param("id"), "pending").Updates(map[string]any{"status": "cancelled", "completed_at": now})
		}
		result := a.db.Model(&store.UpgradeRollout{}).Where("id = ? AND status IN ?", c.Param("id"), []string{"running", "paused"}).Updates(updates)
		if result.RowsAffected == 0 {
			fail(c, 404, "ROLLOUT_NOT_FOUND", "灰度发布不存在或已结束")
			return
		}
		a.upgrades.notify()
		a.hub.Broadcast("upgrade.status", gin.H{"rollout_id": c.Param("id"), "status": status})
		okEmpty(c)
	}
}

func (a *API) downloadAgentUpdate(c *gin.Context) {
	var build store.AgentBuild
	if a.db.First(&build, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "升级构建不存在")
		return
	}
	node := c.MustGet("agent.node").(store.Node)
	if node.OS != build.OS || node.Arch != build.Arch {
		fail(c, 403, "PLATFORM_MISMATCH", "升级构建与节点平台不匹配")
		return
	}
	path, err := verifyAgentBuildArtifact(a.cfg, build, false)
	if err != nil {
		fail(c, 404, "BUILD_NOT_FOUND", "升级构建文件不存在")
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.FileAttachment(path, build.Filename)
}

func verifyAgentBuildArtifact(cfg config.Config, build store.AgentBuild, checksum bool) (string, error) {
	if filepath.Base(build.Filename) != build.Filename || build.Filename == "" {
		return "", errors.New("invalid build filename")
	}
	base := cfg.DownloadsDir
	if strings.HasPrefix(build.Filename, "honeynet-agent-release-") {
		base = filepath.Join(cfg.PKIDir, "agent-releases")
	}
	base, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	path := filepath.Join(base, build.Filename)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() != build.Size {
		return "", errors.New("build size does not match signed metadata")
	}
	if !checksum {
		return path, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), build.SHA256) {
		return "", errors.New("build SHA-256 does not match signed metadata")
	}
	return path, nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func sanitizeUpgradeError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 900 {
		value = value[:900]
	}
	return value
}
