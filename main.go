package main

import (
	"log"

	"github.com/jon4hz/d4eventbot/config"
	"github.com/jon4hz/d4eventbot/d4armory"
	"github.com/jon4hz/d4eventbot/telegram"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	client := d4armory.New()

	tgb, err := telegram.New(cfg.Token, client)
	if err != nil {
		log.Fatalf("failed to create telegram bot: %s", err)
	}
	if err := tgb.Run(); err != nil {
		log.Fatalf("failed to run telegram bot: %s", err)
	}
}
