package main

import (
	"fmt"
	"os/exec"
)

// typeShiftInsert pastes via Cmd+V using System Events (needs Accessibility,
// already required for the global hotkey).
func typeShiftInsert() error {
	err := exec.Command("osascript", "-e",
		`tell application "System Events" to keystroke "v" using command down`,
	).Run()
	if err != nil {
		return fmt.Errorf("paste (Cmd+V): %w", err)
	}
	return nil
}
