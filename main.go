package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	ensureGUIPath()

	root := &cobra.Command{
		Use:   "shutupandtype-x11",
		Short: "Hotkey to record and transcribe speech to clipboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			initConfig()
			if err := validateSTTConfig(); err != nil {
				return err
			}
			enforceInstance()

			hotkeyCh := make(chan func())
			unregCh := make(chan func(), 1)
			go run(hotkeyCh, unregCh)
			systray.Run(func() {
				onTrayReady()
				// Register on the systray main thread (required for macOS CGEventTap).
				onPress := <-hotkeyCh
				unregCh <- listenHotkey(onPress)
			}, onTrayExit)
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
	path := filepath.Join(os.TempDir(), "shutupandtype.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("cannot open lock file: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		log.Fatal("another instance is already running")
	}
	// intentionally not closing — lock held for process lifetime
}

func run(hotkeyCh chan<- func(), unregCh <-chan func()) {
	timeout := cfgTimeout()
	actionNotify(nil, fmt.Sprintf("Ready — %s (auto-stop %s)", hotkeyLabel(), timeout),
		"Press %s to start/stop recording (auto-stop after %s). Ctrl+C to quit.", hotkeyLabel(), timeout)

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
		setTrayAction(StateTranscribing, "Stopping recording…")
		if timer != nil {
			timer.Stop()
			timer = nil
		}
		recording = false
		path, err := rec.Stop()
		if err != nil {
			actionNotifyError("Recording failed: "+err.Error(), "recorder stop: %v", err)
			return
		}
		actionNotify(ptrTray(StateTranscribing), "Transcribing…", "audio: %s", path)
		go func() {
			text, err := transcribe(path)
			resetIfIdle := func(d time.Duration) {
				time.AfterFunc(d, func() {
					mu.Lock()
					idle := !recording
					mu.Unlock()
					if idle {
						resetTrayIdle()
					}
				})
			}
			if err != nil {
				actionNotifyError("Transcription failed: "+err.Error(), "transcribe: %v", err)
				resetIfIdle(4 * time.Second)
				return
			}
			text = strings.TrimSpace(text)
			actionNotify(ptrTray(StateTranscribing), "Transcript: "+preview(text), "transcript: %s", text)
			rawText := text
			if cfgSmartMode() {
				setTrayTooltip("Smart edit…")
				smartOut, smartErr := smartTransform(rawText)
				if smartErr != nil {
					actionNotifyError(
						"Smart edit failed (using raw transcript): "+smartErr.Error(),
						"smart: %v", smartErr)
					text = rawText
				} else {
					text = strings.TrimSpace(smartOut)
					actionNotify(ptrTray(StateTranscribing), "Smart: "+preview(text), "transcript (smart): %s", text)
				}
			}
			setTrayTooltip("Copying to clipboard…")
			if err := toClipboard(text); err != nil {
				actionNotifyError("Clipboard error: "+err.Error(), "clipboard: %v", err)
				resetIfIdle(4 * time.Second)
				return
			}
			setTrayTooltip("Pasting…")
			if err := typeShiftInsert(); err != nil {
				actionNotifyError("Paste failed: "+err.Error(), "typeShiftInsert: %v", err)
				resetIfIdle(4 * time.Second)
				return
			}
			actionNotifyTyped(text)
			resetIfIdle(3 * time.Second)
		}()
	}

	onPress := func() {
		actionNotify(nil, "Hotkey "+hotkeyLabel(), "hotkey %s", hotkeyLabel())
		mu.Lock()
		if !recording {
			recording = true
			mu.Unlock()
			if err := rec.Start(); err != nil {
				actionNotifyError("Recording failed to start: "+err.Error(), "recorder start: %v", err)
				time.AfterFunc(4*time.Second, func() { resetTrayIdle() })
				mu.Lock()
				recording = false
				mu.Unlock()
				return
			}
			actionNotify(ptrTray(StateRecording),
				fmt.Sprintf("Recording… (auto-stop in %s)", timeout),
				"recording started (auto-stop in %s)", timeout)
			mu.Lock()
			timer = time.AfterFunc(timeout, func() {
				actionNotify(nil, "Auto-stop timeout reached", "auto-stop timeout reached")
				stopRecording()
			})
			mu.Unlock()
		} else {
			mu.Unlock()
			stopRecording()
		}
	}

	hotkeyCh <- onPress
	unregister := <-unregCh

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	actionNotify(nil, "Stopping", "stopping")
	unregister()
	systray.Quit()
}

func ptrTray(s TrayState) *TrayState {
	return &s
}

func timestamp() string {
	return time.Now().Format("15:04:05.000")
}
