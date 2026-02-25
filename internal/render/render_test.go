package render

import (
	"testing"

	"github.com/petersr/attn/internal/autocontext"
)

func TestRenderVariables(t *testing.T) {
	info := autocontext.Info{
		Dir:     "myproject",
		Path:    "/home/user/myproject",
		Repo:    "myrepo",
		Branch:  "main",
		Process: "VS Code",
	}

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{"dir", "{{.Dir}}", "myproject"},
		{"path", "{{.Path}}", "/home/user/myproject"},
		{"repo", "{{.Repo}}", "myrepo"},
		{"branch", "{{.Branch}}", "main"},
		{"process", "{{.Process}}", "VS Code"},
		{"process in title", "{{.Process}}: done", "VS Code: done"},
		{"combined", "[{{.Repo}}:{{.Branch}}] ", "[myrepo:main] "},
		{"plain text", "hello world", "hello world"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.tmpl, info)
			if got != tt.want {
				t.Errorf("Render(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestRenderEnvFunction(t *testing.T) {
	t.Setenv("ATTN_TEST_VAR", "hello")
	info := autocontext.Info{}

	got := Render(`{{env "ATTN_TEST_VAR"}}`, info)
	if got != "hello" {
		t.Errorf(`Render(env "ATTN_TEST_VAR") = %q, want "hello"`, got)
	}
}

func TestRenderEnvMissing(t *testing.T) {
	info := autocontext.Info{}

	got := Render(`{{env "ATTN_NONEXISTENT_VAR_12345"}}`, info)
	if got != "" {
		t.Errorf(`Render(env missing) = %q, want ""`, got)
	}
}

func TestRenderMalformedTemplate(t *testing.T) {
	info := autocontext.Info{Repo: "myrepo"}

	// Bad syntax — should return the literal string.
	tmpl := "{{.Repo"
	got := Render(tmpl, info)
	if got != tmpl {
		t.Errorf("Render(%q) = %q, want literal %q", tmpl, got, tmpl)
	}
}

func TestRenderMissingField(t *testing.T) {
	info := autocontext.Info{}

	// Valid template, but fields are empty — should render as empty strings.
	got := Render("[{{.Repo}}:{{.Branch}}]", info)
	if got != "[:]" {
		t.Errorf("Render with empty info = %q, want \"[:]\"", got)
	}
}
