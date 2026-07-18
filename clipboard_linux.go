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
