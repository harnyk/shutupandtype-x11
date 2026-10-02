package main

import (
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
}

func eventLogf(format string, args ...any) {
	log.Printf("[%s] "+format, append([]any{timestamp()}, args...)...)
}

// actionNotify logs an action and shows it in the menu bar tooltip.
// Pass nil for state to update only the tooltip (keep current icon).
func actionNotify(state *TrayState, tooltip, logFormat string, logArgs ...any) {
	eventLogf(logFormat, logArgs...)
	if state != nil {
		setTrayAction(*state, tooltip)
	} else {
		setTrayTooltip(tooltip)
	}
}

func actionNotifyError(tooltip, logFormat string, logArgs ...any) {
	actionNotify(ptrTray(StateError), tooltip, logFormat, logArgs...)
	showNotification("ShutUpAndType", tooltip)
}

func actionNotifyTyped(text string) {
	shown := preview(text)
	actionNotify(ptrTray(StateDone), "Typed: "+shown, "typed: %s", shown)
	showNotification("ShutUpAndType", shown)
}
