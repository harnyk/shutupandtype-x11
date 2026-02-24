package main

import (
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
)

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

	log.Println("hotkey: Ctrl+Shift+F12 registered")

	done := make(chan struct{})

	go func() {
		defer close(done)
		var pending xgb.Event
		for {
			var ev xgb.Event
			if pending != nil {
				ev = pending
				pending = nil
			} else {
				var xerr xgb.Error
				ev, xerr = xu.Conn().WaitForEvent()
				if xerr != nil {
					log.Printf("hotkey: X11 error: %v", xerr)
					continue
				}
				if ev == nil {
					return // connection closed — clean exit
				}
			}

			switch e := ev.(type) {
			case xproto.KeyReleaseEvent:
				if !containsCode(codes, e.Detail) {
					break
				}
				// Skip auto-repeat: a KeyRelease immediately followed by
				// a KeyPress for the same key is a synthetic repeat pair.
				next, _ := xu.Conn().PollForEvent()
				if kp, ok := next.(xproto.KeyPressEvent); ok && kp.Detail == e.Detail {
					break // it's a repeat — ignore both
				}
				if next != nil {
					pending = next
				}

				// Check modifiers: require Ctrl+Shift, ignore the rest.
				if e.State&hotkeyMods == hotkeyMods {
					onPress()
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
