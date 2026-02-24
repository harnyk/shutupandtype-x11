package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func toClipboard(text string) error {
	// Write to both PRIMARY (Shift+Insert) and CLIPBOARD (Ctrl+V).
	wrote := false
	for _, args := range [][]string{
		{"xclip", "-selection", "primary"},
		{"xclip", "-selection", "clipboard"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if cmd.Run() == nil {
			wrote = true
		}
	}
	if !wrote {
		return fmt.Errorf("xclip not available")
	}
	return nil
}

// typeShiftInsert synthesizes a Shift+Insert keypress via xdotool so the
// transcribed text is pasted into the focused window right after being copied.
func typeShiftInsert() error {
	return exec.Command("xdotool", "key", "--clearmodifiers", "shift+Insert").Run()
}

// preview returns the first 50 runes of s, with "…" appended if truncated.
func preview(s string) string {
	runes := []rune(s)
	if len(runes) <= 50 {
		return s
	}
	return string(runes[:50]) + "…"
}
