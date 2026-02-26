package tui

import "fmt"

const esc = "\x1b"

// ClearScreen clears the screen and moves the cursor to the top-left.
func ClearScreen() string { return esc + "[2J" + esc + "[H" }

// HideCursor hides the terminal cursor.
func HideCursor() string { return esc + "[?25l" }

// ShowCursor shows the terminal cursor.
func ShowCursor() string { return esc + "[?25h" }

// MoveTo moves the cursor to the given 1-based row and column.
func MoveTo(row, col int) string { return fmt.Sprintf(esc+"[%d;%dH", row, col) }

// Bold wraps s in bold ANSI escapes.
func Bold(s string) string { return esc + "[1m" + s + esc + "[0m" }

// Dim wraps s in dim ANSI escapes.
func Dim(s string) string { return esc + "[2m" + s + esc + "[0m" }

// Cyan wraps s in cyan ANSI escapes.
func Cyan(s string) string { return esc + "[36m" + s + esc + "[0m" }
