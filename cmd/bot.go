package cmd

import (
	"log"

	"github.com/jon4hz/d4eventbot/config"
	"github.com/jon4hz/d4eventbot/core"
	"github.com/jon4hz/d4eventbot/d4armory"
	"github.com/jon4hz/d4eventbot/telegram"
	"github.com/spf13/cobra"
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Run the telegram bot",
	Run:   botHandler,
}

func botHandler(cmd *cobra.Command, args []string) {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	d4Client := d4armory.New()
	client := core.New(d4Client)

	tgb, err := telegram.New(cfg.Token, client)
	if err != nil {
		log.Fatalf("failed to create telegram bot: %s", err)
	}
	if err := tgb.Run(); err != nil {
		log.Fatalf("failed to run telegram bot: %s", err)
	}
}
