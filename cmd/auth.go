// Package cmd implements the CLI commands for slack-reader.
package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/sethrylan/slack-reader/internal/output"
	islack "github.com/sethrylan/slack-reader/internal/slack"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Slack authentication",
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authentication info (calls auth.test)",
	Run: func(_ *cobra.Command, _ []string) {
		domain := requireWorkspace()
		client, err := islack.NewClient(domain)
		if err != nil {
			output.PrintError(err)
		}

		resp, err := client.API(context.Background(), "auth.test", nil)
		if err != nil {
			output.PrintError(err)
		}

		output.PrintJSON(resp)
	},
}

var authCredsCmd = &cobra.Command{
	Use:   "creds",
	Short: "Import credentials from Slack Desktop (cookie-based auth)",
	Run: func(_ *cobra.Command, _ []string) {
		domain := requireWorkspace()
		client := islack.NewClientNoCreds(domain)
		if err := client.ImportCreds(); err != nil {
			output.PrintError(err)
		}

		// Verify by calling auth.test
		resp, err := client.API(context.Background(), "auth.test", nil)
		if err != nil {
			output.PrintError(err)
		}

		fmt.Println("Credentials imported successfully.")
		output.PrintJSON(resp)
	},
}

func requireWorkspace() string {
	if workspace == "" {
		output.PrintError(errors.New("--workspace is required (e.g., --workspace myteam)"))
	}
	return workspace
}

func init() {
	authCmd.AddCommand(authWhoamiCmd)
	authCmd.AddCommand(authCredsCmd)
	rootCmd.AddCommand(authCmd)
}
