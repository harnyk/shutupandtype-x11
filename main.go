package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	root := &cobra.Command{
		Use:   "shutupandtype-x11",
		Short: "Press Scroll_Lock to record and transcribe speech to clipboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			initConfig()
			enforceInstance()
			go run()
			systray.Run(onTrayReady, onTrayExit)
			return nil
		},
		SilenceUsage: true,
	}

	root.Flags().String("timeout", "90s", "auto-stop recording after this duration")
	_ = viper.BindPFlag("timeout", root.Flags().Lookup("timeout"))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func enforceInstance() {
	f, err := os.OpenFile("/tmp/shutupandtype.lock", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("cannot open lock file: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		log.Fatal("another instance is already running")
	}
	// intentionally not closing — lock held for process lifetime
}

func run() {
	xu, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("cannot connect to X11: %v", err)
	}
	defer xu.Conn().Close()

	keybind.Initialize(xu)
	codes := grabKey(xu, "Scroll_Lock")

	timeout := cfgTimeout()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\nStopping — ungrabbing Scroll_Lock.")
		ungrabKeys(xu, codes)
		systray.Quit()
		os.Exit(0)
	}()

	fmt.Printf("Press Scroll_Lock to start/stop recording (auto-stop after %s). Ctrl+C to quit.\n", timeout)

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
			setTrayState(StateError)
			notifyError("🔴 Recording failed", err.Error())
			return
		}
		fmt.Println(path)
		setTrayState(StateTranscribing)
		notifyInfo("⏹️ Recording stopped", "Transcribing…")
		go func() {
			notifyInfo("⏳ Transcribing…", path)
			text, err := transcribe(path)
			if err != nil {
				log.Printf("transcribe: %v", err)
				setTrayState(StateError)
				notifyError("❌ Transcription failed", err.Error())
				time.AfterFunc(4*time.Second, func() { setTrayState(StateIdle) })
				return
			}
			text = strings.TrimSpace(text)
			fmt.Println(text)
			if err := toClipboard(text); err != nil {
				log.Printf("clipboard: %v", err)
				setTrayState(StateError)
				notifyError("❌ Clipboard error", err.Error())
				time.AfterFunc(4*time.Second, func() { setTrayState(StateIdle) })
				return
			}
			setTrayState(StateDone)
			notifyInfo("📋 Copied to clipboard", preview(text))
			time.AfterFunc(3*time.Second, func() { setTrayState(StateIdle) })
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
			_ = e

		case xproto.KeyReleaseEvent:
			if !containsCode(codes, e.Detail) {
				break
			}
			next, _ := xu.Conn().PollForEvent()
			if kp, ok := next.(xproto.KeyPressEvent); ok && kp.Detail == e.Detail {
				break
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
					setTrayState(StateError)
					notifyError("🔴 Recording failed to start", err.Error())
					time.AfterFunc(4*time.Second, func() { setTrayState(StateIdle) })
					mu.Lock()
					recording = false
					mu.Unlock()
					break
				}
				setTrayState(StateRecording)
				fmt.Printf("[%s] Recording started (auto-stop in %s)\n", timestamp(), timeout)
				notifyInfo("🎙️ Recording started", fmt.Sprintf("Auto-stop in %s", timeout))
				timer = time.AfterFunc(timeout, func() {
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
