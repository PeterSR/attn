package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetCreatesFileAndDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.toml")

	if err := Set(path, "ntfy.topic", "my-topic"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "[ntfy]") {
		t.Errorf("expected [ntfy] section, got:\n%s", got)
	}
	if !strings.Contains(got, `topic = "my-topic"`) {
		t.Errorf("expected topic = \"my-topic\", got:\n%s", got)
	}
}

func TestSetExistingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	initial := "[ntfy]\ntopic = \"old-topic\"\nserver = \"https://ntfy.sh\"\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Set(path, "ntfy.topic", "new-topic"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, `topic = "new-topic"`) {
		t.Errorf("expected updated topic, got:\n%s", got)
	}
	if strings.Contains(got, "old-topic") {
		t.Errorf("old value still present:\n%s", got)
	}
	// Server should be preserved.
	if !strings.Contains(got, `server = "https://ntfy.sh"`) {
		t.Errorf("server line was lost:\n%s", got)
	}
}

func TestSetNewKeyInExistingSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	initial := "[ntfy]\nserver = \"https://ntfy.sh\"\n\n[desktop]\nwhen = \"active\"\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Set(path, "ntfy.topic", "my-topic"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, `topic = "my-topic"`) {
		t.Errorf("expected topic line, got:\n%s", got)
	}
	// Should appear between [ntfy] and [desktop].
	ntfyIdx := strings.Index(got, "[ntfy]")
	topicIdx := strings.Index(got, `topic = "my-topic"`)
	desktopIdx := strings.Index(got, "[desktop]")
	if topicIdx < ntfyIdx || topicIdx > desktopIdx {
		t.Errorf("topic not in ntfy section:\n%s", got)
	}
}

func TestSetNewSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	initial := "[desktop]\nwhen = \"active\"\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Set(path, "ntfy.topic", "my-topic"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "[ntfy]") {
		t.Errorf("expected [ntfy] section, got:\n%s", got)
	}
	if !strings.Contains(got, `topic = "my-topic"`) {
		t.Errorf("expected topic line, got:\n%s", got)
	}
}

func TestSetPreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	initial := "# My config\n[ntfy]\n# The topic\ntopic = \"old\"\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Set(path, "ntfy.topic", "new"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "# My config") {
		t.Errorf("top comment lost:\n%s", got)
	}
	if !strings.Contains(got, "# The topic") {
		t.Errorf("inline comment lost:\n%s", got)
	}
}

func TestSetInvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	err := Set(path, "bogus.key", "val")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetInvalidWhen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	err := Set(path, "desktop.when", "bogus")
	if err == nil {
		t.Fatal("expected error for invalid when")
	}
	if !strings.Contains(err.Error(), "invalid when value") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetValidWhen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	for _, w := range []string{"never", "active", "idle", "always"} {
		if err := Set(path, "desktop.when", w); err != nil {
			t.Errorf("Set desktop.when=%q failed: %v", w, err)
		}
	}
}

func TestSetTemplateValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	tmpl := `[{{.Repo}}:{{.Branch}}] `
	if err := Set(path, "format.prefix", tmpl); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, tmpl) {
		t.Errorf("template value not preserved:\n%s", got)
	}
}

func TestSetKeyMatchPrecision(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// "token" should not match "user_token" or "token_extra".
	initial := "[pushover]\nuser_key = \"old\"\ntoken = \"secret\"\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Set(path, "pushover.token", "new-secret"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, `token = "new-secret"`) {
		t.Errorf("token not updated:\n%s", got)
	}
	if !strings.Contains(got, `user_key = "old"`) {
		t.Errorf("user_key was modified:\n%s", got)
	}
}

func TestGetAfterSet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := Set(path, "ntfy.topic", "round-trip"); err != nil {
		t.Fatal(err)
	}

	val, err := Get(path, "ntfy.topic")
	if err != nil {
		t.Fatal(err)
	}
	if val != "round-trip" {
		t.Errorf("Get returned %q, want %q", val, "round-trip")
	}
}

func TestGetDefaultValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	// File does not exist — should return defaults.

	val, err := Get(path, "desktop.when")
	if err != nil {
		t.Fatal(err)
	}
	if val != "active" {
		t.Errorf("expected default desktop.when=active, got %q", val)
	}

	val, err = Get(path, "ntfy.server")
	if err != nil {
		t.Fatal(err)
	}
	if val != "https://ntfy.sh" {
		t.Errorf("expected default ntfy.server, got %q", val)
	}
}

func TestGetInvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	_, err := Get(path, "nope.bad")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestSetProcessLabel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := Set(path, "processes.code", "VS Code"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "[processes]") {
		t.Errorf("expected [processes] section, got:\n%s", got)
	}
	if !strings.Contains(got, `code = "VS Code"`) {
		t.Errorf("expected code = \"VS Code\", got:\n%s", got)
	}
}

func TestGetProcessLabel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := Set(path, "processes.code", "VS Code"); err != nil {
		t.Fatal(err)
	}

	val, err := Get(path, "processes.code")
	if err != nil {
		t.Fatal(err)
	}
	if val != "VS Code" {
		t.Errorf("Get processes.code = %q, want %q", val, "VS Code")
	}
}

func TestGetProcessLabelMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	val, err := Get(path, "processes.nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if val != "" {
		t.Errorf("Get processes.nonexistent = %q, want empty", val)
	}
}

func TestSetProcessEmptyName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	err := Set(path, "processes.", "label")
	if err == nil {
		t.Fatal("expected error for empty process name")
	}
}
