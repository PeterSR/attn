package autocontext

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const gitTimeout = 200 * time.Millisecond

// Derive returns a context string based on the current environment.
// Tries git repo name + branch first, falls back to directory basename.
// Returns "" if nothing useful can be determined. Never returns an error.
func Derive() string {
	if ctx := deriveGit(); ctx != "" {
		return ctx
	}
	return derivePWD()
}

func deriveGit() string {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	toplevel, err := gitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}

	repoName := filepath.Base(toplevel)
	if repoName == "" || repoName == "." || repoName == "/" {
		return ""
	}

	branch, err := gitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil || branch == "" || branch == "HEAD" {
		return repoName
	}

	return repoName + ":" + branch
}

func derivePWD() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	base := filepath.Base(wd)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

func gitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
