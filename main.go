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
	fmt.Printf("Press %s to start/stop recording (auto-stop after %s). Ctrl+C to quit.\n", hotkeyLabel(), timeout)

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
			setTrayTooltip("Recording failed: " + err.Error())
			return
		}
		fmt.Println(path)
		setTrayState(StateTranscribing)
		setTrayTooltip("Transcribing…")
		go func() {
			text, err := transcribe(path)
			resetIfIdle := func(d time.Duration) {
				time.AfterFunc(d, func() {
					mu.Lock()
					idle := !recording
					mu.Unlock()
					if idle {
						setTrayState(StateIdle)
					}
				})
			}
			if err != nil {
				log.Printf("transcribe: %v", err)
				setTrayState(StateError)
				setTrayTooltip("Transcription failed: " + err.Error())
				resetIfIdle(4 * time.Second)
				return
			}
			text = strings.TrimSpace(text)
			fmt.Println(text)
			if err := toClipboard(text); err != nil {
				log.Printf("clipboard: %v", err)
				setTrayState(StateError)
				setTrayTooltip("Clipboard error: " + err.Error())
				resetIfIdle(4 * time.Second)
				return
			}
			if err := typeShiftInsert(); err != nil {
				log.Printf("typeShiftInsert: %v", err)
			}
			setTrayState(StateDone)
			setTrayTooltip("Typed: " + preview(text))
			resetIfIdle(3 * time.Second)
		}()
	}

	onPress := func() {
		fmt.Printf("[%s] %s\n", timestamp(), hotkeyLabel())
		mu.Lock()
		if !recording {
			recording = true
			mu.Unlock()
			if err := rec.Start(); err != nil {
				log.Printf("recorder start: %v", err)
				setTrayState(StateError)
				setTrayTooltip("Recording failed to start: " + err.Error())
				time.AfterFunc(4*time.Second, func() { setTrayState(StateIdle) })
				mu.Lock()
				recording = false
				mu.Unlock()
				return
			}
			setTrayState(StateRecording)
			setTrayTooltip(fmt.Sprintf("Recording… (auto-stop in %s)", timeout))
			fmt.Printf("[%s] Recording started (auto-stop in %s)\n", timestamp(), timeout)
			mu.Lock()
			timer = time.AfterFunc(timeout, func() {
				fmt.Printf("[%s] Auto-stop timeout reached\n", timestamp())
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
	fmt.Println("\nStopping.")
	unregister()
	systray.Quit()
}

func timestamp() string {
	return time.Now().Format("15:04:05.000")
}
