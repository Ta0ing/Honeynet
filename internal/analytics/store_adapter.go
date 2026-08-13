package analytics

import (
	"time"

	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

// FromAttackEvent is the temporary compatibility boundary for existing
// handlers. New ingestion code should populate Event directly, including rule
// revisions supplied by the Agent and Server detection runtimes.
func FromAttackEvent(event store.AttackEvent) Event {
	return Event{
		EventID: event.EventID, NodeID: event.NodeID, PotID: event.PotID, DecoyID: event.DecoyID,
		Service: event.Service, EventType: event.EventType, EventTime: event.Timestamp, IngestedAt: event.CreatedAt,
		SourceIP: event.SrcIP, SourcePort: safePort(event.SrcPort), TargetIP: event.DstIP, TargetPort: safePort(event.DstPort),
		Geo: event.Geo, ASN: event.ASN, RawPacket: event.RawPacket, Payload: cloneJSON(event.Payload),
		Tags: cloneJSON(event.Tags), Detections: cloneJSON(event.Detections), AgentRuleRevision: event.AgentRuleRevision,
		ServerRuleRevision: event.ServerRuleRevision, SessionID: event.SessionID,
		RecordVersion: StableRecordVersion(event.EventID),
	}
}

func ToAttackEvent(event Event) store.AttackEvent {
	return store.AttackEvent{
		EventID: event.EventID, NodeID: event.NodeID, PotID: event.PotID, DecoyID: event.DecoyID,
		Service: event.Service, EventType: event.EventType, Timestamp: event.EventTime,
		SrcIP: event.SourceIP, SrcPort: int(event.SourcePort), DstIP: event.TargetIP, DstPort: int(event.TargetPort),
		Geo: event.Geo, ASN: event.ASN, RawPacket: event.RawPacket, Payload: datatypes.JSON(cloneJSON(event.Payload)),
		Tags: datatypes.JSON(cloneJSON(event.Tags)), Detections: datatypes.JSON(cloneJSON(event.Detections)),
		AgentRuleRevision: event.AgentRuleRevision, ServerRuleRevision: event.ServerRuleRevision,
		SessionID: event.SessionID, CreatedAt: event.IngestedAt,
	}
}

// RuleRevisions extracts the maximum revision observed at each matching stage.
// The original hit JSON remains untouched in Event.Detections.
func RuleRevisions(hits []detection.Hit, revisions map[string]int64) (agent, server int64) {
	for _, hit := range hits {
		revision := revisions[hit.RuleKey]
		switch hit.Stage {
		case "agent":
			if revision > agent {
				agent = revision
			}
		case "server":
			if revision > server {
				server = revision
			}
		}
	}
	return agent, server
}

func safePort(port int) uint16 {
	if port < 0 || port > 65535 {
		return 0
	}
	return uint16(port)
}

func cloneJSON(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return append([]byte(nil), value...)
}

func DayUTC(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
