package render

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/petersr/attn/internal/autocontext"
)

var funcMap = template.FuncMap{
	"env": os.Getenv,
}

// Render executes tmpl as a Go text/template with info as data.
// On parse or execution error, it logs a warning to stderr and returns
// the literal template string unchanged.
// An empty template string returns "".
func Render(tmpl string, info autocontext.Info) string {
	if tmpl == "" {
		return ""
	}

	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: template parse error: %v\n", err)
		return tmpl
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, info); err != nil {
		fmt.Fprintf(os.Stderr, "attn: template execute error: %v\n", err)
		return tmpl
	}

	return buf.String()
}
