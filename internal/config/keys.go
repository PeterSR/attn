package config

import (
	"fmt"
	"sort"
	"strings"
)

type keyInfo struct {
	section  string
	tomlKey  string
	validate func(string) error
	get      func(Config) string
}

var keys = map[string]keyInfo{
	"format.prefix": {
		section: "format", tomlKey: "prefix",
		get: func(c Config) string { return c.Format.Prefix },
	},
	"desktop.when": {
		section: "desktop", tomlKey: "when", validate: validateWhen,
		get: func(c Config) string { return string(c.Desktop.When) },
	},
	"bell.when": {
		section: "bell", tomlKey: "when", validate: validateWhen,
		get: func(c Config) string { return string(c.Bell.When) },
	},
	"ntfy.when": {
		section: "ntfy", tomlKey: "when", validate: validateWhen,
		get: func(c Config) string { return string(c.Ntfy.When) },
	},
	"ntfy.server": {
		section: "ntfy", tomlKey: "server",
		get: func(c Config) string { return c.Ntfy.Server },
	},
	"ntfy.topic": {
		section: "ntfy", tomlKey: "topic",
		get: func(c Config) string { return c.Ntfy.Topic },
	},
	"ntfy.token": {
		section: "ntfy", tomlKey: "token",
		get: func(c Config) string { return c.Ntfy.Token },
	},
	"pushover.when": {
		section: "pushover", tomlKey: "when", validate: validateWhen,
		get: func(c Config) string { return string(c.Pushover.When) },
	},
	"pushover.token": {
		section: "pushover", tomlKey: "token",
		get: func(c Config) string { return c.Pushover.Token },
	},
	"pushover.user_key": {
		section: "pushover", tomlKey: "user_key",
		get: func(c Config) string { return c.Pushover.UserKey },
	},
	"webhook.when": {
		section: "webhook", tomlKey: "when", validate: validateWhen,
		get: func(c Config) string { return string(c.Webhook.When) },
	},
	"webhook.url": {
		section: "webhook", tomlKey: "url",
		get: func(c Config) string { return c.Webhook.URL },
	},
	"webhook.method": {
		section: "webhook", tomlKey: "method",
		get: func(c Config) string { return c.Webhook.Method },
	},
	"relay.when": {
		section: "relay", tomlKey: "when", validate: validateWhen,
		get: func(c Config) string { return string(c.Relay.When) },
	},
	"relay.socket_path": {
		section: "relay", tomlKey: "socket_path",
		get: func(c Config) string { return c.Relay.SocketPath },
	},
}

// Get loads the config and returns the effective value for the given key.
func Get(path, key string) (string, error) {
	// Dynamic key: processes.<name>
	if name, ok := strings.CutPrefix(key, "processes."); ok {
		if name == "" {
			return "", fmt.Errorf("processes key requires a name: processes.<name>")
		}
		cfg, err := Load(path)
		if err != nil {
			return "", err
		}
		return cfg.Processes[name], nil
	}

	ki, ok := keys[key]
	if !ok {
		return "", unknownKeyError(key)
	}
	cfg, err := Load(path)
	if err != nil {
		return "", err
	}
	return ki.get(cfg), nil
}

// LookupKey returns the section and TOML key name for a dotted key,
// along with an optional validation function. Returns an error for unknown keys.
func LookupKey(key string) (section, tomlKey string, validate func(string) error, err error) {
	// Dynamic key: processes.<name>
	if name, ok := strings.CutPrefix(key, "processes."); ok {
		if name == "" {
			return "", "", nil, fmt.Errorf("processes key requires a name: processes.<name>")
		}
		return "processes", name, nil, nil
	}

	ki, ok := keys[key]
	if !ok {
		return "", "", nil, unknownKeyError(key)
	}
	return ki.section, ki.tomlKey, ki.validate, nil
}

// ValidKeys returns all supported keys in sorted order.
func ValidKeys() []string {
	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func validateWhen(v string) error {
	if !When(v).Valid() || v == "" {
		return fmt.Errorf("invalid when value %q (must be never, active, idle, or always)", v)
	}
	return nil
}

func unknownKeyError(key string) error {
	return fmt.Errorf("unknown key %q\nValid keys:\n  %s\n  processes.<name>", key, strings.Join(ValidKeys(), "\n  "))
}
