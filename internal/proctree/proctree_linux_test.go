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
