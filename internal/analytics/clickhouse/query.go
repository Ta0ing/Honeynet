package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/analytics"
)

const (
	eventCountQueryTimeout = 8 * time.Second
	eventListQueryTimeout  = 15 * time.Second
)

type queryBuilder struct {
	conditions []string
	arguments  []any
}

func (b *queryBuilder) add(condition string, arguments ...any) {
	b.conditions = append(b.conditions, condition)
	b.arguments = append(b.arguments, arguments...)
}

func (b queryBuilder) where() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.conditions, " AND ")
}

func eventConditions(filter analytics.EventFilter) (queryBuilder, error) {
	if err := analytics.ValidateRange(filter.From, filter.To); err != nil {
		return queryBuilder{}, err
	}
	builder := queryBuilder{}
	if filter.NodeID != "" {
		builder.add("node_id = ?", filter.NodeID)
	}
	if filter.PotID != "" {
		builder.add("pot_id = ?", filter.PotID)
	}
	if filter.DecoyID != "" {
		builder.add("decoy_id = ?", filter.DecoyID)
	}
	if filter.Service != "" {
		builder.add("service = ?", filter.Service)
	}
	if filter.EventType != "" {
		builder.add("event_type = ?", filter.EventType)
	}
	switch filter.EventClass {
	case "credential":
		builder.add("has_credential = 1")
	case "web":
		builder.add("startsWith(event_type, 'web.')")
	case "decoy":
		builder.add("startsWith(event_type, 'decoy.')")
	case "":
	default:
		return queryBuilder{}, fmt.Errorf("unsupported event class %q", filter.EventClass)
	}
	if filter.SourceIP != "" {
		builder.add("positionCaseInsensitive(src_ip, ?) > 0", filter.SourceIP)
	}
	if filter.ExactSourceIP != "" {
		builder.add("src_ip = ?", filter.ExactSourceIP)
	}
	if !filter.From.IsZero() {
		builder.add("event_time >= ?", filter.From.UTC())
	}
	if !filter.To.IsZero() {
		builder.add("event_time <= ?", filter.To.UTC())
	}
	if filter.Cursor != nil {
		builder.add("(event_time < ? OR (event_time = ? AND event_id < ?))", filter.Cursor.EventTime.UTC(), filter.Cursor.EventTime.UTC(), filter.Cursor.EventID)
	}
	return builder, nil
}

func (r *Repository) List(ctx context.Context, filter analytics.EventFilter) (analytics.EventPage, error) {
	builder, err := eventConditions(filter)
	if err != nil {
		return analytics.EventPage{}, err
	}
	limit := analytics.ClampLimit(filter.Limit, 100, 1000)
	var total uint64
	totalKnown := !filter.CursorMode && !filter.SkipTotal
	if totalKnown {
		countBuilder, countErr := eventConditions(analytics.EventFilter{
			NodeID: filter.NodeID, PotID: filter.PotID, DecoyID: filter.DecoyID, Service: filter.Service,
			EventType: filter.EventType, EventClass: filter.EventClass, SourceIP: filter.SourceIP, From: filter.From, To: filter.To,
			ExactSourceIP: filter.ExactSourceIP,
		})
		if countErr != nil {
			return analytics.EventPage{}, countErr
		}
		countCtx, cancelCount := context.WithTimeout(ctx, eventCountQueryTimeout)
		err = r.conn.QueryRow(countCtx, "SELECT count() FROM "+r.tableSQL()+" FINAL"+countBuilder.where(), countBuilder.arguments...).Scan(&total)
		cancelCount()
		if err != nil {
			return analytics.EventPage{}, fmt.Errorf("count clickhouse events: %w", err)
		}
	}
	fetchExtra := filter.CursorMode || filter.SkipTotal
	query, arguments, fetchLimit := buildEventListQuery(r.tableSQL(), builder, limit, filter.Offset, filter.CursorMode, filter.SkipTotal)
	queryCtx, cancelQuery := context.WithTimeout(ctx, eventListQueryTimeout)
	defer cancelQuery()
	rows, err := r.conn.Query(queryCtx, query, arguments...)
	if err != nil {
		return analytics.EventPage{}, fmt.Errorf("list clickhouse events: %w", err)
	}
	defer rows.Close()
	items := make([]analytics.Event, 0, fetchLimit)
	for rows.Next() {
		event, scanErr := scanEvent(rows)
		if scanErr != nil {
			return analytics.EventPage{}, fmt.Errorf("scan clickhouse event: %w", scanErr)
		}
		items = append(items, event)
	}
	if err := rows.Err(); err != nil {
		return analytics.EventPage{}, fmt.Errorf("iterate clickhouse events: %w", err)
	}
	hasMore := false
	if fetchExtra {
		hasMore = len(items) > limit
		if hasMore {
			items = items[:limit]
		}
	} else {
		hasMore = uint64(max(filter.Offset, 0)+len(items)) < total
	}
	page := analytics.EventPage{Items: items, Total: total, TotalKnown: totalKnown, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		page.NextCursor = &analytics.Cursor{EventTime: last.EventTime, EventID: last.EventID}
	}
	return page, nil
}

