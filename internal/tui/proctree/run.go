package proctree

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/tui"
)

// Run starts the interactive process tree UI.
// The raw-mode/render/read/update loop below is duplicated per view. If a third
// TUI view is added, extract it into tui.RunLoop with a common Model interface
// (Update(Key) Action + View() string).
func Run(entries []Entry, configPath string) error {
	restore, err := tui.EnterRawMode()
	if err != nil {
		return fmt.Errorf("entering raw mode: %w", err)
	}
	defer restore()

	// Restore terminal on external signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	go func() {
		<-sigCh
		fmt.Fprint(os.Stdout, tui.ShowCursor())
		restore()
		os.Exit(1)
	}()

	model := NewModel(entries, tui.TerminalWidth())

	fmt.Fprint(os.Stdout, tui.HideCursor())
	defer fmt.Fprint(os.Stdout, tui.ShowCursor())

	for {
		fmt.Fprint(os.Stdout, tui.ClearScreen())
		fmt.Fprint(os.Stdout, model.View())

		key, err := tui.ReadKey()
		if err != nil {
			return fmt.Errorf("reading key: %w", err)
		}

		switch model.Update(key) {
		case ActionSave:
			fmt.Fprint(os.Stdout, tui.ClearScreen())
			return saveChanges(model, configPath)
		case ActionQuit:
			fmt.Fprint(os.Stdout, tui.ClearScreen())
			return nil
		}
	}
}

func saveChanges(m *Model, configPath string) error {
	changes := m.PendingChanges()
	if len(changes) == 0 {
		return nil
	}

	if configPath == "" {
		configPath = config.DefaultPath()
	}

	for name, label := range changes {
		if err := config.Set(configPath, "processes."+name, label); err != nil {
			return fmt.Errorf("saving processes.%s: %w", name, err)
		}
	}

	fmt.Fprintf(os.Stderr, "Saved %d change(s) to %s\n", len(changes), configPath)
	return nil
}
