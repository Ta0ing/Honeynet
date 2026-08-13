package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	aimodule "github.com/honeynet/honeynet/internal/ai"
	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func harnessRefs() []harnessEvidenceRef {
	refs := make([]harnessEvidenceRef, 0, 20)
	for index := 0; index < 10; index++ {
		refs = append(refs, harnessEvidenceRef{EventID: "malicious-" + string(rune('a'+index)), Label: "malicious"})
	}
	for index := 0; index < 10; index++ {
		refs = append(refs, harnessEvidenceRef{EventID: "benign-" + string(rune('a'+index)), Label: "benign"})
	}
	return refs
}

func TestSplitHarnessDatasetIsDeterministicStratifiedAndDisjoint(t *testing.T) {
	refs := harnessRefs()
	first, second := splitHarnessDataset(refs), splitHarnessDataset(refs)
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("dataset split is not deterministic: %s != %s", firstJSON, secondJSON)
	}
	if len(first.Training)+len(first.Evaluation) != len(refs) {
		t.Fatalf("split lost samples: %#v", first)
	}
	seen := map[string]string{}
	for _, ref := range first.Training {
		seen[ref.EventID] = "training"
	}
	evalClasses := map[string]int{}
	for _, ref := range first.Evaluation {
		if seen[ref.EventID] != "" {
			t.Fatalf("sample %s leaked into both splits", ref.EventID)
		}
		seen[ref.EventID] = "evaluation"
		evalClasses[ref.Label]++
	}
	if evalClasses["malicious"] < harnessMinimumEvalClass || evalClasses["benign"] < harnessMinimumEvalClass {
		t.Fatalf("holdout is not stratified: %#v", evalClasses)
	}
}

func TestBuildHarnessModelEvidenceContainsOnlyTrainingAndRedactsSecrets(t *testing.T) {
	training := []harnessEvidenceRef{{EventID: "train-1", Label: "malicious"}}
	events := []store.AttackEvent{{EventID: "train-1", EventType: "web.request", Service: "http", RawPacket: "Authorization: super-secret", Payload: datatypes.JSON(`{"password":"super-secret","path":"/admin"}`)}}
	feedback := filterHarnessFeedback([]map[string]any{{"event_id": "hidden-eval-id", "verdict": "false_positive"}, {"event_id": "historical-1", "verdict": "false_positive"}}, []harnessEvidenceRef{{EventID: "hidden-eval-id", Label: "benign"}})
	evidence := buildHarnessModelEvidence(events, training, nil, feedback, map[string]any{"hidden_evaluation_samples": 4})
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	value := string(encoded)
	if strings.Contains(value, "super-secret") || strings.Contains(value, "hidden-eval-id") {
		t.Fatalf("model evidence leaked raw/sensitive or holdout evidence: %s", value)
	}
	if !strings.Contains(value, "train-1") {
		t.Fatalf("training evidence missing: %s", value)
	}
	if !strings.Contains(value, "historical-1") {
		t.Fatalf("unrelated historical feedback missing: %s", value)
	}
}

func TestEvaluateHarnessCandidateUsesHiddenHoldoutMetrics(t *testing.T) {
	refs := []harnessEvidenceRef{
		{EventID: "m1", Label: "malicious"}, {EventID: "m2", Label: "malicious"},
		{EventID: "b1", Label: "benign"}, {EventID: "b2", Label: "benign"},
	}
	events := []store.AttackEvent{
		{EventID: "m1", RawPacket: "GET /?q=exploit"}, {EventID: "m2", RawPacket: "POST / exploit"},
		{EventID: "b1", RawPacket: "GET /health"}, {EventID: "b2", RawPacket: "GET /index"},
	}
	candidate := ruleCandidate{Key: "ai:test", Name: "测试规则", Severity: "high", AgentEnabled: true, ServerEnabled: true, Patterns: []detection.Pattern{{ID: "payload", Field: "raw", Operator: "contains", Value: "exploit", MinCount: 1}}}
	result, err := evaluateHarnessCandidate(candidate, events, refs, 16, 20)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "passed" || result.SampleCount != 4 || result.TrainingSampleCount != 16 || result.TotalSampleCount != 20 || result.Precision != 1 || result.Recall != 1 || result.FalsePositiveRate != 0 {
		t.Fatalf("unexpected holdout evaluation: %#v", result)
	}
}

