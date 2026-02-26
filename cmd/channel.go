package cmd

import (
	"context"

	"github.com/sethrylan/slack-reader/internal/output"
	islack "github.com/sethrylan/slack-reader/internal/slack"
	"github.com/spf13/cobra"
)

var (
	channelUser  string
	channelAll   bool
	channelLimit int
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Channel operations",
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List conversations",
	Long: `List conversations for the current user (default), a specific user, or all workspace conversations.

Examples:
  slack-reader channel list --workspace myteam
  slack-reader channel list --workspace myteam --user "@alice" --limit 50
  slack-reader channel list --workspace myteam --all --limit 100`,
	Run: func(_ *cobra.Command, _ []string) {
		domain := requireWorkspace()
		client, err := islack.NewClient(domain)
		if err != nil {
			output.PrintError(err)
		}

		ctx := context.Background()
		var resp map[string]any

		switch {
		case channelAll:
			resp, err = islack.ListAllConversations(ctx, client, channelLimit, "")
		case channelUser != "":
			// Resolve @handle to user ID
			userID, resolveErr := islack.ResolveUserID(ctx, client, channelUser)
			if resolveErr != nil {
				output.PrintError(resolveErr)
			}
			resp, err = islack.ListUserConversations(ctx, client, userID, channelLimit, "")
		default:
			resp, err = islack.ListUserConversations(ctx, client, "", channelLimit, "")
		}

		if err != nil {
			output.PrintError(err)
		}

		output.PrintJSON(resp)
	},
}

func init() {
	channelListCmd.Flags().StringVar(&channelUser, "user", "", "List conversations for a specific user (e.g., \"@alice\")")
	channelListCmd.Flags().BoolVar(&channelAll, "all", false, "List all workspace conversations (conversations.list)")
	channelListCmd.Flags().IntVar(&channelLimit, "limit", 100, "Maximum number of results")
	channelListCmd.MarkFlagsMutuallyExclusive("user", "all")

	channelCmd.AddCommand(channelListCmd)
	rootCmd.AddCommand(channelCmd)
}
