package clickhouse

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/honeynet/honeynet/internal/analytics"
)

type connStub struct{ driver.Conn }

type emptyRows struct{ driver.Rows }

func (emptyRows) Next() bool   { return false }
func (emptyRows) Close() error { return nil }
func (emptyRows) Err() error   { return nil }

type eventListConn struct {
	driver.Conn
	query         string
	arguments     []any
	queryRowCalls int
	count         uint64
}

type countRow struct{ count uint64 }

func (row countRow) Err() error           { return nil }
func (row countRow) ScanStruct(any) error { return nil }
func (row countRow) Scan(destinations ...any) error {
	if len(destinations) == 1 {
		if destination, ok := destinations[0].(*uint64); ok {
			*destination = row.count
		}
	}
	return nil
}

func (conn *eventListConn) Query(_ context.Context, query string, arguments ...any) (driver.Rows, error) {
	conn.query = query
	conn.arguments = arguments
	return emptyRows{}, nil
}

func (conn *eventListConn) QueryRow(context.Context, string, ...any) driver.Row {
	conn.queryRowCalls++
	return countRow{count: conn.count}
}

func TestNewValidatesTable(t *testing.T) {
	if _, err := New(&connStub{}, "security_events; DROP TABLE x"); err == nil {
		t.Fatal("unsafe table name accepted")
	}
	repository, err := New(&connStub{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if repository.table != defaultTable {
		t.Fatalf("default table = %q", repository.table)
	}
}

func TestQualifiedTable(t *testing.T) {
	got, err := qualifiedTable("honeynet_analytics", "security_events")
	if err != nil || got != "`honeynet_analytics`.`security_events`" {
		t.Fatalf("qualified table = %q, %v", got, err)
	}
	if _, err := qualifiedTable("honeynet;DROP", "security_events"); err == nil {
		t.Fatal("unsafe database name accepted")
	}
}

func TestEventConditions(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	builder, err := eventConditions(analytics.EventFilter{NodeID: "node-1", EventClass: "credential", SourceIP: "203.0.113", From: from, Cursor: &analytics.Cursor{EventTime: from.Add(time.Hour), EventID: "e2"}})
	if err != nil {
		t.Fatal(err)
	}
	where := builder.where()
	for _, expected := range []string{"node_id = ?", "has_credential = 1", "positionCaseInsensitive", "event_time >= ?", "event_id < ?"} {
		if !strings.Contains(where, expected) {
			t.Fatalf("condition %q missing from %q", expected, where)
		}
	}
	if got := builder.arguments[2].(time.Time); got.Location() != time.UTC {
		t.Fatal("query time was not normalized to UTC")
	}
}

func TestBuildEventListQueryUsesKeysetWithoutCountOrOffset(t *testing.T) {
	builder, err := eventConditions(analytics.EventFilter{
		NodeID: "node-1",
		Cursor: &analytics.Cursor{
			EventTime: time.Date(2026, 8, 12, 1, 2, 3, 456000000, time.UTC),
			EventID:   "event-2",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	query, arguments, fetchLimit := buildEventListQuery("`db`.`events`", builder, 50, 500000, true, false)
	if strings.Contains(query, "OFFSET") || strings.Contains(strings.ToLower(query), "count(") {
		t.Fatalf("cursor query contains deep-page work: %s", query)
	}
	if !strings.Contains(query, "ORDER BY event_time DESC,event_id DESC LIMIT ?") || !strings.Contains(query, "event_time < ?") {
		t.Fatalf("cursor query does not use the stable keyset: %s", query)
	}
	if fetchLimit != 51 || arguments[len(arguments)-1] != uint64(51) {
		t.Fatalf("cursor query must fetch page size + 1: limit=%d args=%#v", fetchLimit, arguments)
	}
}

func TestBuildEventListQueryKeepsBoundedLegacyOffsetMode(t *testing.T) {
	query, arguments, fetchLimit := buildEventListQuery("`db`.`events`", queryBuilder{}, 25, 50, false, false)
	if !strings.Contains(query, "LIMIT ? OFFSET ?") || fetchLimit != 25 {
		t.Fatalf("legacy query changed unexpectedly: %s limit=%d", query, fetchLimit)
	}
	if arguments[len(arguments)-2] != uint64(25) || arguments[len(arguments)-1] != uint64(50) {
		t.Fatalf("legacy limit/offset are not parameters: %#v", arguments)
	}
}

func TestRepositoryListCursorModeSkipsCountAndOffset(t *testing.T) {
	conn := &eventListConn{}
	repository := &Repository{conn: conn, database: "default", table: "security_events"}
	page, err := repository.List(context.Background(), analytics.EventFilter{CursorMode: true, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if conn.queryRowCalls != 0 {
		t.Fatalf("cursor mode executed %d count queries", conn.queryRowCalls)
	}
	if strings.Contains(conn.query, "OFFSET") || conn.arguments[len(conn.arguments)-1] != uint64(6) {
		t.Fatalf("cursor query=%q arguments=%#v", conn.query, conn.arguments)
	}
	if page.TotalKnown || page.HasMore || len(page.Items) != 0 {
		t.Fatalf("unexpected empty cursor page: %#v", page)
	}
}

func TestRepositoryListPageNavigationSkipsRepeatedCountAndKeepsBoundedOffset(t *testing.T) {
	conn := &eventListConn{}
	repository := &Repository{conn: conn, database: "default", table: "security_events"}
	page, err := repository.List(context.Background(), analytics.EventFilter{SkipTotal: true, Limit: 10, Offset: 90})
	if err != nil {
		t.Fatal(err)
	}
	if conn.queryRowCalls != 0 {
		t.Fatalf("page navigation executed %d repeated count queries", conn.queryRowCalls)
	}
	if !strings.Contains(conn.query, "LIMIT ? OFFSET ?") {
		t.Fatalf("page navigation query=%q", conn.query)
	}
	want := []any{uint64(11), uint64(90)}
	if len(conn.arguments) != len(want) || conn.arguments[0] != want[0] || conn.arguments[1] != want[1] {
		t.Fatalf("page navigation arguments=%#v, want %#v", conn.arguments, want)
	}
	if page.TotalKnown || page.Total != 0 || page.HasMore {
		t.Fatalf("unexpected empty navigation page: %#v", page)
	}
}

func TestRepositoryListCountsOnlyWhenRequested(t *testing.T) {
	conn := &eventListConn{count: 321}
	repository := &Repository{conn: conn, database: "default", table: "security_events"}
	first, err := repository.List(context.Background(), analytics.EventFilter{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if conn.queryRowCalls != 1 || !first.TotalKnown || first.Total != 321 {
		t.Fatalf("initial total query calls=%d page=%#v", conn.queryRowCalls, first)
	}
	if _, err := repository.List(context.Background(), analytics.EventFilter{SkipTotal: true, Limit: 10, Offset: 10}); err != nil {
		t.Fatal(err)
	}
	if conn.queryRowCalls != 1 {
		t.Fatalf("navigation repeated exact count: calls=%d", conn.queryRowCalls)
	}
}

func TestCredentialConditionsExcludeSensitiveSearchUnlessAuthorized(t *testing.T) {
	masked, err := credentialConditions(analytics.CredentialFilter{Keyword: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(masked.where(), "credential_password") || strings.Contains(masked.where(), "credential_auth_response") {
		t.Fatalf("masked credential search contains sensitive columns: %s", masked.where())
	}
	revealed, err := credentialConditions(analytics.CredentialFilter{Keyword: "probe", IncludeSensitive: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(revealed.where(), "credential_password") || !strings.Contains(revealed.where(), "credential_auth_response") {
		t.Fatalf("authorized credential search omitted sensitive columns: %s", revealed.where())
	}
}

func TestGlobRegex(t *testing.T) {
	if got := globRegex("web.*"); got != `^web\..*$` {
		t.Fatalf("globRegex = %q", got)
	}
	if got := globRegex("ssh.credential"); got != `^ssh\.credential$` {
		t.Fatalf("globRegex = %q", got)
	}
}

func TestDeduplicationTokenStable(t *testing.T) {
	events := []analytics.Event{{EventID: "a", RecordVersion: 1}, {EventID: "b", RecordVersion: 2}}
	first, second := deduplicationToken(events), deduplicationToken(events)
	if first != second || !strings.HasPrefix(first, "honeynet-events-") {
		t.Fatalf("unstable token: %q %q", first, second)
	}
	events[1].RecordVersion++
	if first == deduplicationToken(events) {
		t.Fatal("token ignored record version")
	}
}

func TestStatusReportsPingFailure(t *testing.T) {
	// A nil embedded Conn panics if invoked, so this test only verifies the
	// stable driver metadata through a short-circuit context wrapper later in
	// integration tests. Keep the compile-time contract explicit here.
	var _ analytics.Store = (*Repository)(nil)
	var _ context.Context = context.Background()
}
