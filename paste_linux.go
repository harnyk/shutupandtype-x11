package main

import "os/exec"

// typeShiftInsert synthesizes a Shift+Insert keypress via xdotool so the
// transcribed text is pasted into the focused window right after being copied.
func typeShiftInsert() error {
	return exec.Command("xdotool", "key", "--clearmodifiers", "shift+Insert").Run()
}
