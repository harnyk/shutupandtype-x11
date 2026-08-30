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
	"sync"
	"time"

	"golang.design/x/hotkey"
)

func openPrivacySettings() {
	urls := []string{
		"x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension?Privacy_Accessibility",
		"x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension?Privacy_ListenEvent",
		"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility",
		"x-apple.systempreferences:com.apple.preference.security?Privacy_ListenEvent",
	}
	for _, u := range urls {
		_ = exec.Command("open", u).Start()
	}
}

// listenHotkey registers Ctrl+Shift+F12. If permission is missing it keeps
// retrying so the user can flip the toggle without restarting the app.
func listenHotkey(onPress func()) (unregister func()) {
	stop := make(chan struct{})
	var stopOnce sync.Once
	doStop := func() { stopOnce.Do(func() { close(stop) }) }

	var (
		mu     sync.Mutex
		active *hotkey.Hotkey
		done   chan struct{}
	)

	tryRegister := func() bool {
		mu.Lock()
		defer mu.Unlock()
		if active != nil {
			return true
		}
		if C.isAXTrusted() == 0 {
			_ = C.promptAXTrust()
		}
		hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyF12)
		if err := hk.Register(); err != nil {
			log.Printf("hotkey: register %s failed: %v", hotkeyLabel(), err)
			return false
		}
		active = hk
		done = make(chan struct{})
		go func(hk *hotkey.Hotkey, done chan struct{}) {
			defer close(done)
			for range hk.Keydown() {
				onPress()
			}
		}(hk, done)
		log.Println("hotkey:", hotkeyLabel(), "registered")
		setTrayState(StateIdle)
		setTrayTooltip("Ready — " + hotkeyLabel())
		return true
	}

	if !tryRegister() {
		setTrayState(StateError)
		setTrayTooltip("Orange = no hotkey. Enable ShutUpAndType in Accessibility AND Input Monitoring")
		openPrivacySettings()
		go func() {
			t := time.NewTicker(2 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-stop:
					return
				case <-t.C:
					if tryRegister() {
						return
					}
				}
			}
		}()
	}

	return func() {
		doStop()
		mu.Lock()
		hk := active
		d := done
		active = nil
		mu.Unlock()
		if hk != nil {
			_ = hk.Unregister()
		}
		if d != nil {
			select {
			case <-d:
			case <-time.After(2 * time.Second):
			}
		}
	}
}
