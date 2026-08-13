package sense

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type Status = protocol.SenseStatus

type Manager struct {
	mu         sync.Mutex
	sink       func(protocol.Event)
	config     protocol.SenseConfig
	status     Status
	capture    packetCapture
	cancel     context.CancelFunc
	generation uint64
}

func New(sink func(protocol.Event)) *Manager {
	return &Manager{sink: sink, status: Status{ActualStatus: "disabled"}}
}

func (m *Manager) Apply(parent context.Context, config protocol.SenseConfig) (Status, error) {
	normalized, err := NormalizeConfig(config)
	if err != nil {
		return m.setError(config, "error", err), err
	}
	m.mu.Lock()
	if reflect.DeepEqual(m.config, normalized) && (m.status.ActualStatus == "running" || !normalized.Enabled && m.status.ActualStatus == "disabled") {
		status := m.status
		m.mu.Unlock()
		return status, nil
	}
	m.stopLocked()
	m.config = normalized
	m.status = Status{Enabled: normalized.Enabled, Interface: normalized.Interface, ActualStatus: "disabled"}
	if !normalized.Enabled {
		status := m.status
		m.mu.Unlock()
		return status, nil
	}
	detector, detectorErr := NewDetector(normalized)
	if detectorErr != nil {
		m.mu.Unlock()
		return m.setError(normalized, "error", detectorErr), detectorErr
	}
	capture, captureErr := openCapture(normalized.Interface)
	if captureErr != nil {
		statusName := "error"
		if errors.Is(captureErr, ErrUnsupported) {
			statusName = "unsupported"
		}
		m.mu.Unlock()
		return m.setError(normalized, statusName, captureErr), captureErr
	}
	runContext, cancel := context.WithCancel(parent)
	m.capture, m.cancel = capture, cancel
	m.generation++
	generation := m.generation
	started := time.Now()
	m.status.ActualStatus = "running"
	m.status.StartedAt = &started
	status := m.status
	m.mu.Unlock()
	go m.run(runContext, generation, capture, detector)
	return status, nil
}

func (m *Manager) run(ctx context.Context, generation uint64, capture packetCapture, detector *Detector) {
	err := capture.Run(ctx, func(probe Probe) {
		now := time.Now()
		detection := detector.Observe(now, probe)
		m.mu.Lock()
		if generation != m.generation {
			m.mu.Unlock()
			return
		}
		m.status.ObservedPackets++
		if detection != nil {
			m.status.Detections++
			m.status.LastDetectionAt = &now
		}
		m.mu.Unlock()
		if detection != nil {
			m.emit(*detection)
		}
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	if generation != m.generation || ctx.Err() != nil {
		return
	}
	m.status.ActualStatus = "error"
	if err != nil {
		m.status.LastError = err.Error()
	} else {
		m.status.LastError = "packet capture stopped unexpectedly"
	}
}

func (m *Manager) emit(detection Detection) {
	payload := map[string]any{
		"mode": "passive", "protocol": detection.Protocol, "ports": detection.Ports,
		"distinct_ports": detection.DistinctPorts, "attempts": detection.Attempts,
		"first_seen": detection.FirstSeen.Unix(), "last_seen": detection.LastSeen.Unix(),
	}
	event := protocol.NewEvent("port.scan",
		protocol.Endpoint{IP: detection.SourceIP, Port: detection.SourcePort},
		protocol.Endpoint{IP: detection.DestIP, Port: detection.DestPort},
		payload, "scan", "passive", detection.Protocol,
	)
	event.Service = "sense"
	m.sink(event)
}

func (m *Manager) setError(config protocol.SenseConfig, statusName string, err error) Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
	m.config = config
	m.status = Status{Enabled: config.Enabled, Interface: config.Interface, ActualStatus: statusName, LastError: err.Error()}
	return m.status
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
	m.status.ActualStatus = "disabled"
}

func (m *Manager) stopLocked() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.capture != nil {
		_ = m.capture.Close()
	}
	m.cancel, m.capture = nil, nil
	m.generation++
}
