package proctree

import (
	"strings"
	"testing"

	"github.com/petersr/attn/internal/tui"
)

func testEntries() []Entry {
	return []Entry{
		{PID: 100, Name: "zsh", Label: "", OrigLabel: ""},
		{PID: 50, Name: "tmux", Label: "terminal-mux", OrigLabel: "terminal-mux"},
		{PID: 10, Name: "gnome-shell", Label: "", OrigLabel: ""},
		{PID: 1, Name: "systemd", Label: "", OrigLabel: ""},
	}
}

func TestNavigateDown(t *testing.T) {
	m := NewModel(testEntries(), 80)

	m.Update(tui.Key{Type: tui.KeyRune, Rune: 'j'})
	if m.Cursor != 1 {
		t.Fatalf("expected cursor=1, got %d", m.Cursor)
	}

	m.Update(tui.Key{Type: tui.KeyDown})
	if m.Cursor != 2 {
		t.Fatalf("expected cursor=2, got %d", m.Cursor)
	}
}

func TestNavigateUp(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Cursor = 2

	m.Update(tui.Key{Type: tui.KeyRune, Rune: 'k'})
	if m.Cursor != 1 {
		t.Fatalf("expected cursor=1, got %d", m.Cursor)
	}

	m.Update(tui.Key{Type: tui.KeyUp})
	if m.Cursor != 0 {
		t.Fatalf("expected cursor=0, got %d", m.Cursor)
	}
}

func TestNavigateBounds(t *testing.T) {
	m := NewModel(testEntries(), 80)

	m.Update(tui.Key{Type: tui.KeyRune, Rune: 'k'})
	if m.Cursor != 0 {
		t.Fatalf("expected cursor=0, got %d", m.Cursor)
	}

	m.Cursor = 3
	m.Update(tui.Key{Type: tui.KeyRune, Rune: 'j'})
	if m.Cursor != 3 {
		t.Fatalf("expected cursor=3, got %d", m.Cursor)
	}
}

func TestEditAndConfirm(t *testing.T) {
	m := NewModel(testEntries(), 80)

	m.Update(tui.Key{Type: tui.KeyEnter})
	if m.Mode != ModeEdit {
		t.Fatal("expected ModeEdit")
	}

	for _, r := range "shell" {
		m.Update(tui.Key{Type: tui.KeyRune, Rune: r})
	}

	m.Update(tui.Key{Type: tui.KeyEnter})
	if m.Mode != ModeNavigate {
		t.Fatal("expected ModeNavigate after confirm")
	}
	if m.Entries[0].Label != "shell" {
		t.Fatalf("expected label='shell', got %q", m.Entries[0].Label)
	}
}

func TestEditAndCancel(t *testing.T) {
	m := NewModel(testEntries(), 80)

	m.Update(tui.Key{Type: tui.KeyEnter})
	for _, r := range "test" {
		m.Update(tui.Key{Type: tui.KeyRune, Rune: r})
	}

	m.Update(tui.Key{Type: tui.KeyEscape})
	if m.Mode != ModeNavigate {
		t.Fatal("expected ModeNavigate after cancel")
	}
	if m.Entries[0].Label != "" {
		t.Fatalf("expected empty label after cancel, got %q", m.Entries[0].Label)
	}
}

func TestEditBackspace(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Cursor = 1 // tmux, label = "terminal-mux"

	m.Update(tui.Key{Type: tui.KeyEnter})
	// Cursor at end: "terminal-mux|"
	for i := 0; i < 4; i++ {
		m.Update(tui.Key{Type: tui.KeyBackspace})
	}
	m.Update(tui.Key{Type: tui.KeyEnter})

	if m.Entries[1].Label != "terminal" {
		t.Fatalf("expected 'terminal', got %q", m.Entries[1].Label)
	}
}

func TestEditCursorMovement(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Cursor = 1

	m.Update(tui.Key{Type: tui.KeyEnter})
	// EditBuf = "terminal-mux", EditCursor = 12

	// Move left 4 times to position before "-mux"
	for i := 0; i < 4; i++ {
		m.Update(tui.Key{Type: tui.KeyLeft})
	}
	if m.EditCursor != 8 {
		t.Fatalf("expected EditCursor=8, got %d", m.EditCursor)
	}

	// Delete character at cursor
	m.Update(tui.Key{Type: tui.KeyDelete})
	m.Update(tui.Key{Type: tui.KeyEnter})

	if m.Entries[1].Label != "terminalmux" {
		t.Fatalf("expected 'terminalmux', got %q", m.Entries[1].Label)
	}
}

