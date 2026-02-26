package cmd

import (
	"context"
	"errors"

	"github.com/sethrylan/slack-reader/internal/output"
	islack "github.com/sethrylan/slack-reader/internal/slack"
	"github.com/spf13/cobra"
)

var (
	messageTS    string
	messageLimit int
)

var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "Message operations",
}

var messageGetCmd = &cobra.Command{
	Use:   "get <channel>",
	Short: "Fetch a single message",
	Long: `Fetch a single message by channel and timestamp.
If the message is in a thread, includes thread metadata (reply count).

Examples:
  slack-reader message get "#general" --workspace myteam --ts "1770165109.628379"
  slack-reader message get C0123ABC --workspace myteam --ts "1770165109.628379"`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if messageTS == "" {
			output.PrintError(errors.New("--ts is required"))
		}

		domain := requireWorkspace()
		client, err := islack.NewClient(domain)
		if err != nil {
			output.PrintError(err)
		}

		ctx := context.Background()
		channelID, err := islack.ResolveChannelID(ctx, client, args[0])
		if err != nil {
			output.PrintError(err)
		}

		result, err := islack.GetMessage(ctx, client, channelID, messageTS)
		if err != nil {
			output.PrintError(err)
		}

		output.PrintJSON(result)
	},
}

var messageListCmd = &cobra.Command{
	Use:   "list <channel>",
	Short: "List all messages in a thread",
	Long: `List all messages in a thread by channel and thread root timestamp.

Examples:
  slack-reader message list "#general" --workspace myteam --ts "1770165109.628379"
  slack-reader message list C0123ABC --workspace myteam --ts "1770165109.628379"`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		domain := requireWorkspace()
		client, err := islack.NewClient(domain)
		if err != nil {
			output.PrintError(err)
		}

		ctx := context.Background()
		channelID, err := islack.ResolveChannelID(ctx, client, args[0])
		if err != nil {
			output.PrintError(err)
		}

		var messages []map[string]any
		if messageTS == "" {
			// No --ts: list recent channel messages
			messages, err = islack.ListChannelHistory(ctx, client, channelID, messageLimit)
		} else {
			// With --ts: list thread replies
			messages, err = islack.ListThread(ctx, client, channelID, messageTS, messageLimit)
		}
		if err != nil {
			output.PrintError(err)
		}

		output.PrintJSON(map[string]any{
			"messages": messages,
		})
	},
}

func init() {
	messageGetCmd.Flags().StringVar(&messageTS, "ts", "", "Message timestamp (required)")
	messageListCmd.Flags().StringVar(&messageTS, "ts", "", "Thread root timestamp (required)")
	messageListCmd.Flags().IntVar(&messageLimit, "limit", 0, "Maximum number of messages (0 = unlimited)")

	messageCmd.AddCommand(messageGetCmd)
	messageCmd.AddCommand(messageListCmd)
	rootCmd.AddCommand(messageCmd)
}
