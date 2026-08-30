package main

// preview returns the first 50 runes of s, with "…" appended if truncated.
func preview(s string) string {
	runes := []rune(s)
	if len(runes) <= 50 {
		return s
	}
	return string(runes[:50]) + "…"
}