func TestHarnessBaselineGateRequiresNonDegradingMeasurableImprovement(t *testing.T) {
	identical := harnessEvaluation{Status: "passed", Precision: .9, Recall: .8, FalsePositiveRate: .05}
	applyHarnessBaselineGate(&identical, harnessEvaluation{Status: "passed", Precision: .9, Recall: .8, FalsePositiveRate: .05})
	if identical.Status != "failed" || !strings.Contains(identical.Reason, "未优于") {
		t.Fatalf("identical candidate passed baseline gate: %#v", identical)
	}
	degraded := harnessEvaluation{Status: "passed", Precision: .85, Recall: .9, FalsePositiveRate: .05}
	applyHarnessBaselineGate(&degraded, harnessEvaluation{Status: "passed", Precision: .9, Recall: .8, FalsePositiveRate: .05})
	if degraded.Status != "failed" || !strings.Contains(degraded.Reason, "退化") {
		t.Fatalf("degraded candidate passed baseline gate: %#v", degraded)
	}
	improved := harnessEvaluation{Status: "passed", Precision: .95, Recall: .85, FalsePositiveRate: .04}
	applyHarnessBaselineGate(&improved, harnessEvaluation{Status: "passed", Precision: .9, Recall: .8, FalsePositiveRate: .05})
	if improved.Status != "passed" || improved.Improvement == "" || !improved.BaselineEvaluated {
		t.Fatalf("measurably improved candidate failed baseline gate: %#v", improved)
	}
}

func TestNormalizeHarnessCandidateRejectsInactiveOrOversizedRule(t *testing.T) {
	candidate := ruleCandidate{Key: "ai:test", Name: "test", Severity: "high", Patterns: []detection.Pattern{{ID: "p", Field: "raw", Operator: "contains", Value: "x"}}}
	if err := normalizeHarnessCandidate(&candidate, "create", ""); err == nil {
		t.Fatal("inactive candidate was accepted")
	}
	candidate.AgentEnabled = true
	candidate.Patterns = make([]detection.Pattern, 17)
	for index := range candidate.Patterns {
		candidate.Patterns[index] = detection.Pattern{ID: "p", Field: "raw", Operator: "contains", Value: "x"}
	}
	if err := normalizeHarnessCandidate(&candidate, "create", ""); err == nil {
		t.Fatal("oversized candidate was accepted")
	}
}

func TestDecodeHarnessProposalRejectsUnknownFieldsAndNonTrainingEvidence(t *testing.T) {
	base := `{"action":"create","title":"提案","rationale":"理由","candidate":{"key":"ai:test","name":"规则","description":"描述","severity":"high","agent_enabled":true,"server_enabled":true,"patterns":[{"id":"p","field":"raw","operator":"contains","value":"exploit","min_count":1}]},"evidence":[{"event_id":"train-1","label":"malicious"}]}`
	request := createHarnessRunRequest{Evidence: []harnessEvidenceRef{{EventID: "train-1", Label: "malicious"}}}
	if _, err := decodeHarnessProposal(aimodule.Response{JSON: json.RawMessage(base)}, request, nil); err != nil {
		t.Fatalf("valid proposal rejected: %v", err)
	}
	withUnknown := strings.TrimSuffix(base, "}") + `,"shell_command":"curl example"}`
	if _, err := decodeHarnessProposal(aimodule.Response{JSON: json.RawMessage(withUnknown)}, request, nil); err == nil {
		t.Fatal("unknown model output field was accepted")
	}
	withHoldout := strings.Replace(base, "train-1", "hidden-eval-1", 1)
	if _, err := decodeHarnessProposal(aimodule.Response{JSON: json.RawMessage(withHoldout)}, request, nil); err == nil {
		t.Fatal("proposal was allowed to cite hidden evaluation evidence")
	}
}

