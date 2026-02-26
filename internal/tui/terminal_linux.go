//go:build linux

package tui

import (
	"os"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// KeyType identifies the type of key event.
type KeyType int

const (
	KeyRune      KeyType = iota // Printable character; check Key.Rune.
	KeyEnter                    // Enter / Return.
	KeyEscape                   // Escape (standalone, not part of a sequence).
	KeyBackspace                // Backspace / Ctrl+H.
	KeyDelete                   // Delete (escape sequence).
	KeyUp                       // Arrow up.
	KeyDown                     // Arrow down.
	KeyLeft                     // Arrow left.
	KeyRight                    // Arrow right.
	KeyCtrlC                    // Ctrl+C.
	KeyHome                     // Home.
	KeyEnd                      // End.
)

// Key represents a single key press.
type Key struct {
	Type KeyType
	Rune rune
}

// EnterRawMode puts stdin into raw mode and returns a restore function.
func EnterRawMode() (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	orig, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	raw := *orig
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Cflag |= unix.CS8
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &raw); err != nil {
		return nil, err
	}

	return func() {
		_ = unix.IoctlSetTermios(fd, unix.TCSETS, orig)
	}, nil
}

// TerminalWidth returns the width of the terminal, defaulting to 80.
func TerminalWidth() int {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		return 80
	}
	return int(ws.Col)
}

// ReadKey reads a single key press from stdin.
func ReadKey() (Key, error) {
	buf := make([]byte, 32)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return Key{}, err
	}
	return ParseKey(buf[:n]), nil
}

// ParseKey interprets a byte sequence as a Key.
func ParseKey(b []byte) Key {
	if len(b) == 0 {
		return Key{}
	}

	if len(b) == 1 {
		switch b[0] {
		case 3:
			return Key{Type: KeyCtrlC}
		case 8:
			return Key{Type: KeyBackspace}
		case 13:
			return Key{Type: KeyEnter}
		case 27:
			return Key{Type: KeyEscape}
		case 127:
			return Key{Type: KeyBackspace}
		default:
			r, _ := utf8.DecodeRune(b)
			return Key{Type: KeyRune, Rune: r}
		}
	}

	// Escape sequences.
	if b[0] == 27 && len(b) >= 3 && b[1] == '[' {
		switch b[2] {
		case 'A':
			return Key{Type: KeyUp}
		case 'B':
			return Key{Type: KeyDown}
		case 'C':
			return Key{Type: KeyRight}
		case 'D':
			return Key{Type: KeyLeft}
		case 'H':
			return Key{Type: KeyHome}
		case 'F':
			return Key{Type: KeyEnd}
		case '3':
			if len(b) >= 4 && b[3] == '~' {
				return Key{Type: KeyDelete}
			}
		case '1':
			if len(b) >= 4 && b[3] == '~' {
				return Key{Type: KeyHome}
			}
		case '4':
			if len(b) >= 4 && b[3] == '~' {
				return Key{Type: KeyEnd}
			}
		}
		return Key{Type: KeyEscape}
	}

	if b[0] == 27 {
		return Key{Type: KeyEscape}
	}

	// Multi-byte UTF-8.
	r, _ := utf8.DecodeRune(b)
	return Key{Type: KeyRune, Rune: r}
}
