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

// Info holds structured context about the current environment.
type Info struct {
	Dir     string // basename of CWD
	Path    string // full CWD path
	Repo    string // git repo name (basename of git toplevel)
	Branch  string // git branch name
	Process string // friendly label of the parent process (from config)
}

// Gather collects environment context (CWD, git repo/branch).
// Never returns an error; missing fields are left empty.
func Gather() Info {
	var info Info

	if wd, err := os.Getwd(); err == nil {
		info.Path = wd
		base := filepath.Base(wd)
		if base != "." && base != "/" {
			info.Dir = base
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	toplevel, err := gitCommand(ctx, "rev-parse", "--show-toplevel")
	if err == nil {
		name := filepath.Base(toplevel)
		if name != "" && name != "." && name != "/" {
			info.Repo = name
		}
	}

	branch, err := gitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil && branch != "" && branch != "HEAD" {
		info.Branch = branch
	}

	return info
}

func gitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
