package cmd

import "fmt"

// VersionCmd prints version information.
type VersionCmd struct{}

func (v *VersionCmd) Run() error {
	fmt.Println("attn " + version)
	return nil
}

// SetVersion sets the version string (called from main via ldflags).
func SetVersion(v string) {
	if v != "" {
		version = v
	}
}
