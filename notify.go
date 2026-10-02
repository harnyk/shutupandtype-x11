package main

// showNotification displays a short OS notification (errors and typed text only).
func showNotification(title, body string) {
	if title == "" {
		title = "ShutUpAndType"
	}
	if body == "" {
		return
	}
	platformShowNotification(title, body)
}
