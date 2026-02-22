package main

import "os/exec"

func notify(summary, body string, urgency string) {
	exec.Command("notify-send",
		"--urgency", urgency,
		"--app-name", "shutupandtype",
		summary, body,
	).Run()
}

func notifyInfo(summary, body string)  { notify(summary, body, "normal") }
func notifyError(summary, body string) { notify(summary, body, "critical") }
