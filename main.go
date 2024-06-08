package main

import (
	"os"

	"github.com/jon4hz/d4eventbot/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(0)
	}
}
