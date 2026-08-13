package httpapi

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/analytics"
)

const maxEventPage = 100

var (
	errInvalidEventPagination = errors.New("invalid event pagination")
	errEventPageTooDeep       = errors.New("event page is too deep")
)

type eventPagination struct {
	Page         int
	PageSize     int
	IncludeTotal bool
	CursorMode   bool
	Cursor       *analytics.Cursor
}

// parseEventPagination keeps the old bounded page contract while making
// keyset pagination explicit. An absent cursor in cursor mode means the first
// page; an explicitly empty/malformed cursor is rejected instead of silently
// restarting the result set.
func parseEventPagination(values url.Values) (eventPagination, error) {
	mode := strings.TrimSpace(values.Get("pagination"))
	_, cursorProvided := values["cursor"]
	if mode == "" && cursorProvided {
		mode = "cursor"
	}
	if mode != "" && mode != "page" && mode != "cursor" {
		return eventPagination{}, errInvalidEventPagination
	}
	if mode == "page" && cursorProvided {
		return eventPagination{}, errInvalidEventPagination
	}
	pageNumber, err := strictPositiveInt(values.Get("page"), 1)
	if err != nil {
		return eventPagination{}, errInvalidEventPagination
	}
	pageSize, err := strictPositiveInt(values.Get("page_size"), 10)
	if err != nil {
		return eventPagination{}, errInvalidEventPagination
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if mode == "cursor" {
		includeTotal, err := strictEventBoolean(values.Get("include_total"), false)
		if err != nil || includeTotal {
			return eventPagination{}, errInvalidEventPagination
		}
		if pageNumber > 1 {
			return eventPagination{}, errInvalidEventPagination
		}
		result := eventPagination{Page: 1, PageSize: pageSize, IncludeTotal: false, CursorMode: true}
		if cursorProvided {
			cursor, err := analytics.DecodeCursor(values.Get("cursor"))
			if err != nil {
				return eventPagination{}, errInvalidEventPagination
			}
			result.Cursor = &cursor
		}
		return result, nil
	}
	if pageNumber > maxEventPage {
		return eventPagination{}, errEventPageTooDeep
	}
	includeTotal, err := strictEventBoolean(values.Get("include_total"), true)
	if err != nil {
		return eventPagination{}, errInvalidEventPagination
	}
	return eventPagination{Page: pageNumber, PageSize: pageSize, IncludeTotal: includeTotal}, nil
}

func strictEventBoolean(value string, fallback bool) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return fallback, nil
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, errInvalidEventPagination
	}
}

func strictPositiveInt(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, errInvalidEventPagination
	}
	return parsed, nil
}

func encodeEventCursor(cursor *analytics.Cursor) (any, error) {
	if cursor == nil {
		return nil, nil
	}
	encoded, err := analytics.EncodeCursor(*cursor)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// parseEventTimeRange keeps the HTTP contract strict and timezone-safe. The
// console always submits RFC3339 timestamps; accepting local/ambiguous values
// here would make ClickHouse and MySQL return different event windows.
func parseEventTimeRange(fromValue, toValue string) (time.Time, time.Time, error) {
	parse := func(name, value string) (time.Time, error) {
		value = strings.TrimSpace(value)
		if value == "" {
			return time.Time{}, nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return time.Time{}, fmt.Errorf("%s must be RFC3339: %w", name, err)
		}
		return parsed.UTC(), nil
	}
	from, err := parse("from", fromValue)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := parse("to", toValue)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("from must not be after to")
	}
	return from, to, nil
}
