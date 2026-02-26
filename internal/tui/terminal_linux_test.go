//go:build linux

package tui

import "testing"

func TestParseKeyRune(t *testing.T) {
	k := ParseKey([]byte{'a'})
	if k.Type != KeyRune || k.Rune != 'a' {
		t.Fatalf("expected KeyRune 'a', got type=%d rune=%c", k.Type, k.Rune)
	}
}

func TestParseKeyEnter(t *testing.T) {
	k := ParseKey([]byte{13})
	if k.Type != KeyEnter {
		t.Fatalf("expected KeyEnter, got %d", k.Type)
	}
}

func TestParseKeyEscape(t *testing.T) {
	k := ParseKey([]byte{27})
	if k.Type != KeyEscape {
		t.Fatalf("expected KeyEscape, got %d", k.Type)
	}
}

func TestParseKeyBackspace(t *testing.T) {
	for _, b := range []byte{127, 8} {
		k := ParseKey([]byte{b})
		if k.Type != KeyBackspace {
			t.Fatalf("byte %d: expected KeyBackspace, got %d", b, k.Type)
		}
	}
}

func TestParseKeyCtrlC(t *testing.T) {
	k := ParseKey([]byte{3})
	if k.Type != KeyCtrlC {
		t.Fatalf("expected KeyCtrlC, got %d", k.Type)
	}
}

func TestParseKeyArrows(t *testing.T) {
	tests := []struct {
		name string
		seq  []byte
		want KeyType
	}{
		{"Up", []byte{27, '[', 'A'}, KeyUp},
		{"Down", []byte{27, '[', 'B'}, KeyDown},
		{"Right", []byte{27, '[', 'C'}, KeyRight},
		{"Left", []byte{27, '[', 'D'}, KeyLeft},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			k := ParseKey(tc.seq)
			if k.Type != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, k.Type)
			}
		})
	}
}

func TestParseKeyDelete(t *testing.T) {
	k := ParseKey([]byte{27, '[', '3', '~'})
	if k.Type != KeyDelete {
		t.Fatalf("expected KeyDelete, got %d", k.Type)
	}
}

func TestParseKeyHomeEnd(t *testing.T) {
	// CSI H and CSI F
	k := ParseKey([]byte{27, '[', 'H'})
	if k.Type != KeyHome {
		t.Fatalf("expected KeyHome from CSI H, got %d", k.Type)
	}

	k = ParseKey([]byte{27, '[', 'F'})
	if k.Type != KeyEnd {
		t.Fatalf("expected KeyEnd from CSI F, got %d", k.Type)
	}

	// CSI 1~ and CSI 4~
	k = ParseKey([]byte{27, '[', '1', '~'})
	if k.Type != KeyHome {
		t.Fatalf("expected KeyHome from CSI 1~, got %d", k.Type)
	}

	k = ParseKey([]byte{27, '[', '4', '~'})
	if k.Type != KeyEnd {
		t.Fatalf("expected KeyEnd from CSI 4~, got %d", k.Type)
	}
}

func TestParseKeyUTF8(t *testing.T) {
	k := ParseKey([]byte{0xc3, 0xa9}) // é
	if k.Type != KeyRune || k.Rune != 'é' {
		t.Fatalf("expected KeyRune 'é', got type=%d rune=%c", k.Type, k.Rune)
	}
}

func TestParseKeyUnknownEscapeSequence(t *testing.T) {
	k := ParseKey([]byte{27, '[', 'Z'}) // Unknown CSI sequence
	if k.Type != KeyEscape {
		t.Fatalf("expected KeyEscape for unknown sequence, got %d", k.Type)
	}
}

func TestParseKeyEmpty(t *testing.T) {
	k := ParseKey(nil)
	if k.Type != KeyRune || k.Rune != 0 {
		t.Fatalf("expected zero Key, got type=%d rune=%d", k.Type, k.Rune)
	}
}
