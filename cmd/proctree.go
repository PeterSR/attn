package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/proctree"
	proctreeui "github.com/petersr/attn/internal/tui/proctree"
)

// ProctreeCmd shows the process ancestor chain.
type ProctreeCmd struct {
	JSON        bool `help:"Output as JSON." default:"false"`
	Interactive bool `help:"Interactive mode: browse ancestors and assign labels." short:"i" default:"false"`
}

type proctreeEntry struct {
	PID   int    `json:"pid"`
	Name  string `json:"name"`
	Label string `json:"label,omitempty"`
}

func (p *ProctreeCmd) Run(globals *CLI) error {
	chain := proctree.AncestorsNamed(os.Getpid())
	if len(chain) == 0 {
		fmt.Fprintln(os.Stderr, "attn: could not read process tree (unsupported platform?)")
		return nil
	}

	cfg, err := config.Load(globals.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
	}

	entries := make([]proctreeEntry, len(chain))
	for i, pi := range chain {
		var label string
		if cfg.Processes != nil {
			label = cfg.Processes[pi.Name]
		}
		entries[i] = proctreeEntry{PID: pi.PID, Name: pi.Name, Label: label}
	}

	if p.Interactive {
		uiEntries := make([]proctreeui.Entry, len(entries))
		for i, e := range entries {
			uiEntries[i] = proctreeui.Entry{
				PID:       e.PID,
				Name:      e.Name,
				Label:     e.Label,
				OrigLabel: e.Label,
			}
		}
		return proctreeui.Run(uiEntries, globals.ConfigFile)
	}

	if p.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tNAME\tLABEL")
	for _, e := range entries {
		fmt.Fprintf(w, "%d\t%s\t%s\n", e.PID, e.Name, e.Label)
	}
	return w.Flush()
}
