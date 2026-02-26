package slack

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// MessageResult represents a single message fetch result.
type MessageResult struct {
	Message map[string]any `json:"message"`
	Thread  map[string]any `json:"thread,omitempty"`
}

// GetMessage fetches a single message by channel and timestamp.
func GetMessage(ctx context.Context, client *Client, channelID string, ts string) (*MessageResult, error) {
	resp, err := client.API(ctx, "conversations.history", map[string]string{
		"channel":   channelID,
		"latest":    ts,
		"oldest":    ts,
		"inclusive": "true",
		"limit":     "1",
	})
	if err != nil {
		return nil, fmt.Errorf("conversations.history: %w", err)
	}

	messages, _ := resp["messages"].([]any)
	if len(messages) == 0 {
		return nil, fmt.Errorf("message not found at ts=%s", ts)
	}

	msg, _ := messages[0].(map[string]any)
	if msg == nil {
		return nil, errors.New("invalid message format")
	}

	result := &MessageResult{Message: msg}

	// If message has replies, include thread metadata
	replyCount, _ := msg["reply_count"].(float64)
	if replyCount > 0 {
		threadTS, _ := msg["thread_ts"].(string)
		if threadTS == "" {
			threadTS = ts
		}
		result.Thread = map[string]any{
			"ts":     threadTS,
			"length": int(replyCount),
		}
	}

	return result, nil
}

// ListChannelHistory fetches recent messages from a channel.
func ListChannelHistory(ctx context.Context, client *Client, channelID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}

	resp, err := client.API(ctx, "conversations.history", map[string]string{
		"channel": channelID,
		"limit":   fmt.Sprintf("%d", limit),
	})
	if err != nil {
		return nil, fmt.Errorf("conversations.history: %w", err)
	}

	messages, _ := resp["messages"].([]any)
	var out []map[string]any
	for _, m := range messages {
		msg, _ := m.(map[string]any)
		if msg != nil {
			out = append(out, msg)
		}
	}

	// Sort chronologically (oldest first)
	sort.Slice(out, func(i, j int) bool {
		tsI, _ := out[i]["ts"].(string)
		tsJ, _ := out[j]["ts"].(string)
		return tsI < tsJ
	})

	return out, nil
}

// ListThread fetches all replies in a thread, paginated.
func ListThread(ctx context.Context, client *Client, channelID string, threadTS string, limit int) ([]map[string]any, error) {
	var allMessages []map[string]any
	cursor := ""

	for {
		params := map[string]string{
			"channel": channelID,
			"ts":      threadTS,
			"limit":   "200",
		}
		if cursor != "" {
			params["cursor"] = cursor
		}

		resp, err := client.API(ctx, "conversations.replies", params)
		if err != nil {
			return nil, fmt.Errorf("conversations.replies: %w", err)
		}

		messages, _ := resp["messages"].([]any)
		for _, m := range messages {
			msg, _ := m.(map[string]any)
			if msg != nil {
				allMessages = append(allMessages, msg)
			}
		}

		meta, _ := resp["response_metadata"].(map[string]any)
		next, _ := meta["next_cursor"].(string)
		if next == "" {
			break
		}
		cursor = next

		if limit > 0 && len(allMessages) >= limit {
			allMessages = allMessages[:limit]
			break
		}
	}

	// Sort chronologically
	sort.Slice(allMessages, func(i, j int) bool {
		tsI, _ := allMessages[i]["ts"].(string)
		tsJ, _ := allMessages[j]["ts"].(string)
		return tsI < tsJ
	})

	return allMessages, nil
}
