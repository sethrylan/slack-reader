package slack_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/sethrylan/slack-reader/internal/slack"
)

// mockAPI records calls and returns canned responses keyed by cursor.
type mockAPI struct {
	calls []map[string]string
	pages []map[string]any // responses returned in order
	idx   int
}

func (m *mockAPI) API(_ context.Context, _ string, params map[string]string) (map[string]any, error) {
	m.calls = append(m.calls, params)
	if m.idx >= len(m.pages) {
		return nil, fmt.Errorf("unexpected call #%d", m.idx)
	}
	resp := m.pages[m.idx]
	m.idx++
	return resp, nil
}

// makePage builds a fake conversations.history response with n messages and an optional next cursor.
func makePage(n int, nextCursor string) map[string]any {
	msgs := make([]any, n)
	for i := range msgs {
		msgs[i] = map[string]any{"ts": fmt.Sprintf("1770000000.%06d", i)}
	}
	resp := map[string]any{
		"ok":       true,
		"messages": msgs,
	}
	if nextCursor != "" {
		resp["has_more"] = true
		resp["response_metadata"] = map[string]any{
			"next_cursor": nextCursor,
		}
	}
	return resp
}

// Bug 1: unlimited (limit=0) must paginate through all pages,
// sending limit=200 (not "0") on each request.
func TestListChannelHistory_UnlimitedPaginates(t *testing.T) {
	mock := &mockAPI{
		pages: []map[string]any{
			makePage(200, "cursor_page2"),
			makePage(200, "cursor_page3"),
			makePage(150, ""),
		},
	}

	msgs, err := slack.ListChannelHistory(t.Context(), mock, "C123", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(msgs); got != 550 {
		t.Errorf("got %d messages, want 550", got)
	}

	// Every page request must send limit=200, never "0".
	for i, call := range mock.calls {
		if lim := call["limit"]; lim != "200" {
			t.Errorf("page %d: sent limit=%q, want \"200\"", i, lim)
		}
	}

	// Must have made 3 API calls (followed both cursors).
	if got := len(mock.calls); got != 3 {
		t.Errorf("made %d API calls, want 3", got)
	}

	// Second call must include the cursor from the first response.
	if c := mock.calls[1]["cursor"]; c != "cursor_page2" {
		t.Errorf("page 1 cursor=%q, want \"cursor_page2\"", c)
	}
}

// Bug 2: explicit limit > 200 must paginate across multiple pages.
func TestListChannelHistory_ExplicitLimitPaginates(t *testing.T) {
	mock := &mockAPI{
		pages: []map[string]any{
			makePage(200, "cursor_page2"),
			makePage(200, "cursor_page3"),
			makePage(200, ""),
		},
	}

	msgs, err := slack.ListChannelHistory(t.Context(), mock, "C123", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(msgs); got != 500 {
		t.Errorf("got %d messages, want 500", got)
	}

	// Must have made 3 API calls.
	if got := len(mock.calls); got != 3 {
		t.Errorf("made %d API calls, want 3", got)
	}

	// Third page should request only the remaining 100.
	if lim := mock.calls[2]["limit"]; lim != "100" {
		t.Errorf("page 2 limit=%q, want \"100\"", lim)
	}
}

// Sanity: single page within limit should not over-fetch.
func TestListChannelHistory_SinglePage(t *testing.T) {
	mock := &mockAPI{
		pages: []map[string]any{
			makePage(50, ""),
		},
	}

	msgs, err := slack.ListChannelHistory(t.Context(), mock, "C123", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(msgs); got != 50 {
		t.Errorf("got %d messages, want 50", got)
	}

	// Should request exactly 100.
	if lim := mock.calls[0]["limit"]; lim != "100" {
		t.Errorf("sent limit=%q, want \"100\"", lim)
	}

	// Only one API call.
	if got := len(mock.calls); got != 1 {
		t.Errorf("made %d API calls, want 1", got)
	}
}

// Exact page boundary: limit=200 should fetch one full page and stop.
func TestListChannelHistory_ExactPageBoundary(t *testing.T) {
	mock := &mockAPI{
		pages: []map[string]any{
			makePage(200, "cursor_page2"),
		},
	}

	msgs, err := slack.ListChannelHistory(t.Context(), mock, "C123", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(msgs); got != 200 {
		t.Errorf("got %d messages, want 200", got)
	}

	// Must NOT fetch a second page.
	if got := len(mock.calls); got != 1 {
		t.Errorf("made %d API calls, want 1", got)
	}
}

// Small explicit limit requests only the needed amount.
func TestListChannelHistory_SmallLimit(t *testing.T) {
	mock := &mockAPI{
		pages: []map[string]any{
			makePage(25, ""),
		},
	}

	msgs, err := slack.ListChannelHistory(t.Context(), mock, "C123", 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(msgs); got != 25 {
		t.Errorf("got %d messages, want 25", got)
	}

	lim, _ := strconv.Atoi(mock.calls[0]["limit"])
	if lim != 25 {
		t.Errorf("sent limit=%d, want 25", lim)
	}
}
