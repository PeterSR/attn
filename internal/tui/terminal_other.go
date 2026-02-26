//go:build !linux

package tui

import "fmt"

var errUnsupported = fmt.Errorf("interactive mode is not supported on this platform")

// KeyType identifies the type of key event.
type KeyType int

const (
	KeyRune      KeyType = iota
	KeyEnter
	KeyEscape
	KeyBackspace
	KeyDelete
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyCtrlC
	KeyHome
	KeyEnd
)

// Key represents a single key press.
type Key struct {
	Type KeyType
	Rune rune
}

// EnterRawMode is not supported on this platform.
func EnterRawMode() (func(), error) { return nil, errUnsupported }

// TerminalWidth returns a default width on unsupported platforms.
func TerminalWidth() int { return 80 }

// ReadKey is not supported on this platform.
func ReadKey() (Key, error) { return Key{}, errUnsupported }

// ParseKey is a no-op on unsupported platforms.
func ParseKey([]byte) Key { return Key{} }
