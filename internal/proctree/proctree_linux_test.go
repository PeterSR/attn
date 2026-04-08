//go:build linux

package proctree

import (
	"os"
	"testing"
)

func TestAncestorsSelf(t *testing.T) {
	chain := Ancestors(os.Getpid())
	if len(chain) == 0 {
		t.Fatal("Ancestors(self) returned empty chain")
	}
	if chain[0] != os.Getpid() {
		t.Errorf("first entry = %d, want %d (self)", chain[0], os.Getpid())
	}
	if chain[len(chain)-1] != 1 {
		t.Errorf("last entry = %d, want 1 (init)", chain[len(chain)-1])
	}
}

func TestAncestorsNonexistent(t *testing.T) {
	chain := Ancestors(99999999)
	if len(chain) != 0 {
		t.Errorf("Ancestors(99999999) = %v, want empty", chain)
	}
}

func TestIsAncestorParent(t *testing.T) {
	if !IsAncestor(os.Getpid(), os.Getppid()) {
		t.Error("IsAncestor(self, parent) = false, want true")
	}
}

func TestIsAncestorSelf(t *testing.T) {
	if !IsAncestor(os.Getpid(), os.Getpid()) {
		t.Error("IsAncestor(self, self) = false, want true")
	}
}

func TestIsAncestorUnrelated(t *testing.T) {
	// PID 1 (init/systemd) is an ancestor of everything, verify it works.
	if !IsAncestor(os.Getpid(), 1) {
		t.Error("IsAncestor(self, 1) = false, want true")
	}
}

func TestIsAncestorReversed(t *testing.T) {
	// Parent is not a descendant of child.
	if IsAncestor(os.Getppid(), os.Getpid()) {
		t.Error("IsAncestor(parent, self) = true, want false")
	}
}

func TestAncestorsNamedSelf(t *testing.T) {
	chain := AncestorsNamed(os.Getpid())
	if len(chain) == 0 {
		t.Fatal("AncestorsNamed(self) returned empty chain")
	}
	if chain[0].PID != os.Getpid() {
		t.Errorf("first entry PID = %d, want %d (self)", chain[0].PID, os.Getpid())
	}
	for i, p := range chain {
		if p.Name == "" {
			t.Errorf("entry %d (PID %d) has empty name", i, p.PID)
		}
	}
}

func TestMatchKnown(t *testing.T) {
	chain := []ProcessInfo{
		{PID: 100, Name: "attn"},
		{PID: 99, Name: "bash"},
		{PID: 98, Name: "node"},
		{PID: 97, Name: "code"},
		{PID: 1, Name: "systemd"},
	}
	known := map[string]string{
		"code": "VS Code",
		"warp": "Warp",
	}

	got := MatchKnown(chain, known)
	if got != "VS Code" {
		t.Errorf("MatchKnown = %q, want %q", got, "VS Code")
	}
}

func TestMatchKnownSkipsSelf(t *testing.T) {
	chain := []ProcessInfo{
		{PID: 100, Name: "code"}, // self — should be skipped
		{PID: 99, Name: "bash"},
	}
	known := map[string]string{
		"code": "VS Code",
	}

	got := MatchKnown(chain, known)
	if got != "" {
		t.Errorf("MatchKnown should skip self, got %q", got)
	}
}

func TestCmdlineSelf(t *testing.T) {
	cmdline := Cmdline(os.Getpid())
	if cmdline == "" {
		t.Fatal("Cmdline(self) returned empty string")
	}
}

func TestCmdlineNonexistent(t *testing.T) {
	if got := Cmdline(99999999); got != "" {
		t.Errorf("Cmdline(99999999) = %q, want empty", got)
	}
}

func TestMatchKnownNoMatch(t *testing.T) {
	chain := []ProcessInfo{
		{PID: 100, Name: "attn"},
		{PID: 99, Name: "bash"},
	}
	known := map[string]string{
		"code": "VS Code",
	}

	got := MatchKnown(chain, known)
	if got != "" {
		t.Errorf("MatchKnown = %q, want empty", got)
	}
}
