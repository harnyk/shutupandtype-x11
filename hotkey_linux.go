package main

import (
	"log"
	"time"

	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
)

// debounceWindow guards against switch bounce on flaky macro keys: presses
// within this window of the last accepted trigger are ignored.
const debounceWindow = 300 * time.Millisecond

// hotkeyCombo is the hotkey we register: Ctrl+Shift+F12.
// keysym 0xffc9 = XK_F12.
const hotkeySymName = "F12"
const hotkeyMods = xproto.ModMaskControl | xproto.ModMaskShift

// ignoredModMasks are the extra modifier bits to tolerate on top of hotkeyMods
// (NumLock=Mod2, CapsLock=Lock, ScrollLock=Mod5 and their combos).
// Grabbing each combo explicitly (instead of ModMaskAny) ensures bare F12
// and other unrelated combos are NOT intercepted.
var ignoredModMasks = []uint16{
	0,
	xproto.ModMaskLock,                   // CapsLock
	xproto.ModMask2,                      // NumLock
	xproto.ModMask5,                      // ScrollLock
	xproto.ModMaskLock | xproto.ModMask2, // CapsLock+NumLock
	xproto.ModMaskLock | xproto.ModMask5, // CapsLock+ScrollLock
	xproto.ModMask2 | xproto.ModMask5,    // NumLock+ScrollLock
	xproto.ModMaskLock | xproto.ModMask2 | xproto.ModMask5, // all three
}

// listenHotkey grabs Ctrl+Shift+F12 globally and calls onPress on each
// key-down event.  Only the exact modifier set (plus NumLock/CapsLock/
// ScrollLock variants) is grabbed, so bare F12 reaches other applications.
// Returns an unregister func that releases the grab and stops the loop.
func listenHotkey(onPress func()) (unregister func()) {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("hotkey: cannot connect to X11: %v", err)
	}
	keybind.Initialize(xu)

	codes := keybind.StrToKeycodes(xu, hotkeySymName)
	if len(codes) == 0 {
		log.Fatalf("hotkey: no keycode for F12")
	}
	root := xu.RootWin()
	for _, code := range codes {
		for _, extra := range ignoredModMasks {
			mods := hotkeyMods | extra
			if err := xproto.GrabKeyChecked(
				xu.Conn(),
				false,
				root,
				mods,
				code,
				xproto.GrabModeAsync,
				xproto.GrabModeAsync,
			).Check(); err != nil {
				log.Fatalf("hotkey: GrabKey F12 (keycode %d, mods %d): %v", code, mods, err)
			}
		}
	}

	log.Println("hotkey:", hotkeyLabel(), "registered")

	done := make(chan struct{})

	go func() {
		defer close(done)
		down := false
		var lastTrigger time.Time
		for {
			ev, xerr := xu.Conn().WaitForEvent()
			if xerr != nil {
				log.Printf("hotkey: X11 error: %v", xerr)
				continue
			}
			if ev == nil {
				return // connection closed — clean exit
			}

			switch e := ev.(type) {
			case xproto.KeyPressEvent:
				if !containsCode(codes, e.Detail) {
					break
				}
				if down {
					break // auto-repeat — ignore until released
				}
				down = true

				now := time.Now()
				if now.Sub(lastTrigger) < debounceWindow {
					break // debounced — likely switch bounce
				}

				// Check modifiers: require Ctrl+Shift, ignore the rest.
				if e.State&hotkeyMods == hotkeyMods {
					lastTrigger = now
					onPress()
				}
			case xproto.KeyReleaseEvent:
				if containsCode(codes, e.Detail) {
					down = false
				}
			}
		}
	}()

	return func() {
		for _, code := range codes {
			for _, extra := range ignoredModMasks {
				xproto.UngrabKey(xu.Conn(), code, root, hotkeyMods|extra)
			}
		}
		xu.Conn().Close() // unblocks WaitForEvent → goroutine exits
		<-done
	}
}

func containsCode(codes []xproto.Keycode, code xproto.Keycode) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}
	return false
}
