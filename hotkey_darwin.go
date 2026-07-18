package main

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>

static int promptAXTrust(void) {
	const void *keys[] = { kAXTrustedCheckOptionPrompt };
	const void *values[] = { kCFBooleanTrue };
	CFDictionaryRef opts = CFDictionaryCreate(
		kCFAllocatorDefault, keys, values, 1,
		&kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	Boolean ok = AXIsProcessTrustedWithOptions(opts);
	if (opts) CFRelease(opts);
	return ok ? 1 : 0;
}

static int isAXTrusted(void) {
	return AXIsProcessTrusted() ? 1 : 0;
}
*/
import "C"
import (
	"log"
	"os/exec"
	"time"

	"golang.design/x/hotkey"
)

func openPrivacySettings() {
	// Best-effort deep links (vary by macOS version).
	urls := []string{
		"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility",
		"x-apple.systempreferences:com.apple.preference.security?Privacy_ListenEvent",
		"x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension?Privacy_Accessibility",
		"x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension?Privacy_ListenEvent",
	}
	for _, u := range urls {
		_ = exec.Command("open", u).Start()
	}
}

// listenHotkey registers Ctrl+Shift+F12 globally and calls onPress on each
// key-down. Requires Accessibility + Input Monitoring for this .app.
func listenHotkey(onPress func()) (unregister func()) {
	if C.isAXTrusted() == 0 {
		log.Println("hotkey: requesting Accessibility permission…")
		_ = C.promptAXTrust()
		openPrivacySettings()
	}

	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyF12)
	if err := hk.Register(); err != nil {
		log.Printf("hotkey: register %s failed: %v", hotkeyLabel(), err)
		setTrayState(StateError)
		setTrayTooltip("Enable Accessibility + Input Monitoring for ShutUpAndType, then restart")
		openPrivacySettings()
		// Keep the app alive so the tray stays visible; hotkey won't work until restart after grant.
		return func() {}
	}
	log.Println("hotkey:", hotkeyLabel(), "registered")

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range hk.Keydown() {
			onPress()
		}
	}()

	return func() {
		_ = hk.Unregister()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}
}
