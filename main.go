package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
)

func main() {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("cannot connect to X11: %v", err)
	}
	defer xu.Conn().Close()

	keybind.Initialize(xu)
	root := xu.RootWin()

	codes := keybind.StrToKeycodes(xu, "Scroll_Lock")
	if len(codes) == 0 {
		log.Fatal("no keycode found for Scroll_Lock")
	}

	// Passive grab on root with AnyModifier — captures the key regardless of
	// other held modifiers. ownerEvents=false routes events only to us.
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
			log.Fatalf("GrabKey Scroll_Lock (keycode %d): %v", code, err)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\nStopping — ungrabbing Scroll_Lock.")
		for _, code := range codes {
			xproto.UngrabKey(xu.Conn(), code, root, xproto.ModMaskAny)
		}
		os.Exit(0)
	}()

	fmt.Println("Listening for Scroll_Lock (native behavior blocked) — press Ctrl+C to stop.")

	var pending xgb.Event
	pressed := false

	for {
		var ev xgb.Event
		if pending != nil {
			ev = pending
			pending = nil
		} else {
			var xerr xgb.Error
			ev, xerr = xu.Conn().WaitForEvent()
			if xerr != nil {
				log.Printf("X11 error: %v", xerr)
				continue
			}
			if ev == nil {
				break
			}
		}

		switch e := ev.(type) {
		case xproto.KeyPressEvent:
			if pressed {
				break // auto-repeat, ignore
			}
			pressed = true
			sym := keybind.LookupString(xu, e.State, e.Detail)
			fmt.Printf("[%s] KEYDOWN  key=%q state=0x%04x keycode=%d\n",
				timestamp(), sym, e.State, e.Detail)

		case xproto.KeyReleaseEvent:
			// Detect auto-repeat: X11 queues KeyRelease+KeyPress as a pair.
			next, _ := xu.Conn().PollForEvent()
			if kp, ok := next.(xproto.KeyPressEvent); ok && kp.Detail == e.Detail {
				break // auto-repeat pair, discard both
			}
			if next != nil {
				pending = next // unrelated event, process next iteration
			}
			pressed = false
			sym := keybind.LookupString(xu, e.State, e.Detail)
			fmt.Printf("[%s] KEYUP    key=%q state=0x%04x keycode=%d\n",
				timestamp(), sym, e.State, e.Detail)
		}
	}
}

func timestamp() string {
	return time.Now().Format("15:04:05.000")
}
