package proctree

import (
	"fmt"
	"strings"

	"github.com/petersr/attn/internal/tui"
)

// Mode represents the current interaction mode.
type Mode int

const (
	ModeNavigate Mode = iota
	ModeEdit
)

// Action represents the result of an update.
type Action int

const (
	ActionNone Action = iota
	ActionSave                // Save pending changes and quit.
	ActionQuit                // Quit without saving.
)

// Entry represents a single process in the ancestor chain.
type Entry struct {
	PID       int
	Name      string
	Label     string
	OrigLabel string
}

// Model is the state of the interactive proctree UI.
type Model struct {
	Entries    []Entry
	Cursor     int
	Mode       Mode
	EditBuf    []rune
	EditCursor int
	Width      int
}

// NewModel creates a new Model from process entries.
func NewModel(entries []Entry, width int) *Model {
	if width <= 0 {
		width = 80
	}
	return &Model{
		Entries: entries,
		Width:   width,
	}
}

// Update processes a key event and returns the resulting action.
func (m *Model) Update(key tui.Key) Action {
	if m.Mode == ModeEdit {
		return m.updateEdit(key)
	}
	return m.updateNavigate(key)
}

func (m *Model) updateNavigate(key tui.Key) Action {
	switch key.Type {
	case tui.KeyRune:
		switch key.Rune {
		case 'q':
			return ActionSave
		case 'j':
			m.moveCursor(1)
		case 'k':
			m.moveCursor(-1)
		case 'd':
			m.deleteLabel()
		}
	case tui.KeyUp:
		m.moveCursor(-1)
	case tui.KeyDown:
		m.moveCursor(1)
	case tui.KeyEnter:
		m.enterEdit()
	case tui.KeyEscape:
		return ActionSave
	case tui.KeyCtrlC:
		return ActionQuit
	}
	return ActionNone
}

func (m *Model) updateEdit(key tui.Key) Action {
	switch key.Type {
	case tui.KeyEnter:
		m.confirmEdit()
	case tui.KeyEscape:
		m.cancelEdit()
	case tui.KeyCtrlC:
		m.cancelEdit()
		return ActionQuit
	case tui.KeyBackspace:
		if m.EditCursor > 0 {
			m.EditBuf = append(m.EditBuf[:m.EditCursor-1], m.EditBuf[m.EditCursor:]...)
			m.EditCursor--
		}
	case tui.KeyDelete:
		if m.EditCursor < len(m.EditBuf) {
			m.EditBuf = append(m.EditBuf[:m.EditCursor], m.EditBuf[m.EditCursor+1:]...)
		}
	case tui.KeyLeft:
		if m.EditCursor > 0 {
			m.EditCursor--
		}
	case tui.KeyRight:
		if m.EditCursor < len(m.EditBuf) {
			m.EditCursor++
		}
	case tui.KeyHome:
		m.EditCursor = 0
	case tui.KeyEnd:
		m.EditCursor = len(m.EditBuf)
	case tui.KeyRune:
		if key.Rune >= 32 {
			buf := make([]rune, len(m.EditBuf)+1)
			copy(buf, m.EditBuf[:m.EditCursor])
			buf[m.EditCursor] = key.Rune
			copy(buf[m.EditCursor+1:], m.EditBuf[m.EditCursor:])
			m.EditBuf = buf
			m.EditCursor++
		}
	}
	return ActionNone
}

func (m *Model) moveCursor(delta int) {
	m.Cursor += delta
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Entries) {
		m.Cursor = len(m.Entries) - 1
	}
}

func (m *Model) deleteLabel() {
	if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		m.Entries[m.Cursor].Label = ""
	}
}

func (m *Model) enterEdit() {
	if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		m.Mode = ModeEdit
		m.EditBuf = []rune(m.Entries[m.Cursor].Label)
		m.EditCursor = len(m.EditBuf)
	}
}

func (m *Model) confirmEdit() {
	if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		m.Entries[m.Cursor].Label = string(m.EditBuf)
	}
	m.Mode = ModeNavigate
	m.EditBuf = nil
	m.EditCursor = 0
}

func (m *Model) cancelEdit() {
	m.Mode = ModeNavigate
	m.EditBuf = nil
	m.EditCursor = 0
}

// PendingChanges returns entries where the label differs from the original.
// Keys are process names, values are the new labels.
func (m *Model) PendingChanges() map[string]string {
	changes := make(map[string]string)
	for _, e := range m.Entries {
		if e.Label != e.OrigLabel {
			changes[e.Name] = e.Label
		}
	}
	return changes
}

// View renders the current state as a string for display.
func (m *Model) View() string {
	var b strings.Builder

	b.WriteString(tui.Bold("  attn: process ancestors"))
	b.WriteString("\r\n")

	lineWidth := m.Width
	if lineWidth > 60 {
		lineWidth = 60
	}
	separator := "  " + strings.Repeat("\u2500", lineWidth-2)

	b.WriteString(tui.Dim(separator))
	b.WriteString("\r\n")
	if m.Mode == ModeEdit {
		b.WriteString(tui.Dim("  Enter confirm  Esc cancel"))
	} else {
		b.WriteString(tui.Dim("  j/k navigate  Enter edit  d delete  q save & quit"))
	}
	b.WriteString("\r\n")
	b.WriteString(tui.Dim(separator))
	b.WriteString("\r\n\r\n")

	// Column headers.
	b.WriteString(fmt.Sprintf("    %-10s%-18s%s\r\n", "PID", "NAME", "LABEL"))

	for i, e := range m.Entries {
		prefix := "    "
		if i == m.Cursor {
			prefix = "  \u276f "
		}

		label := e.Label
		if m.Mode == ModeEdit && i == m.Cursor {
			before := string(m.EditBuf[:m.EditCursor])
			after := string(m.EditBuf[m.EditCursor:])
			label = "[" + before + "\u258c" + after + "]"
		} else if label == "" {
			label = tui.Dim("\u00b7")
		} else {
			label = tui.Cyan(label)
		}

		b.WriteString(fmt.Sprintf("%s%-10d%-18s%s\r\n", prefix, e.PID, e.Name, label))
	}

	b.WriteString("\r\n")
	b.WriteString(tui.Dim(separator))
	b.WriteString("\r\n")

	labeled := 0
	for _, e := range m.Entries {
		if e.Label != "" {
			labeled++
		}
	}
	b.WriteString(tui.Dim(fmt.Sprintf("  %d processes, %d labeled", len(m.Entries), labeled)))
	b.WriteString("\r\n")

	return b.String()
}
