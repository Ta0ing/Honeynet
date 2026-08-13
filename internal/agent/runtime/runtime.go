package runtime

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/honeynet/honeynet/internal/agent/pots"
	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type instance struct {
	target  protocol.PotTarget
	service pots.Service
	status  string
}
type Runtime struct {
	mu        sync.Mutex
	instances map[string]*instance
	sink      pots.Sink
	tls       pots.TLSProvider
	templates string
}

type serviceProvider struct {
	tls       pots.TLSProvider
	templates string
}

func (p serviceProvider) TLSConfig() *tls.Config {
	if p.tls == nil {
		return nil
	}
	return p.tls.TLSConfig()
}

func (p serviceProvider) TemplateRoot() string { return p.templates }

func New(sink pots.Sink) *Runtime { return &Runtime{instances: map[string]*instance{}, sink: sink} }
func NewWithTLS(sink pots.Sink, provider pots.TLSProvider) *Runtime {
	return &Runtime{instances: map[string]*instance{}, sink: sink, tls: provider}
}
func NewWithTLSAndTemplates(sink pots.Sink, provider pots.TLSProvider, templateRoot string) *Runtime {
	return &Runtime{instances: map[string]*instance{}, sink: sink, tls: provider, templates: templateRoot}
}

func (r *Runtime) Apply(ctx context.Context, targets []protocol.PotTarget) []protocol.PotResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	desired := make(map[string]protocol.PotTarget, len(targets))
	for _, target := range targets {
		desired[target.ID] = target
	}
	results := make([]protocol.PotResult, 0, len(targets))
	for id, current := range r.instances {
		target, exists := desired[id]
		if !exists || target.DesiredStatus != "running" || !sameTarget(current.target, target) {
			_ = current.service.Stop()
			delete(r.instances, id)
			if !exists {
				continue
			}
		}
	}
	for _, target := range targets {
		if target.DesiredStatus != "running" {
			results = append(results, protocol.PotResult{PotID: target.ID, Status: "stopped", Success: true})
			continue
		}
		if current := r.instances[target.ID]; current != nil {
			results = append(results, protocol.PotResult{PotID: target.ID, Status: current.status, Success: true})
			continue
		}
		provider := serviceProvider{tls: r.tls, templates: r.templates}
		service, err := pots.New(target.Service, provider)
		unsupported := errors.Is(err, pots.ErrUnsupportedService)
		if err == nil {
			sink := func(event protocol.Event) {
				event.PotID = target.ID
				event.Service = target.Service
				if event.Dst.Port == 0 {
					event.Dst.Port = target.Port
				}
				r.sink(event)
			}
			err = service.Start(ctx, target, sink)
		}
		if err != nil {
			status := "error"
			if unsupported {
				status = "unsupported"
			}
			results = append(results, protocol.PotResult{PotID: target.ID, Status: status, Success: false, Error: err.Error()})
			continue
		}
		r.instances[target.ID] = &instance{target: target, service: service, status: "running"}
		results = append(results, protocol.PotResult{PotID: target.ID, Status: "running", Success: true})
	}
	return results
}

func (r *Runtime) Statuses() []map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]map[string]string, 0, len(r.instances))
	for id, item := range r.instances {
		out = append(out, map[string]string{"id": id, "status": item.status})
	}
	return out
}
func (r *Runtime) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, item := range r.instances {
		_ = item.service.Stop()
		delete(r.instances, id)
	}
}
func sameTarget(a, b protocol.PotTarget) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}
func (r *Runtime) Count() int     { r.mu.Lock(); defer r.mu.Unlock(); return len(r.instances) }
func (r *Runtime) String() string { return fmt.Sprintf("%d pot instances", r.Count()) }
