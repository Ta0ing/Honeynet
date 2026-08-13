package ai

import (
	"context"
	"encoding/json"
	"testing"
)

type agentAnalyzerStub struct{ request Request }

func (s *agentAnalyzerStub) Analyze(_ context.Context, request Request) (Response, error) {
	s.request = request
	return Response{Provider: "stub", Model: "stub", Content: "ok"}, nil
}
func (s *agentAnalyzerStub) Status() Status { return Status{Enabled: true, Configured: true} }

type agentToolStub struct{ name string }

func (s agentToolStub) Name() string        { return s.name }
func (s agentToolStub) Description() string { return "只读测试工具" }
func (s agentToolStub) Execute(context.Context, json.RawMessage) (any, error) {
	return map[string]any{"count": 2}, nil
}

func TestAgentRunsAllowlistedReadOnlyToolsBeforeAnalysis(t *testing.T) {
	analyzer := &agentAnalyzerStub{}
	agent, err := NewAgent(analyzer, agentToolStub{name: "recent-events"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := agent.Run(context.Background(), AgentRequest{Goal: "研判攻击活动", ToolCalls: []ToolCall{{Name: "recent-events"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mode != "bounded-read-only" || len(result.Steps) != 1 || result.Steps[0].Tool != "recent-events" {
		t.Fatalf("unexpected agent result: %#v", result)
	}
	if _, ok := analyzer.request.Evidence["tool_results"].(map[string]any)["recent-events"]; !ok {
		t.Fatalf("tool evidence was not passed to analyzer: %#v", analyzer.request.Evidence)
	}
}

func TestAgentRejectsUnknownTool(t *testing.T) {
	agent, err := NewAgent(&agentAnalyzerStub{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = agent.Run(context.Background(), AgentRequest{Goal: "test", ToolCalls: []ToolCall{{Name: "shell"}}})
	if err == nil {
		t.Fatal("expected unknown tool rejection")
	}
}

func TestAgentRejectsInvalidGoalAndStepLimit(t *testing.T) {
	agent, _ := NewAgent(&agentAnalyzerStub{})
	if _, err := agent.Run(context.Background(), AgentRequest{}); err == nil {
		t.Fatal("expected empty goal rejection")
	}
	if _, err := agent.Run(context.Background(), AgentRequest{Goal: "test", MaxSteps: 9}); err == nil {
		t.Fatal("expected max step rejection")
	}
}
