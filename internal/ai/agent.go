package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Tool is a read-only capability exposed to the security AI Agent. Tool
// implementations must enforce their own input limits and never return
// credentials or provider secrets. Mutating tools are intentionally excluded
// from the first agent runtime.
type Tool interface {
	Name() string
	Description() string
	Execute(context.Context, json.RawMessage) (any, error)
}

type ToolSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AgentRequest struct {
	Goal      string         `json:"goal"`
	Context   map[string]any `json:"context,omitempty"`
	ToolCalls []ToolCall     `json:"tool_calls,omitempty"`
	MaxSteps  int            `json:"max_steps,omitempty"`
}

type ToolCall struct {
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input,omitempty"`
}

type AgentStep struct {
	Index   int             `json:"index"`
	Tool    string          `json:"tool"`
	Input   json.RawMessage `json:"input,omitempty"`
	Output  any             `json:"output,omitempty"`
	Error   string          `json:"error,omitempty"`
	Planned bool            `json:"planned,omitempty"`
}

type AgentResult struct {
	Response Response    `json:"response"`
	Steps    []AgentStep `json:"steps"`
	Mode     string      `json:"mode"`
}

// Agent is the provider-neutral orchestration boundary. This first version
// deliberately executes only tools explicitly selected by the caller and
// performs a single evidence-grounded model synthesis. It gives later phases
// a stable place to add model-directed planning without mixing tool execution
// into HTTP handlers.
type Agent struct {
	analyzer Analyzer
	tools    map[string]Tool
}

func NewAgent(analyzer Analyzer, tools ...Tool) (*Agent, error) {
	if analyzer == nil {
		return nil, errors.New("AI Agent analyzer is required")
	}
	result := &Agent{analyzer: analyzer, tools: make(map[string]Tool, len(tools))}
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		name := strings.TrimSpace(tool.Name())
		if name == "" {
			return nil, errors.New("AI Agent tool name is required")
		}
		if _, exists := result.tools[name]; exists {
			return nil, fmt.Errorf("duplicate AI Agent tool %q", name)
		}
		result.tools[name] = tool
	}
	return result, nil
}

func (a *Agent) Tools() []ToolSpec {
	items := make([]ToolSpec, 0, len(a.tools))
	for name, tool := range a.tools {
		items = append(items, ToolSpec{Name: name, Description: strings.TrimSpace(tool.Description())})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func (a *Agent) Run(ctx context.Context, request AgentRequest) (AgentResult, error) {
	goal := strings.TrimSpace(request.Goal)
	if goal == "" {
		return AgentResult{}, errors.New("AI Agent goal is required")
	}
	if len(goal) > 4000 {
		return AgentResult{}, errors.New("AI Agent goal is too long")
	}
	maxSteps := request.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 4
	}
	if maxSteps > 8 {
		return AgentResult{}, errors.New("AI Agent max_steps must not exceed 8")
	}
	steps := make([]AgentStep, 0, min(len(request.ToolCalls), maxSteps))
	evidence := map[string]any{"provided_context": request.Context, "tool_results": map[string]any{}}
	results := evidence["tool_results"].(map[string]any)
	seen := make(map[string]struct{}, len(request.ToolCalls))
	for _, call := range request.ToolCalls {
		name := strings.TrimSpace(call.Name)
		if name == "" {
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			continue
		}
		seen[name] = struct{}{}
		if len(steps) >= maxSteps {
			break
		}
		tool, exists := a.tools[name]
		if !exists {
			return AgentResult{}, fmt.Errorf("AI Agent tool %q is not available", name)
		}
		step := AgentStep{Index: len(steps) + 1, Tool: name, Input: call.Input, Planned: true}
		output, err := tool.Execute(ctx, call.Input)
		if err != nil {
			step.Error = err.Error()
			steps = append(steps, step)
			return AgentResult{Steps: steps, Mode: "bounded-read-only"}, fmt.Errorf("execute AI Agent tool %s: %w", name, err)
		}
		step.Output = output
		steps = append(steps, step)
		results[name] = output
	}
	response, err := a.analyzer.Analyze(ctx, Request{Task: goal, Evidence: evidence})
	if err != nil {
		return AgentResult{Steps: steps, Mode: "bounded-read-only"}, err
	}
	return AgentResult{Response: response, Steps: steps, Mode: "bounded-read-only"}, nil
}
