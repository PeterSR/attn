package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Set validates and writes a key=value pair into the config file at path.
// Creates the file and parent directories if they don't exist.
func Set(path, key, value string) error {
	section, tomlKey, validate, err := LookupKey(key)
	if err != nil {
		return err
	}
	if validate != nil {
		if err := validate(value); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := splitLines(string(data))
	lines = setKeyInLines(lines, section, tomlKey, value)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	return atomicWrite(path, joinLines(lines))
}

// setKeyInLines edits or inserts a key in the given TOML lines.
func setKeyInLines(lines []string, section, key, value string) []string {
	sectionHeader := "[" + section + "]"
	quotedValue := fmt.Sprintf("%q", value)
	newLine := key + " = " + quotedValue

	// Find the section.
	sectionIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			sectionIdx = i
			break
		}
	}

	if sectionIdx == -1 {
		// Section not found — append it.
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, sectionHeader, newLine)
		return lines
	}

	// Scan within section for the key or next section header.
	for i := sectionIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// Hit the next section — key not found in this section.
		if strings.HasPrefix(trimmed, "[") {
			// Insert before this next section header.
			lines = insertLine(lines, i, newLine)
			return lines
		}

		// Check if this line sets the key.
		if matchesKey(trimmed, key) {
			lines[i] = newLine
			return lines
		}
	}

	// Reached end of file within the section — append key.
	lines = append(lines, newLine)
	return lines
}

// matchesKey returns true if the line is a TOML assignment for the given key.
func matchesKey(trimmed, key string) bool {
	eqIdx := strings.Index(trimmed, "=")
	if eqIdx < 0 {
		return false
	}
	return strings.TrimSpace(trimmed[:eqIdx]) == key
}

func insertLine(lines []string, at int, line string) []string {
	lines = append(lines, "")
	copy(lines[at+1:], lines[at:])
	lines[at] = line
	return lines
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func joinLines(lines []string) string {
	result := strings.Join(lines, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

func atomicWrite(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