func TestEditHome(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Cursor = 1

	m.Update(tui.Key{Type: tui.KeyEnter})
	m.Update(tui.Key{Type: tui.KeyHome})
	if m.EditCursor != 0 {
		t.Fatalf("expected EditCursor=0, got %d", m.EditCursor)
	}

	m.Update(tui.Key{Type: tui.KeyEnd})
	if m.EditCursor != 12 {
		t.Fatalf("expected EditCursor=12, got %d", m.EditCursor)
	}
	m.Update(tui.Key{Type: tui.KeyEscape})
}

func TestDeleteLabel(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Cursor = 1

	m.Update(tui.Key{Type: tui.KeyRune, Rune: 'd'})
	if m.Entries[1].Label != "" {
		t.Fatalf("expected empty label, got %q", m.Entries[1].Label)
	}
}

func TestPendingChanges(t *testing.T) {
	m := NewModel(testEntries(), 80)

	changes := m.PendingChanges()
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes, got %d", len(changes))
	}

	// Add a label to zsh.
	m.Update(tui.Key{Type: tui.KeyEnter})
	for _, r := range "Z Shell" {
		m.Update(tui.Key{Type: tui.KeyRune, Rune: r})
	}
	m.Update(tui.Key{Type: tui.KeyEnter})

	// Delete existing label on tmux.
	m.Cursor = 1
	m.Update(tui.Key{Type: tui.KeyRune, Rune: 'd'})

	changes = m.PendingChanges()
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes["zsh"] != "Z Shell" {
		t.Fatalf("expected zsh='Z Shell', got %q", changes["zsh"])
	}
	if changes["tmux"] != "" {
		t.Fatalf("expected tmux='', got %q", changes["tmux"])
	}
}

func TestPendingChangesNoOpWhenUnchanged(t *testing.T) {
	m := NewModel(testEntries(), 80)

	// Edit and confirm with same value.
	m.Cursor = 1
	m.Update(tui.Key{Type: tui.KeyEnter})
	m.Update(tui.Key{Type: tui.KeyEnter}) // Confirm with original value.

	changes := m.PendingChanges()
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes, got %d", len(changes))
	}
}

func TestQuitSave(t *testing.T) {
	m := NewModel(testEntries(), 80)
	action := m.Update(tui.Key{Type: tui.KeyRune, Rune: 'q'})
	if action != ActionSave {
		t.Fatalf("expected ActionSave, got %d", action)
	}
}

func TestEscapeSaves(t *testing.T) {
	m := NewModel(testEntries(), 80)
	action := m.Update(tui.Key{Type: tui.KeyEscape})
	if action != ActionSave {
		t.Fatalf("expected ActionSave, got %d", action)
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := NewModel(testEntries(), 80)
	action := m.Update(tui.Key{Type: tui.KeyCtrlC})
	if action != ActionQuit {
		t.Fatalf("expected ActionQuit, got %d", action)
	}
}

func TestCtrlCInEditCancelsAndQuits(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Update(tui.Key{Type: tui.KeyEnter})
	for _, r := range "test" {
		m.Update(tui.Key{Type: tui.KeyRune, Rune: r})
	}

	action := m.Update(tui.Key{Type: tui.KeyCtrlC})
	if action != ActionQuit {
		t.Fatalf("expected ActionQuit, got %d", action)
	}
	if m.Mode != ModeNavigate {
		t.Fatal("expected ModeNavigate after Ctrl+C in edit")
	}
	if m.Entries[0].Label != "" {
		t.Fatalf("expected empty label (cancelled), got %q", m.Entries[0].Label)
	}
}

func TestViewContainsEntries(t *testing.T) {
	m := NewModel(testEntries(), 80)
	view := m.View()

	for _, want := range []string{"zsh", "tmux", "systemd", "4 processes", "1 labeled"} {
		if !strings.Contains(view, want) {
			t.Errorf("view should contain %q", want)
		}
	}
}

func TestViewEditMode(t *testing.T) {
	m := NewModel(testEntries(), 80)
	m.Update(tui.Key{Type: tui.KeyEnter})
	for _, r := range "hi" {
		m.Update(tui.Key{Type: tui.KeyRune, Rune: r})
	}

	view := m.View()
	if !strings.Contains(view, "Enter confirm") {
		t.Error("edit mode should show confirm/cancel help")
	}
	if !strings.Contains(view, "[hi\u258c]") {
		t.Error("edit mode should show inline editor")
	}
}
