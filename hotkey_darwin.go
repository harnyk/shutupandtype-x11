package main

import (
	"log"

	"golang.design/x/hotkey"
)

// listenHotkey registers Ctrl+Shift+F12 globally and calls onPress on each
// key-down. Requires Accessibility permission for this process.
func listenHotkey(onPress func()) (unregister func()) {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyF12)
	if err := hk.Register(); err != nil {
		log.Fatalf("hotkey: register %s failed (grant Accessibility to the terminal/app): %v", hotkeyLabel(), err)
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
		<-done
	}
}
