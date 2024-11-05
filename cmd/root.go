package cmd

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jon4hz/d4eventbot/core"
	"github.com/jon4hz/d4eventbot/d4armory"
	"github.com/jon4hz/d4eventbot/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "d4eventbot",
	Short: "Tracking Diablo 4 events",
	Run:   rootHandler,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(botCmd)
}

func rootHandler(cmd *cobra.Command, args []string) {
	d4Client := d4armory.New()
	client := core.New(d4Client)

	model := tui.New(client)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		log.Fatalf("failed to run tui: %s", err)
	}
}
