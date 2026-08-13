package httpapi

import (
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/analytics"
)

func TestParseEventTimeRange(t *testing.T) {
	from, to, err := parseEventTimeRange("2026-08-12T08:00:00+08:00", "2026-08-12T09:30:00.123+08:00")
	if err != nil {
		t.Fatalf("parse event time range: %v", err)
	}
	if got := from.Format(time.RFC3339Nano); got != "2026-08-12T00:00:00Z" {
		t.Fatalf("unexpected normalized from: %s", got)
	}
	if got := to.Format(time.RFC3339Nano); got != "2026-08-12T01:30:00.123Z" {
		t.Fatalf("unexpected normalized to: %s", got)
	}
}

func TestParseEventTimeRangeRejectsAmbiguousOrReversedRange(t *testing.T) {
	for _, values := range [][2]string{
		{"2026-08-12 08:00:00", ""},
		{"2026-08-12T09:00:00Z", "2026-08-12T08:00:00Z"},
	} {
		if _, _, err := parseEventTimeRange(values[0], values[1]); err == nil {
			t.Fatalf("expected %q..%q to fail", values[0], values[1])
		}
	}
}

func TestParseEventTimeRangeAllowsOpenBounds(t *testing.T) {
	from, to, err := parseEventTimeRange("", "")
	if err != nil || !from.IsZero() || !to.IsZero() {
		t.Fatalf("empty bounds should be accepted: from=%v to=%v err=%v", from, to, err)
	}
}

func TestParseEventPaginationCursorRoundTrip(t *testing.T) {
	want := analytics.Cursor{EventTime: time.Date(2026, 8, 12, 1, 2, 3, 456000000, time.UTC), EventID: "event-2"}
	token, err := analytics.EncodeCursor(want)
	if err != nil {
		t.Fatal(err)
	}
	first, err := parseEventPagination(url.Values{"pagination": {"cursor"}, "page_size": {"50"}})
	if err != nil || !first.CursorMode || first.Cursor != nil || first.PageSize != 50 {
		t.Fatalf("unexpected first cursor page: %#v err=%v", first, err)
	}
	next, err := parseEventPagination(url.Values{"pagination": {"cursor"}, "cursor": {token}, "page_size": {"1000"}})
	if err != nil || next.Cursor == nil || next.Cursor.EventID != want.EventID || !next.Cursor.EventTime.Equal(want.EventTime) || next.PageSize != 200 {
		t.Fatalf("unexpected next cursor page: %#v err=%v", next, err)
	}
}

func TestParseEventPaginationControlsExactTotal(t *testing.T) {
	first, err := parseEventPagination(url.Values{"pagination": {"page"}, "page": {"1"}, "page_size": {"10"}})
	if err != nil || !first.IncludeTotal || first.PageSize != 10 {
		t.Fatalf("first page should include exact total: %#v err=%v", first, err)
	}
	defaults, err := parseEventPagination(url.Values{})
	if err != nil || defaults.Page != 1 || defaults.PageSize != 10 || !defaults.IncludeTotal {
		t.Fatalf("default attack-event page should contain 10 rows and total: %#v err=%v", defaults, err)
	}
	next, err := parseEventPagination(url.Values{"pagination": {"page"}, "page": {"2"}, "page_size": {"10"}, "include_total": {"false"}})
	if err != nil || next.IncludeTotal || next.Page != 2 {
		t.Fatalf("page navigation should be able to reuse total: %#v err=%v", next, err)
	}
	for _, values := range []url.Values{
		{"pagination": {"page"}, "include_total": {"sometimes"}},
		{"pagination": {"cursor"}, "include_total": {"true"}},
	} {
		if _, err := parseEventPagination(values); !errors.Is(err, errInvalidEventPagination) {
			t.Fatalf("values=%v err=%v, want invalid pagination", values, err)
		}
	}
}

func TestParseEventPaginationRejectsBadCursorAndDeepPage(t *testing.T) {
	for _, values := range []url.Values{
		{"pagination": {"cursor"}, "cursor": {""}},
		{"cursor": {strings.Repeat("A", analytics.MaxCursorLength+1)}},
		{"pagination": {"page"}, "cursor": {"abc"}},
		{"pagination": {"cursor"}, "page": {"2"}},
		{"pagination": {"cursor"}, "page_size": {"nope"}},
		{"page": {"0"}},
	} {
		if _, err := parseEventPagination(values); !errors.Is(err, errInvalidEventPagination) {
			t.Fatalf("values=%v err=%v, want invalid pagination", values, err)
		}
	}
	if _, err := parseEventPagination(url.Values{"page": {"101"}}); !errors.Is(err, errEventPageTooDeep) {
		t.Fatalf("deep page err=%v", err)
	}
	legacy, err := parseEventPagination(url.Values{"page": {"100"}, "page_size": {"50"}})
	if err != nil || legacy.CursorMode || legacy.Page != 100 {
		t.Fatalf("bounded legacy page rejected: %#v err=%v", legacy, err)
	}
}