func buildEventListQuery(table string, builder queryBuilder, limit, offset int, cursorMode, skipTotal bool) (string, []any, int) {
	base := "SELECT " + eventColumns + " FROM " + table + " FINAL" + builder.where() + " ORDER BY event_time DESC,event_id DESC LIMIT ?"
	arguments := append([]any(nil), builder.arguments...)
	if cursorMode || skipTotal {
		fetchLimit := limit + 1
		if skipTotal && offset > 0 {
			return base + " OFFSET ?", append(arguments, uint64(fetchLimit), uint64(offset)), fetchLimit
		}
		return base, append(arguments, uint64(fetchLimit)), fetchLimit
	}
	return base + " OFFSET ?", append(arguments, uint64(limit), uint64(max(offset, 0))), limit
}

func (r *Repository) Dashboard(ctx context.Context, from, to time.Time) (analytics.DashboardStats, error) {
	if err := analytics.ValidateRange(from, to); err != nil {
		return analytics.DashboardStats{}, err
	}
	builder := rangeConditions(from, to)
	var stats analytics.DashboardStats
	if err := r.conn.QueryRow(ctx, "SELECT count() FROM "+r.tableSQL()+" FINAL"+builder.where(), builder.arguments...).Scan(&stats.Events); err != nil {
		return analytics.DashboardStats{}, fmt.Errorf("query clickhouse dashboard: %w", err)
	}
	return stats, nil
}

func (r *Repository) Trends(ctx context.Context, from, to time.Time) ([]analytics.DayCount, error) {
	if err := analytics.ValidateRange(from, to); err != nil {
		return nil, err
	}
	builder := rangeConditions(from, to)
	rows, err := r.conn.Query(ctx, "SELECT toStartOfDay(event_time) AS day,count() AS count FROM "+r.tableSQL()+" FINAL"+builder.where()+" GROUP BY day ORDER BY day", builder.arguments...)
	if err != nil {
		return nil, fmt.Errorf("query clickhouse trends: %w", err)
	}
	defer rows.Close()
	items := []analytics.DayCount{}
	for rows.Next() {
		var item analytics.DayCount
		var day time.Time
		if err := rows.Scan(&day, &item.Count); err != nil {
			return nil, err
		}
		item.Day = day.UTC().Format("2006-01-02")
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) TopAttackers(ctx context.Context, from, to time.Time, limit int) ([]analytics.AttackerCount, error) {
	if err := analytics.ValidateRange(from, to); err != nil {
		return nil, err
	}
	builder := rangeConditions(from, to)
	builder.add("src_ip != ''")
	limit = analytics.ClampLimit(limit, 10, 1000)
	query := "SELECT src_ip,argMax(geo,event_time) AS geo,count() AS count,max(event_time) AS last_seen FROM " + r.tableSQL() + " FINAL" + builder.where() + " GROUP BY src_ip ORDER BY count DESC,last_seen DESC LIMIT ?"
	arguments := append(append([]any(nil), builder.arguments...), uint64(limit))
	rows, err := r.conn.Query(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query clickhouse top attackers: %w", err)
	}
	defer rows.Close()
	items := make([]analytics.AttackerCount, 0, limit)
	for rows.Next() {
		var item analytics.AttackerCount
		if err := rows.Scan(&item.SourceIP, &item.Geo, &item.Count, &item.LastSeen); err != nil {
			return nil, err
		}
		item.LastSeen = item.LastSeen.UTC()
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CountByNodes(ctx context.Context, nodeIDs []string) (map[string]uint64, error) {
	result := make(map[string]uint64, len(nodeIDs))
	if len(nodeIDs) == 0 {
		return result, nil
	}
	query := "SELECT node_id,count() FROM " + r.tableSQL() + " FINAL WHERE node_id IN (?) GROUP BY node_id"
	rows, err := r.conn.Query(ctx, query, nodeIDs)
	if err != nil {
		return nil, fmt.Errorf("count clickhouse events by node: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var nodeID string
		var count uint64
		if err := rows.Scan(&nodeID, &count); err != nil {
			return nil, err
		}
		result[nodeID] = count
	}
	return result, rows.Err()
}

func rangeConditions(from, to time.Time) queryBuilder {
	builder := queryBuilder{}
	if !from.IsZero() {
		builder.add("event_time >= ?", from.UTC())
	}
	if !to.IsZero() {
		builder.add("event_time <= ?", to.UTC())
	}
	return builder
}
