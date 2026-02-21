package focus

import "regexp"

// ShouldSuppress returns true if the currently focused window matches
// the given regex pattern, indicating the notification should be skipped.
// Returns false if the pattern is empty, focus detection fails, or
// the focused window doesn't match.
func ShouldSuppress(pattern string) bool {
	if pattern == "" {
		return false
	}

	cls := FocusedWindow()
	if cls == "" {
		return false
	}

	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return false
	}

	return re.MatchString(cls)
}