func TestHarnessProposalRequiresApprovalThenSupportsRollback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:harness-publish?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.DetectionRule{}, &store.DetectionRuleProposal{}, &store.AIHarnessRun{}, &store.Node{}); err != nil {
		t.Fatal(err)
	}
	run := store.AIHarnessRun{Base: store.NewBase(), Goal: "test", Kind: "rule-improvement", Status: "pending_review", Stage: "human_review", CreatedBy: store.NewBase().ID, Evidence: datatypes.JSON(`{}`), Trace: datatypes.JSON(`[]`), Result: datatypes.JSON(`{}`)}
	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}
	candidate, _ := json.Marshal(ruleCandidate{Key: "ai:approved", Name: "人工审批规则", Severity: "high", AgentEnabled: true, ServerEnabled: true, Patterns: []detection.Pattern{{ID: "p", Field: "raw", Operator: "contains", Value: "exploit", MinCount: 1}}})
	evaluation, _ := json.Marshal(harnessEvaluation{Status: "passed", TotalSampleCount: 20, TrainingSampleCount: 16, SampleCount: 4, PositiveCount: 2, NegativeCount: 2, Precision: 1, Recall: 1})
	proposal := store.DetectionRuleProposal{Base: store.NewBase(), RunID: run.ID, Action: "create", Status: "pending_review", Title: "test", Candidate: candidate, Baseline: datatypes.JSON(`{}`), Evidence: datatypes.JSON(`{}`), Evaluation: evaluation, CreatedBy: store.NewBase().ID}
	if err := db.Create(&proposal).Error; err != nil {
		t.Fatal(err)
	}
	api := &API{db: db, detection: &detectionRuntime{}, hub: NewHub()}
	api.agents = NewAgentGateway(db, api.hub)
	router := gin.New()
	router.POST("/proposals/:id/review", api.reviewDetectionRuleProposal)
	router.POST("/proposals/:id/publish", api.publishDetectionRuleProposal)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/proposals/"+proposal.ID+"/publish", nil))
	if recorder.Code != http.StatusConflict {
		t.Fatalf("unapproved proposal published: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	reviewRequest := httptest.NewRequest(http.MethodPost, "/proposals/"+proposal.ID+"/review", strings.NewReader(`{"decision":"approve","comment":"人工复核通过"}`))
	reviewRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, reviewRequest)
	if recorder.Code != http.StatusOK {
		t.Fatalf("review failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	staleReview := httptest.NewRecorder()
	staleRequest := httptest.NewRequest(http.MethodPost, "/proposals/"+proposal.ID+"/review", strings.NewReader(`{"decision":"reject"}`))
	staleRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(staleReview, staleRequest)
	if staleReview.Code != http.StatusConflict {
		t.Fatalf("stale review changed state: status=%d body=%s", staleReview.Code, staleReview.Body.String())
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/proposals/"+proposal.ID+"/publish", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("approved proposal did not publish: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var published store.DetectionRule
	if err := db.First(&published, "rule_key = ?", "ai:approved").Error; err != nil || !published.Enabled {
		t.Fatalf("published rule missing: %#v err=%v", published, err)
	}

	router.POST("/proposals/:id/rollback", api.rollbackDetectionRuleProposal)
	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/proposals/"+proposal.ID+"/rollback", strings.NewReader(`{"reason":"灰度反馈误报率上升"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("rollback failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if err := db.First(&published, "id = ?", published.ID).Error; err != nil || published.Enabled || published.AgentEnabled || published.ServerEnabled {
		t.Fatalf("rolled back created rule remained active: %#v err=%v", published, err)
	}
	var updated store.DetectionRuleProposal
	if err := db.First(&updated, "id = ?", proposal.ID).Error; err != nil || updated.Status != "rolled_back" || updated.RolledBackAt == nil {
		t.Fatalf("proposal rollback state invalid: %#v err=%v", updated, err)
	}
	if err := db.First(&run, "id = ?", run.ID).Error; err != nil || run.Status != "rolled_back" || !strings.Contains(string(run.Trace), "human_review") || !strings.Contains(string(run.Trace), "release") || !strings.Contains(string(run.Trace), "rollback") {
		t.Fatalf("persistent harness trace did not follow the state machine: %#v err=%v", run, err)
	}
}

func TestConcurrentHarnessReviewsHaveSingleStateTransition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+store.NewBase().ID+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&store.DetectionRuleProposal{}, &store.AIHarnessRun{}); err != nil {
		t.Fatal(err)
	}
	run := store.AIHarnessRun{Base: store.NewBase(), Goal: "test", Kind: "rule-improvement", Status: "pending_review", Stage: "human_review", CreatedBy: store.NewBase().ID, Evidence: datatypes.JSON(`{}`), Trace: datatypes.JSON(`[]`), Result: datatypes.JSON(`{}`)}
	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}
	proposal := store.DetectionRuleProposal{Base: store.NewBase(), RunID: run.ID, Action: "create", Status: "pending_review", Title: "test", Candidate: datatypes.JSON(`{}`), Baseline: datatypes.JSON(`{}`), Evidence: datatypes.JSON(`{}`), Evaluation: datatypes.JSON(`{}`), CreatedBy: store.NewBase().ID}
	if err := db.Create(&proposal).Error; err != nil {
		t.Fatal(err)
	}
	api := &API{db: db}
	router := gin.New()
	router.POST("/proposals/:id/review", api.reviewDetectionRuleProposal)

	statuses := make(chan int, 2)
	var wait sync.WaitGroup
	for _, decision := range []string{"approve", "reject"} {
		wait.Add(1)
		go func(decision string) {
			defer wait.Done()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/proposals/"+proposal.ID+"/review", strings.NewReader(`{"decision":"`+decision+`"}`))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)
			statuses <- recorder.Code
		}(decision)
	}
	wait.Wait()
	close(statuses)
	counts := map[int]int{}
	for status := range statuses {
		counts[status]++
	}
	if counts[http.StatusOK] != 1 || counts[http.StatusConflict] != 1 {
		t.Fatalf("concurrent reviews did not produce one winner: %#v", counts)
	}
	if err := db.First(&run, "id = ?", run.ID).Error; err != nil {
		t.Fatal(err)
	}
	var trace []map[string]any
	if json.Unmarshal(run.Trace, &trace) != nil || len(trace) != 1 || trace[0]["stage"] != "human_review" {
		t.Fatalf("concurrent review appended multiple state transitions: %s", run.Trace)
	}
}
