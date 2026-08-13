package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/analytics"
)

func credentialConditions(filter analytics.CredentialFilter) (queryBuilder, error) {
	if err := analytics.ValidateRange(filter.From, filter.To); err != nil {
		return queryBuilder{}, err
	}
	builder := queryBuilder{}
	builder.add("has_credential = 1")
	if filter.Service != "" {
		builder.add("service = ?", filter.Service)
	}
	if !filter.From.IsZero() {
		builder.add("event_time >= ?", filter.From.UTC())
	}
	if !filter.To.IsZero() {
		builder.add("event_time <= ?", filter.To.UTC())
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		searchColumns := `concat(src_ip,' ',geo,' ',credential_username,' ',credential_mechanism,' ',service,' ',event_type)`
		if filter.IncludeSensitive {
			searchColumns = `concat(src_ip,' ',geo,' ',credential_username,' ',credential_password,' ',credential_auth_response,' ',credential_mechanism,' ',service,' ',event_type)`
		}
		builder.add(`positionCaseInsensitive(`+searchColumns+`, ?) > 0`, keyword)
	}
	return builder, nil
}

func (r *Repository) ListCredentials(ctx context.Context, filter analytics.CredentialFilter) (analytics.CredentialPage, error) {
	builder, err := credentialConditions(filter)
	if err != nil {
		return analytics.CredentialPage{}, err
	}
	var total uint64
	if err := r.conn.QueryRow(ctx, "SELECT count() FROM "+r.tableSQL()+" FINAL"+builder.where(), builder.arguments...).Scan(&total); err != nil {
		return analytics.CredentialPage{}, fmt.Errorf("count clickhouse credentials: %w", err)
	}
	limit := analytics.ClampLimit(filter.Limit, 100, 1000)
	query := "SELECT event_id,node_id,event_type,event_time,src_ip,geo,credential_username,credential_password,credential_auth_response,credential_mechanism,service FROM " + r.tableSQL() + " FINAL" + builder.where() + " ORDER BY event_time DESC,event_id DESC LIMIT ? OFFSET ?"
	arguments := append(append([]any(nil), builder.arguments...), uint64(limit), uint64(max(filter.Offset, 0)))
	rows, err := r.conn.Query(ctx, query, arguments...)
	if err != nil {
		return analytics.CredentialPage{}, fmt.Errorf("list clickhouse credentials: %w", err)
	}
	defer rows.Close()
	items := make([]analytics.CredentialResource, 0, limit)
	for rows.Next() {
		var item analytics.CredentialResource
		if err := rows.Scan(&item.EventID, &item.NodeID, &item.EventType, &item.EventTime, &item.SourceIP, &item.Geo, &item.Username, &item.Password, &item.AuthResponse, &item.Mechanism, &item.Service); err != nil {
			return analytics.CredentialPage{}, err
		}
		item.EventTime = item.EventTime.UTC()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return analytics.CredentialPage{}, err
	}
	usernames, err := r.credentialTop(ctx, builder, "credential_username")
	if err != nil {
		return analytics.CredentialPage{}, err
	}
	passwords, err := r.credentialTop(ctx, builder, "credential_password")
	if err != nil {
		return analytics.CredentialPage{}, err
	}
	return analytics.CredentialPage{Items: items, Total: total, TopUsernames: usernames, TopPasswords: passwords}, nil
}

