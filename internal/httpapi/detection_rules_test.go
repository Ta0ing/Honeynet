package httpapi

import (
	"encoding/json"
	"testing"

	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/store"
)

func TestMergeDetectionHitsConfirmsAgentAndServerAgreement(t *testing.T) {
	server := []detection.Hit{{RuleKey: "builtin:CVE", Name: "CVE", Stage: "server"}}
	agent := []detection.Hit{{RuleKey: "builtin:CVE", Name: "CVE", Stage: "agent"}, {RuleKey: "custom:agent-only", Name: "Agent only", Stage: "agent"}}
	merged := mergeDetectionHits(server, agent)
	if len(merged) != 2 || !merged[0].Confirmed || merged[0].Stage != "server" {
		t.Fatalf("unexpected merged detections: %#v", merged)
	}
	if merged[1].RuleKey != "custom:agent-only" || merged[1].Confirmed || merged[1].Stage != "agent" {
		t.Fatalf("unexpected agent-only detection: %#v", merged[1])
	}
}

func TestDetectionRuleViewPreservesExactRevisionForBrowser(t *testing.T) {
	const revision int64 = 1786499376223386001
	view := detectionRuleView(store.DetectionRule{Revision: revision})
	if got := view["revision_text"]; got != "1786499376223386001" {
		t.Fatalf("revision_text = %v", got)
	}
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Fatal("empty JSON view")
	}
}

func TestDetectionPipelineNodeViewPreservesExactRevisionsForBrowser(t *testing.T) {
	const current int64 = 1786499376223386001
	const target int64 = 1786499376223386999
	view := detectionPipelineNodeView(store.Node{DetectionRuleRevision: current}, target, "stale")
	if view["revision_text"] != "1786499376223386001" || view["target_revision_text"] != "1786499376223386999" {
		t.Fatalf("pipeline revision strings are not exact: %#v", view)
	}
}

func TestDetectionRuleSetStatusUsesMaximumRevision(t *testing.T) {
	items := []store.DetectionRule{
		{Enabled: true, AgentEnabled: true, ServerEnabled: true, Revision: 10},
		{Enabled: true, AgentEnabled: false, ServerEnabled: true, Revision: 20},
		{Enabled: true, AgentEnabled: true, ServerEnabled: true, Revision: 30, ValidationError: "unsupported"},
	}
	revision, count := summarizeDetectionRules(items, true)
	if revision != 10 || count != 1 {
		t.Fatalf("agent rules = revision %d, count %d; want 10, 1", revision, count)
	}
	revision, count = summarizeDetectionRules(items, false)
	if revision != 20 || count != 2 {
		t.Fatalf("server rules = revision %d, count %d; want 20, 2", revision, count)
	}
}
