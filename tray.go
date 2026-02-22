package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/getlantern/systray"
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
	systray.SetTooltip("ShutUpAndType")
	systray.SetTooltip("Scroll_Lock to start/stop recording")
	setTrayState(StateIdle)

	mQuit := systray.AddMenuItem("Quit", "Stop shutupandtype")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onTrayExit() {}