func (r *Repository) credentialTop(ctx context.Context, builder queryBuilder, column string) ([]analytics.CredentialCount, error) {
	query := "SELECT " + column + ",count() AS count FROM " + r.tableSQL() + " FINAL" + builder.where() + " AND " + column + " != '' GROUP BY " + column + " ORDER BY count DESC," + column + " ASC LIMIT 10"
	rows, err := r.conn.Query(ctx, query, builder.arguments...)
	if err != nil {
		return nil, fmt.Errorf("query clickhouse credential top: %w", err)
	}
	defer rows.Close()
	items := []analytics.CredentialCount{}
	for rows.Next() {
		var item analytics.CredentialCount
		if err := rows.Scan(&item.Value, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) AttackerEvidence(ctx context.Context, sourceIP string, from, to time.Time, limit int) (analytics.AttackerEvidence, error) {
	if !analytics.ValidIP(sourceIP) {
		return analytics.AttackerEvidence{}, fmt.Errorf("invalid source IP")
	}
	if err := analytics.ValidateRange(from, to); err != nil {
		return analytics.AttackerEvidence{}, err
	}
	builder := rangeConditions(from, to)
	builder.add("src_ip = ?", sourceIP)
	evidence := analytics.AttackerEvidence{SourceIP: sourceIP}
	metaQuery := "SELECT count(),min(event_time),max(event_time),argMax(geo,event_time),argMax(asn,event_time) FROM " + r.tableSQL() + " FINAL" + builder.where()
	if err := r.conn.QueryRow(ctx, metaQuery, builder.arguments...).Scan(&evidence.Count, &evidence.FirstSeen, &evidence.LastSeen, &evidence.Geo, &evidence.ASN); err != nil {
		return analytics.AttackerEvidence{}, fmt.Errorf("query clickhouse attacker evidence: %w", err)
	}
	if evidence.Count == 0 {
		return analytics.AttackerEvidence{}, analytics.ErrNotFound
	}
	evidence.FirstSeen, evidence.LastSeen = evidence.FirstSeen.UTC(), evidence.LastSeen.UTC()
	services, err := r.namedCounts(ctx, builder, "service")
	if err != nil {
		return analytics.AttackerEvidence{}, err
	}
	types, err := r.namedCounts(ctx, builder, "event_type")
	if err != nil {
		return analytics.AttackerEvidence{}, err
	}
	evidence.Services, evidence.EventTypes = services, types
	page, err := r.List(ctx, analytics.EventFilter{ExactSourceIP: sourceIP, From: from, To: to, Limit: analytics.ClampLimit(limit, 100, 500)})
	if err != nil {
		return analytics.AttackerEvidence{}, err
	}
	evidence.Events = page.Items
	return evidence, nil
}

func (r *Repository) namedCounts(ctx context.Context, builder queryBuilder, column string) ([]analytics.NamedCount, error) {
	query := "SELECT " + column + ",count() AS count FROM " + r.tableSQL() + " FINAL" + builder.where() + " AND " + column + " != '' GROUP BY " + column + " ORDER BY count DESC," + column + " ASC LIMIT 100"
	rows, err := r.conn.Query(ctx, query, builder.arguments...)
	if err != nil {
		return nil, fmt.Errorf("query clickhouse named counts: %w", err)
	}
	defer rows.Close()
	items := []analytics.NamedCount{}
	for rows.Next() {
		var item analytics.NamedCount
		if err := rows.Scan(&item.Name, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CountAlertWindow(ctx context.Context, filter analytics.AlertWindowFilter) (uint64, error) {
	if err := analytics.ValidateRange(filter.From, filter.To); err != nil {
		return 0, err
	}
	builder := rangeConditions(filter.From, filter.To)
	if filter.Service != "" {
		builder.add("service = ?", filter.Service)
	}
	if filter.SourceIP != "" {
		builder.add("src_ip = ?", filter.SourceIP)
	} else if filter.NodeID != "" {
		builder.add("node_id = ?", filter.NodeID)
	}
	if pattern := strings.TrimSpace(filter.EventTypePattern); pattern != "" {
		builder.add("match(event_type, ?)", globRegex(pattern))
	}
	var count uint64
	if err := r.conn.QueryRow(ctx, "SELECT count() FROM "+r.tableSQL()+" FINAL"+builder.where(), builder.arguments...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count clickhouse alert window: %w", err)
	}
	return count, nil
}

func globRegex(pattern string) string {
	var result strings.Builder
	result.WriteByte('^')
	for _, character := range pattern {
		switch character {
		case '*':
			result.WriteString(".*")
		case '.', '+', '?', '(', ')', '[', ']', '{', '}', '^', '$', '|', '\\':
			result.WriteByte('\\')
			result.WriteRune(character)
		default:
			result.WriteRune(character)
		}
	}
	result.WriteByte('$')
	return result.String()
}
