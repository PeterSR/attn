package cmd

var version = "dev"

// CLI is the root command definition for Kong.
type CLI struct {
	Send    SendCmd    `cmd:"" help:"Send a notification."`
	Serve   ServeCmd   `cmd:"" help:"Start the relay server for remote notifications."`
	Version VersionCmd `cmd:"" help:"Print version information."`

	// Global flags.
	Config string `short:"C" help:"Config file path." default:"" env:"ATTN_CONFIG"`
}
