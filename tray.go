package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"runtime"

	"github.com/getlantern/systray"
	"github.com/spf13/viper"
)

type TrayState int

const (
	StateIdle         TrayState = iota // gray  — waiting
	StateRecording                     // red   — mic active
	StateTranscribing                  // yellow — API call in progress
	StateDone                          // green  — copied to clipboard
	StateError                         // orange — something went wrong
)

var trayIcons map[TrayState][]byte

func initTrayIcons() {
	trayIcons = map[TrayState][]byte{
		StateIdle:         circleIcon(130, 130, 130), // gray
		StateRecording:    circleIcon(220, 50, 50),   // red
		StateTranscribing: circleIcon(230, 170, 0),   // amber
		StateDone:         circleIcon(50, 200, 80),   // green
		StateError:        circleIcon(255, 100, 0),   // orange
	}
}

func setTrayState(s TrayState) {
	if icon, ok := trayIcons[s]; ok {
		systray.SetIcon(icon)
	}
}

func setTrayTooltip(text string) {
	systray.SetTooltip(text)
}

func trayReadyTooltip() string {
	return "Ready — " + hotkeyLabel()
}

func setTrayAction(state TrayState, tooltip string) {
	setTrayState(state)
	setTrayTooltip(tooltip)
}

func resetTrayIdle() {
	setTrayAction(StateIdle, trayReadyTooltip())
}

func circleIcon(r, g, b uint8) []byte {
	const size = 22
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2, float64(size)/2
	outer := float64(size)/2 - 1
	inner := outer - 1.2 // anti-alias edge width

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= inner {
				img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
			} else if dist <= outer {
				// soft edge
				alpha := uint8(255 * (outer - dist) / (outer - inner))
				img.SetRGBA(x, y, color.RGBA{r, g, b, alpha})
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func onTrayReady() {
	initTrayIcons()
	resetTrayIdle()

	if runtime.GOOS == "darwin" {
		mPerm := systray.AddMenuItem("Privacy settings…", "Open Accessibility + Input Monitoring")
		go func() {
			for range mPerm.ClickedCh {
				openPrivacySettings()
			}
		}()
		systray.AddSeparator()
	}

	mSmart := systray.AddMenuItemCheckbox("Smart mode", "LLM post-process after transcription", cfgSmartMode())
	if !smartModeCanEnable() {
		mSmart.Disable()
	}
	go func() {
		for range mSmart.ClickedCh {
			// Checked() is the pre-click state; the user toggles to the opposite.
			enable := !mSmart.Checked()
			if enable {
				if !smartModeCanEnable() {
					mSmart.Uncheck()
					actionNotifyError(
						"Smart mode needs smart_llm_api_key or openai_api_key",
						"smart mode: no API key configured")
					continue
				}
				mSmart.Check()
			} else {
				mSmart.Uncheck()
			}
			if err := setSmartModeAndPersist(enable); err != nil {
				viperMu.Lock()
				viper.Set("smart_mode", !enable)
				viperMu.Unlock()
				if enable {
					mSmart.Uncheck()
				} else {
					mSmart.Check()
				}
				actionNotifyError(
					"Could not save smart_mode to config: "+err.Error(),
					"persist config: %v", err)
				continue
			}
			if enable {
				actionNotify(nil, "Smart mode on", "smart mode enabled")
			} else {
				actionNotify(nil, "Smart mode off", "smart mode disabled")
			}
		}
	}()
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Stop shutupandtype")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onTrayExit() {}
