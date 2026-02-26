package slack

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var channelIDPattern = regexp.MustCompile(`^[CDG][A-Z0-9]{8,}$`)

// NormalizeChannelInput parses a channel reference (e.g., "#general", "general", "C0123ABC")
// and returns the cleaned value and whether it's an ID.
func NormalizeChannelInput(input string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if strings.HasPrefix(trimmed, "#") {
		return trimmed[1:], false
	}
	if channelIDPattern.MatchString(trimmed) {
		return trimmed, true
	}
	return trimmed, false
}

// ResolveChannelID resolves a channel name or ID to a channel ID.
// Uses search.messages for fast resolution, falling back to conversations.list pagination.
func ResolveChannelID(ctx context.Context, client *Client, input string) (string, error) {
	name, isID := NormalizeChannelInput(input)
	if isID {
		return name, nil
	}
	if name == "" {
		return "", errors.New("channel name is empty")
	}

	// Fast path: search.messages with in:#name resolves in 1 API call
	channelID, err := resolveViaSearch(ctx, client, name)
	if err == nil && channelID != "" {
		return channelID, nil
	}

	// Slow path: paginate conversations.list
	return resolveViaPagination(ctx, client, name)
}

func resolveViaSearch(ctx context.Context, client *Client, name string) (string, error) {
	resp, err := client.API(ctx, "search.messages", map[string]string{
		"query":    "in:#" + name,
		"count":    "1",
		"sort":     "timestamp",
		"sort_dir": "desc",
	})
	if err != nil {
		return "", err
	}

	messages, _ := resp["messages"].(map[string]any)
	if messages == nil {
		return "", errors.New("no messages field")
	}

	matches, _ := messages["matches"].([]any)
	if len(matches) == 0 {
		return "", errors.New("no matches")
	}

	first, _ := matches[0].(map[string]any)
	if first == nil {
		return "", errors.New("invalid match")
	}

	channel, _ := first["channel"].(map[string]any)
	if channel == nil {
		return "", errors.New("no channel in match")
	}

	channelID, _ := channel["id"].(string)
	if channelID == "" {
		return "", errors.New("no channel ID in match")
	}

	return channelID, nil
}

func resolveViaPagination(ctx context.Context, client *Client, name string) (string, error) {
	cursor := ""
	for {
		params := map[string]string{
			"exclude_archived": "true",
			"limit":            "200",
			"types":            "public_channel,private_channel",
		}
		if cursor != "" {
			params["cursor"] = cursor
		}

		resp, err := client.API(ctx, "conversations.list", params)
		if err != nil {
			return "", fmt.Errorf("conversations.list: %w", err)
		}

		channels, _ := resp["channels"].([]any)
		for _, ch := range channels {
			c, _ := ch.(map[string]any)
			if c == nil {
				continue
			}
			if cName, _ := c["name"].(string); cName == name {
				if cID, _ := c["id"].(string); cID != "" {
					return cID, nil
				}
			}
		}

		meta, _ := resp["response_metadata"].(map[string]any)
		next, _ := meta["next_cursor"].(string)
		if next == "" {
			break
		}
		cursor = next
	}

	return "", fmt.Errorf("could not resolve channel name: #%s", name)
}

// ListUserConversations calls users.conversations to list channels for a user.
func ListUserConversations(ctx context.Context, client *Client, user string, limit int, cursor string) (map[string]any, error) {
	params := map[string]string{
		"limit":            strconv.Itoa(normalizeLimit(limit)),
		"types":            "public_channel,private_channel,im,mpim",
		"exclude_archived": "true",
	}
	if user != "" {
		// Strip leading @ if present
		params["user"] = strings.TrimPrefix(strings.TrimSpace(user), "@")
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	return client.API(ctx, "users.conversations", params)
}

// ListAllConversations calls conversations.list to list all workspace channels.
func ListAllConversations(ctx context.Context, client *Client, limit int, cursor string) (map[string]any, error) {
	params := map[string]string{
		"limit":            strconv.Itoa(normalizeLimit(limit)),
		"types":            "public_channel,private_channel,im,mpim",
		"exclude_archived": "true",
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	return client.API(ctx, "conversations.list", params)
}

// ResolveUserID resolves a @handle to a user ID by listing workspace members.
func ResolveUserID(ctx context.Context, client *Client, handle string) (string, error) {
	cleaned := strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if cleaned == "" {
		return "", errors.New("user handle is empty")
	}

	// If it already looks like a user ID, return it
	if regexp.MustCompile(`^U[A-Z0-9]{8,}$`).MatchString(cleaned) {
		return cleaned, nil
	}

	// Paginate users.list to find the user by name/display_name
	cursor := ""
	for {
		params := map[string]string{"limit": "200"}
		if cursor != "" {
			params["cursor"] = cursor
		}

		resp, err := client.API(ctx, "users.list", params)
		if err != nil {
			return "", fmt.Errorf("users.list: %w", err)
		}

		members, _ := resp["members"].([]any)
		for _, m := range members {
			member, _ := m.(map[string]any)
			if member == nil {
				continue
			}
			name, _ := member["name"].(string)
			if name == cleaned {
				if id, _ := member["id"].(string); id != "" {
					return id, nil
				}
			}
		}

		meta, _ := resp["response_metadata"].(map[string]any)
		next, _ := meta["next_cursor"].(string)
		if next == "" {
			break
		}
		cursor = next
	}

	return "", fmt.Errorf("could not resolve user: @%s", cleaned)
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}
