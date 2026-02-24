package cmd

import (
	"fmt"

	"github.com/petersr/attn/internal/config"
)

// ConfigCmd groups the config subcommands.
type ConfigCmd struct {
	Set  ConfigSetCmd  `cmd:"" help:"Set a configuration value."`
	Get  ConfigGetCmd  `cmd:"" help:"Get a configuration value."`
	Path ConfigPathCmd `cmd:"" help:"Print the configuration file path."`
}

// ConfigSetCmd sets a config value.
type ConfigSetCmd struct {
	Key   string `arg:"" help:"Config key (e.g. ntfy.topic)."`
	Value string `arg:"" help:"Value to set."`
}

func (c *ConfigSetCmd) Run(globals *CLI) error {
	path := configPath(globals)
	if err := config.Set(path, c.Key, c.Value); err != nil {
		return err
	}
	fmt.Printf("%s = %q\n", c.Key, c.Value)
	return nil
}

// ConfigGetCmd gets a config value.
type ConfigGetCmd struct {
	Key string `arg:"" help:"Config key (e.g. ntfy.topic)."`
}

func (c *ConfigGetCmd) Run(globals *CLI) error {
	path := configPath(globals)
	val, err := config.Get(path, c.Key)
	if err != nil {
		return err
	}
	fmt.Println(val)
	return nil
}

// ConfigPathCmd prints the config file path.
type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run(globals *CLI) error {
	fmt.Println(configPath(globals))
	return nil
}

func configPath(globals *CLI) string {
	if globals.ConfigFile != "" {
		return globals.ConfigFile
	}
	return config.DefaultPath()
}
