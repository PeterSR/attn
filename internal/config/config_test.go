package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDesktopActive(t *testing.T) {
	cfg := Default()
	if cfg.Desktop.When != WhenActive {
		t.Errorf("Default desktop.when = %q, want %q", cfg.Desktop.When, WhenActive)
	}
}

func TestDefaultBellNever(t *testing.T) {
	cfg := Default()
	if cfg.Bell.When != WhenNever {
		t.Errorf("Default bell.when = %q, want %q", cfg.Bell.When, WhenNever)
	}
}

func TestLoadMigrateOldEnabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[desktop]
enabled = true

[bell]
enabled = true

[ntfy]
enabled = true
topic = "test"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Desktop.When != WhenActive {
		t.Errorf("desktop.when = %q, want %q (migrated from enabled=true)", cfg.Desktop.When, WhenActive)
	}
	if cfg.Bell.When != WhenAlways {
		t.Errorf("bell.when = %q, want %q (migrated from enabled=true)", cfg.Bell.When, WhenAlways)
	}
	if cfg.Ntfy.When != WhenAlways {
		t.Errorf("ntfy.when = %q, want %q (migrated from enabled=true)", cfg.Ntfy.When, WhenAlways)
	}
}

func TestLoadMigrateOldDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[desktop]
enabled = false

[bell]
enabled = false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Desktop.When != WhenNever {
		t.Errorf("desktop.when = %q, want %q (migrated from enabled=false)", cfg.Desktop.When, WhenNever)
	}
	if cfg.Bell.When != WhenNever {
		t.Errorf("bell.when = %q, want %q (migrated from enabled=false)", cfg.Bell.When, WhenNever)
	}
}

func TestLoadNewWhenField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[desktop]
when = "always"

[bell]
when = "idle"

[ntfy]
when = "idle"
topic = "test"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Desktop.When != WhenAlways {
		t.Errorf("desktop.when = %q, want %q", cfg.Desktop.When, WhenAlways)
	}
	if cfg.Bell.When != WhenIdle {
		t.Errorf("bell.when = %q, want %q", cfg.Bell.When, WhenIdle)
	}
	if cfg.Ntfy.When != WhenIdle {
		t.Errorf("ntfy.when = %q, want %q", cfg.Ntfy.When, WhenIdle)
	}
}

func TestLoadWhenOverridesEnabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// When both are set, "when" takes precedence.
	content := `
[desktop]
enabled = false
when = "always"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Desktop.When != WhenAlways {
		t.Errorf("desktop.when = %q, want %q (when should override enabled)", cfg.Desktop.When, WhenAlways)
	}
}

func TestLoadInvalidWhen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[desktop]
when = "bogus"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() with invalid when value should return error")
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Errorf("Load() with missing file should not error, got: %v", err)
	}
	if cfg.Desktop.When != WhenActive {
		t.Errorf("desktop.when = %q, want %q (default)", cfg.Desktop.When, WhenActive)
	}
}

func TestLoadRelayConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[relay]
when = "always"
socket_path = "/run/user/2000/attn.sock"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Relay.When != WhenAlways {
		t.Errorf("relay.when = %q, want %q", cfg.Relay.When, WhenAlways)
	}
	if cfg.Relay.SocketPath != "/run/user/2000/attn.sock" {
		t.Errorf("relay.socket_path = %q, want %q", cfg.Relay.SocketPath, "/run/user/2000/attn.sock")
	}
}

func TestLoadRelayDefaultNever(t *testing.T) {
	cfg := Default()
	if cfg.Relay.When != "" {
		t.Errorf("Default relay.when = %q, want empty (defaults to never)", cfg.Relay.When)
	}
}

func TestLoadRelayInvalidWhen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[relay]
when = "bogus"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() with invalid relay.when value should return error")
	}
}

func TestLoadFormatPrefix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[format]
prefix = "[{{.Repo}}:{{.Branch}}] "
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	want := "[{{.Repo}}:{{.Branch}}] "
	if cfg.Format.Prefix != want {
		t.Errorf("format.prefix = %q, want %q", cfg.Format.Prefix, want)
	}
}

func TestDefaultFormatPrefixEmpty(t *testing.T) {
	cfg := Default()
	if cfg.Format.Prefix != "" {
		t.Errorf("Default format.prefix = %q, want empty", cfg.Format.Prefix)
	}
}

func TestLoadProctreeMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[[proctree.marker]]
name = "node"
type = "delegate"
label = "WebTerm"
match_env = ["WEBTERM_ID"]
cmdline_contains = "webterm"

[[proctree.marker]]
name = "warp-terminal"
type = "focus_check"
label = "Warp"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Proctree.Markers) != 2 {
		t.Fatalf("got %d markers, want 2", len(cfg.Proctree.Markers))
	}

	m0 := cfg.Proctree.Markers[0]
	if m0.Name != "node" || m0.Type != "delegate" || m0.Label != "WebTerm" {
		t.Errorf("marker[0] = %+v", m0)
	}
	if len(m0.MatchEnv) != 1 || m0.MatchEnv[0] != "WEBTERM_ID" {
		t.Errorf("marker[0].MatchEnv = %v", m0.MatchEnv)
	}
	if m0.CmdlineContains != "webterm" {
		t.Errorf("marker[0].CmdlineContains = %q", m0.CmdlineContains)
	}

	m1 := cfg.Proctree.Markers[1]
	if m1.Name != "warp-terminal" || m1.Type != "focus_check" || m1.Label != "Warp" {
		t.Errorf("marker[1] = %+v", m1)
	}
}

func TestLoadSuppressIfEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[suppress]
if_env = ["IN_MEETING", "DND"]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"IN_MEETING", "DND"}
	if len(cfg.Suppress.IfEnv) != len(want) {
		t.Fatalf("Suppress.IfEnv = %v, want %v", cfg.Suppress.IfEnv, want)
	}
	for i, v := range want {
		if cfg.Suppress.IfEnv[i] != v {
			t.Errorf("Suppress.IfEnv[%d] = %q, want %q", i, cfg.Suppress.IfEnv[i], v)
		}
	}
}

func TestLoadForceIfEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[force]
if_env = ["ATTN_FORCE"]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Force.IfEnv) != 1 || cfg.Force.IfEnv[0] != "ATTN_FORCE" {
		t.Errorf("Force.IfEnv = %v", cfg.Force.IfEnv)
	}
}

func TestLoadMarkerMissingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[[proctree.marker]]
type = "delegate"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("Load() should fail when marker name is missing")
	}
}

func TestLoadMarkerMissingType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[[proctree.marker]]
name = "node"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("Load() should fail when marker type is missing")
	}
}

func TestLoadMarkerInvalidType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[[proctree.marker]]
name = "node"
type = "bogus"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should fail when marker type is invalid")
	}
	if !contains(err.Error(), "bogus") {
		t.Errorf("error %q should mention the bad value", err.Error())
	}
}

func TestLoadDefaultsUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Desktop.When != WhenActive {
		t.Errorf("desktop.when = %q, want %q", cfg.Desktop.When, WhenActive)
	}
	if cfg.Bell.When != WhenNever {
		t.Errorf("bell.when = %q, want %q", cfg.Bell.When, WhenNever)
	}
	if len(cfg.Proctree.Markers) != 0 {
		t.Errorf("Proctree.Markers = %v, want empty", cfg.Proctree.Markers)
	}
	if len(cfg.Suppress.IfEnv) != 0 {
		t.Errorf("Suppress.IfEnv = %v, want empty", cfg.Suppress.IfEnv)
	}
	if len(cfg.Force.IfEnv) != 0 {
		t.Errorf("Force.IfEnv = %v, want empty", cfg.Force.IfEnv)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestWhenValid(t *testing.T) {
	valid := []When{WhenNever, WhenActive, WhenIdle, WhenAlways, ""}
	for _, w := range valid {
		if !w.Valid() {
			t.Errorf("When(%q).Valid() = false, want true", w)
		}
	}

	if When("bogus").Valid() {
		t.Error(`When("bogus").Valid() = true, want false`)
	}
}
