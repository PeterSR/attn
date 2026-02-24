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
