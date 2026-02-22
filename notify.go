package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func toClipboard(text string) error {
	// Try xclip first (X11), fall back to wl-copy (Wayland).
	for _, args := range [][]string{
		{"xclip", "-selection", "clipboard"},
		{"wl-copy"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no clipboard tool available (tried xclip, wl-copy)")
}

// preview returns the first 50 runes of s, with "…" appended if truncated.
func preview(s string) string {
	runes := []rune(s)
	if len(runes) <= 50 {
		return s
	}
	return string(runes[:50]) + "…"
}

func notify(summary, body string, urgency string) {
	exec.Command("notify-send",
		"--urgency", urgency,
		"--app-name", "shutupandtype",
		summary, body,
	).Run()
}

func notifyInfo(summary, body string)  { notify(summary, body, "normal") }
func notifyError(summary, body string) { notify(summary, body, "critical") }
