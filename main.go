package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/petersr/attn/cmd"
)

// Set via ldflags.
var version = "dev"

func main() {
	cmd.SetVersion(version)

	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("attn"),
		kong.Description("Send notifications when processes need your attention."),
		kong.UsageOnError(),
	)
	if err := ctx.Run(&cli); err != nil {
		fmt.Fprintf(os.Stderr, "attn: %v\n", err)
		os.Exit(1)
	}
}
