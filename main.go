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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	root := &cobra.Command{
		Use:   "shutupandtype-x11",
		Short: "Press Ctrl+Shift+F12 to record and transcribe speech to clipboard",
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
	timeout := cfgTimeout()
	fmt.Printf("Press Ctrl+Shift+F12 to start/stop recording (auto-stop after %s). Ctrl+C to quit.\n", timeout)

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
			if err != nil {
				log.Printf("transcribe: %v", err)
				setTrayState(StateError)
				setTrayTooltip("Transcription failed: " + err.Error())
				time.AfterFunc(4*time.Second, func() { setTrayState(StateIdle) })
				return
			}
			text = strings.TrimSpace(text)
			fmt.Println(text)
			if err := toClipboard(text); err != nil {
				log.Printf("clipboard: %v", err)
				setTrayState(StateError)
				setTrayTooltip("Clipboard error: " + err.Error())
				time.AfterFunc(4*time.Second, func() { setTrayState(StateIdle) })
				return
			}
			setTrayState(StateDone)
			setTrayTooltip("Copied: " + preview(text))
			time.AfterFunc(3*time.Second, func() { setTrayState(StateIdle) })
		}()
	}

	onPress := func() {
		fmt.Printf("[%s] Ctrl+Shift+F12\n", timestamp())
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

	unregister := listenHotkey(onPress)

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
