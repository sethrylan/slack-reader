package slack

import (
	"context"
	"sync"
)

// UserProvider resolves Slack user IDs to display names.
// It implements the rneatherway/slack/pkg/markdown.UserProvider interface.
type UserProvider struct {
	client *Client
	mu     sync.Mutex
	cache  map[string]string
}

// NewUserProvider creates a UserProvider backed by the Slack users.info API.
func NewUserProvider(client *Client) *UserProvider {
	return &UserProvider{
		client: client,
		cache:  make(map[string]string),
	}
}

// UsernameForID resolves a Slack user ID to a display name.
func (u *UserProvider) UsernameForID(id string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if name, ok := u.cache[id]; ok {
		return name, nil
	}

	resp, err := u.client.API(context.Background(), "users.info", map[string]string{"user": id})
	if err != nil {
		// Fall back to raw ID on error
		u.cache[id] = id
		return id, nil
	}

	user, _ := resp["user"].(map[string]any)
	if user == nil {
		u.cache[id] = id
		return id, nil
	}

	name := resolveDisplayName(user)
	u.cache[id] = name
	return name, nil
}

// resolveDisplayName picks the best display name from a user object.
func resolveDisplayName(user map[string]any) string {
	profile, _ := user["profile"].(map[string]any)
	if profile != nil {
		if dn, _ := profile["display_name"].(string); dn != "" {
			return dn
		}
		if rn, _ := profile["real_name"].(string); rn != "" {
			return rn
		}
	}
	if rn, _ := user["real_name"].(string); rn != "" {
		return rn
	}
	if name, _ := user["name"].(string); name != "" {
		return name
	}
	if id, _ := user["id"].(string); id != "" {
		return id
	}
	return "unknown"
}

// UsernameForMessage returns the display name for a message's author.
func (u *UserProvider) UsernameForMessage(msg map[string]any) (string, error) {
	if userID, _ := msg["user"].(string); userID != "" {
		return u.UsernameForID(userID)
	}
	if botID, _ := msg["bot_id"].(string); botID != "" {
		// Try bot profile name first
		if profile, _ := msg["bot_profile"].(map[string]any); profile != nil {
			if name, _ := profile["name"].(string); name != "" {
				return name, nil
			}
		}
		return "bot " + botID, nil
	}
	if username, _ := msg["username"].(string); username != "" {
		return username, nil
	}
	return "unknown", nil
}
