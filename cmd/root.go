package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var workspace string

var rootCmd = &cobra.Command{
	Use:   "slack-reader",
	Short: "Read-only Slack CLI using cookie-based authentication",
	Long:  "A CLI tool for reading Slack messages, threads, and channel lists using cookie-based authentication from Slack Desktop.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&workspace, "workspace", "", "Slack team domain (e.g., \"myteam\" for myteam.slack.com)")
}
