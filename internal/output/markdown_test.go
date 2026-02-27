package output_test

import (
	"strings"
	"testing"

	"github.com/sethrylan/slack-reader/internal/output"
)

type testUserResolver struct {
	users map[string]string
}

func (t *testUserResolver) UsernameForID(id string) (string, error) {
	if name, ok := t.users[id]; ok {
		return name, nil
	}
	return id, nil
}

func (t *testUserResolver) UsernameForMessage(msg map[string]any) (string, error) {
	if userID, _ := msg["user"].(string); userID != "" {
		return t.UsernameForID(userID)
	}
	if botID, _ := msg["bot_id"].(string); botID != "" {
		return "bot " + botID, nil
	}
	return "unknown", nil
}

func TestFormatMarkdown_SingleMessage(t *testing.T) {
	users := &testUserResolver{users: map[string]string{"U123": "alice"}}
	messages := []map[string]any{
		{"user": "U123", "text": "hello world", "ts": "1679058753.0"},
	}

	result, err := output.FormatMarkdown(messages, users)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "> **alice** at 2023-03-17") {
		t.Errorf("expected header with username, got:\n%s", result)
	}
	if !strings.Contains(result, "> hello world") {
		t.Errorf("expected message text, got:\n%s", result)
	}
}

func TestFormatMarkdown_AdjacentSameUser(t *testing.T) {
	users := &testUserResolver{users: map[string]string{"U123": "alice"}}
	messages := []map[string]any{
		{"user": "U123", "text": "first", "ts": "123.456"},
		{"user": "U123", "text": "second", "ts": "124.567"},
	}

	result, err := output.FormatMarkdown(messages, users)
	if err != nil {
		t.Fatal(err)
	}

	// Should only have one header
	count := strings.Count(result, "**alice**")
	if count != 1 {
		t.Errorf("expected 1 header, got %d in:\n%s", count, result)
	}
	if !strings.Contains(result, "> first") {
		t.Errorf("expected first message, got:\n%s", result)
	}
	if !strings.Contains(result, "> second") {
		t.Errorf("expected second message, got:\n%s", result)
	}
}

func TestFormatMarkdown_DifferentUsers(t *testing.T) {
	users := &testUserResolver{users: map[string]string{"U1": "alice", "U2": "bob"}}
	messages := []map[string]any{
		{"user": "U1", "text": "hello", "ts": "123.456"},
		{"user": "U2", "text": "hi there", "ts": "124.567"},
	}

	result, err := output.FormatMarkdown(messages, users)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "**alice**") {
		t.Errorf("expected alice header, got:\n%s", result)
	}
	if !strings.Contains(result, "**bob**") {
		t.Errorf("expected bob header, got:\n%s", result)
	}
}

func TestFormatMarkdown_SameUserFarApart(t *testing.T) {
	users := &testUserResolver{users: map[string]string{"U1": "alice"}}
	messages := []map[string]any{
		{"user": "U1", "text": "hello", "ts": "1679058753.0"},
		{"user": "U1", "text": "much later", "ts": "1679064168.0"}, // ~90 min later
	}

	result, err := output.FormatMarkdown(messages, users)
	if err != nil {
		t.Fatal(err)
	}

	// Should have two headers since messages are >60min apart
	count := strings.Count(result, "**alice**")
	if count != 2 {
		t.Errorf("expected 2 headers for far-apart messages, got %d in:\n%s", count, result)
	}
}

func TestFormatMarkdown_UserMentions(t *testing.T) {
	users := &testUserResolver{users: map[string]string{"U1": "alice", "U2": "bob"}}
	messages := []map[string]any{
		{"user": "U1", "text": "hey <@U2> check this", "ts": "123.456"},
	}

	result, err := output.FormatMarkdown(messages, users)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "`@bob`") {
		t.Errorf("expected interpolated mention, got:\n%s", result)
	}
}

func TestFormatMarkdown_Links(t *testing.T) {
	users := &testUserResolver{users: map[string]string{"U1": "alice"}}
	messages := []map[string]any{
		{"user": "U1", "text": "see <https://example.com|this link>", "ts": "123.456"},
	}

	result, err := output.FormatMarkdown(messages, users)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "[this link](https://example.com)") {
		t.Errorf("expected markdown link, got:\n%s", result)
	}
}
