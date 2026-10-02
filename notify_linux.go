//go:build linux

package main

import (
	"os/exec"
)

func platformShowNotification(title, body string) {
	if _, err := exec.LookPath("notify-send"); err != nil {
		eventLogf("notification: notify-send not found")
		return
	}
	cmd := exec.Command("notify-send", title, body)
	if err := cmd.Run(); err != nil {
		eventLogf("notification: %v", err)
	}
}
