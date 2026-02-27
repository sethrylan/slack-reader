package output

import (
	"fmt"
	"math"
	"strings"

	slackmd "github.com/rneatherway/slack/pkg/markdown"
)

// UserResolver resolves user IDs and message authors to display names.
type UserResolver interface {
	slackmd.UserProvider
	UsernameForMessage(msg map[string]any) (string, error)
}

// FormatMarkdown converts Slack messages to GitHub-flavored markdown,
// following the rneatherway/gh-slack blockquote style.
func FormatMarkdown(messages []map[string]any, users UserResolver) (string, error) {
	b := &strings.Builder{}

	type msgMeta struct {
		ts        string
		seconds   float64
		speakerID string
	}

	metas := make([]msgMeta, len(messages))
	for i, msg := range messages {
		ts, _ := msg["ts"].(string)
		tm, err := slackmd.ParseUnixTimestamp(ts)
		if err != nil {
			return "", fmt.Errorf("parse timestamp %q: %w", ts, err)
		}

		speakerID, _ := msg["user"].(string)
		if speakerID == "" {
			speakerID, _ = msg["bot_id"].(string)
		}

		metas[i] = msgMeta{
			ts:        ts,
			seconds:   float64(tm.Unix()),
			speakerID: speakerID,
		}
	}

	lastSpeakerID := ""

	for i, msg := range messages {
		meta := metas[i]
		tm, _ := slackmd.ParseUnixTimestamp(meta.ts)

		minutesDiff := 0
		if i > 0 {
			minutesDiff = int(math.Abs(metas[i].seconds-metas[i-1].seconds) / 60)
		}

		// How far apart in minutes can two messages be, by the same author, before we repeat the header?
		const messageTimeMinuteCutoff = 60

		speakerChanged := lastSpeakerID != "" && meta.speakerID != lastSpeakerID
		timeCutoff := minutesDiff > messageTimeMinuteCutoff

		if speakerChanged || timeCutoff {
			fmt.Fprintf(b, "\n")
		}

		includeSpeakerHeader := lastSpeakerID == "" || speakerChanged || timeCutoff

		if includeSpeakerHeader {
			username, err := users.UsernameForMessage(msg)
			if err != nil {
				return "", err
			}
			fmt.Fprintf(b, "> **%s** at %s\n",
				username,
				tm.UTC().Format("2006-01-02 15:04 MST"))
		}
		fmt.Fprintf(b, ">\n")

		text, _ := msg["text"].(string)
		if text != "" {
			converted, err := slackmd.Convert(users, text)
			if err != nil {
				return "", err
			}
			for line := range strings.SplitSeq(converted, "\n") {
				fmt.Fprintf(b, "> %s\n", line)
			}
		}

		// Include attachment text (common in bot messages)
		if attachments, _ := msg["attachments"].([]any); len(attachments) > 0 {
			for _, a := range attachments {
				att, _ := a.(map[string]any)
				if att == nil {
					continue
				}
				attText, _ := att["text"].(string)
				if attText != "" {
					converted, err := slackmd.Convert(users, attText)
					if err != nil {
						return "", err
					}
					for line := range strings.SplitSeq(converted, "\n") {
						fmt.Fprintf(b, "> %s\n", line)
					}
				}
			}
		}

		if !includeSpeakerHeader {
			b.WriteString("\n")
		}

		lastSpeakerID = meta.speakerID
	}

	return b.String(), nil
}

// PrintMarkdown formats messages as markdown and prints to stdout.
func PrintMarkdown(messages []map[string]any, users UserResolver) {
	md, err := FormatMarkdown(messages, users)
	if err != nil {
		PrintError(err)
	}
	fmt.Print(md)
}
