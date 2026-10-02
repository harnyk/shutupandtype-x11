//go:build darwin

package main

import (
	"os/exec"
	"strings"
)

func platformShowNotification(title, body string) {
	script := "display notification " + appleScriptString(body) + " with title " + appleScriptString(title)
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		eventLogf("notification: %v", err)
	}
}

func appleScriptString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n', '\r':
			b.WriteString(" ")
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
