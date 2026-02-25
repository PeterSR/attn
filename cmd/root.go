package cmd

var version = "dev"

// CLI is the root command definition for Kong.
type CLI struct {
	Send     SendCmd     `cmd:"" help:"Send a notification."`
	Serve    ServeCmd    `cmd:"" help:"Start the relay server for remote notifications."`
	Config   ConfigCmd   `cmd:"" help:"Get and set configuration values."`
	Proctree ProctreeCmd `cmd:"" help:"Show the process ancestor chain."`
	Version  VersionCmd  `cmd:"" help:"Print version information."`

	// Global flags.
	ConfigFile string `name:"config" short:"C" help:"Config file path." default:"" env:"ATTN_CONFIG_PATH"`
}
