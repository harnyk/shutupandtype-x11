package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	timeout := flag.Duration("timeout", 90*time.Second, "auto-stop recording after this duration")
	flag.Parse()

	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("cannot connect to X11: %v", err)
	}
	defer xu.Conn().Close()

	keybind.Initialize(xu)

	codes := grabKey(xu, "Scroll_Lock")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\nStopping — ungrabbing Scroll_Lock.")
		ungrabKeys(xu, codes)
		os.Exit(0)
	}()

	fmt.Printf("Press Scroll_Lock to start/stop recording (auto-stop after %s). Ctrl+C to quit.\n", *timeout)

	var (
		rec       Recorder
		mu        sync.Mutex
		recording bool
		timer     *time.Timer
	)

	stopRecording := func() {
		mu.Lock()
		defer mu.Unlock()
		if !recording {
			return
		}
		if timer != nil {
			timer.Stop()
			timer = nil
		}
		recording = false
		path, err := rec.Stop()
		if err != nil {
			log.Printf("recorder stop: %v", err)
			notifyError("🔴 Recording failed", err.Error())
			return
		}
		fmt.Println(path)
		notifyInfo("⏹️ Recording stopped", "Transcribing…")
		go func() {
			notifyInfo("⏳ Transcribing…", path)
			text, err := transcribe(path)
			if err != nil {
				log.Printf("transcribe: %v", err)
				notifyError("❌ Transcription failed", err.Error())
				return
			}
			text = strings.TrimSpace(text)
			fmt.Println(text)
			if err := toClipboard(text); err != nil {
				log.Printf("clipboard: %v", err)
				notifyError("❌ Clipboard error", err.Error())
				return
			}
			notifyInfo("📋 Copied to clipboard", preview(text))
		}()
	}

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
				log.Printf("X11 error: %v", xerr)
				continue
			}
			if ev == nil {
				break
			}
		}

		switch e := ev.(type) {
		case xproto.KeyPressEvent:
			// Ignore — we act on key release only.
			_ = e

		case xproto.KeyReleaseEvent:
			// During an active grab all keyboard events are delivered to us;
			// ignore releases that aren't Scroll_Lock.
			if !containsCode(codes, e.Detail) {
				break
			}

			// Detect auto-repeat: X11 queues KeyRelease+KeyPress as a pair.
			next, _ := xu.Conn().PollForEvent()
			if kp, ok := next.(xproto.KeyPressEvent); ok && kp.Detail == e.Detail {
				break // auto-repeat pair, discard both
			}
			if next != nil {
				pending = next
			}

			sym := keybind.LookupString(xu, e.State, e.Detail)
			fmt.Printf("[%s] KEY      key=%q state=0x%04x keycode=%d\n",
				timestamp(), sym, e.State, e.Detail)

			mu.Lock()
			if !recording {
				recording = true
				mu.Unlock()
				if err := rec.Start(); err != nil {
					log.Printf("recorder start: %v", err)
					notifyError("🔴 Recording failed to start", err.Error())
					mu.Lock()
					recording = false
					mu.Unlock()
					break
				}
				fmt.Printf("[%s] Recording started (auto-stop in %s)\n", timestamp(), *timeout)
				notifyInfo("🎙️ Recording started", fmt.Sprintf("Auto-stop in %s", *timeout))
				timer = time.AfterFunc(*timeout, func() {
					fmt.Printf("[%s] Auto-stop timeout reached\n", timestamp())
					stopRecording()
				})
			} else {
				mu.Unlock()
				stopRecording()
			}
		}
	}
}

func timestamp() string {
	return time.Now().Format("15:04:05.000")
}
