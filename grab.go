package main

import (
	"log"

	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
)

// grabKey registers a passive grab for the named key on the root window.
// Returns the grabbed keycodes so the caller can ungrab on exit.
func grabKey(xu *xgbutil.XUtil, name string) []xproto.Keycode {
	codes := keybind.StrToKeycodes(xu, name)
	if len(codes) == 0 {
		log.Fatalf("no keycode found for %s", name)
	}
	root := xu.RootWin()
	for _, code := range codes {
		if err := xproto.GrabKeyChecked(
			xu.Conn(),
			false,
			root,
			xproto.ModMaskAny,
			code,
			xproto.GrabModeAsync,
			xproto.GrabModeAsync,
		).Check(); err != nil {
			log.Fatalf("GrabKey %s (keycode %d): %v", name, code, err)
		}
	}
	return codes
}

// ungrabKeys releases previously grabbed keycodes.
func ungrabKeys(xu *xgbutil.XUtil, codes []xproto.Keycode) {
	root := xu.RootWin()
	for _, code := range codes {
		xproto.UngrabKey(xu.Conn(), code, root, xproto.ModMaskAny)
	}
}
