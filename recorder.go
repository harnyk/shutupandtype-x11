package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Recorder captures audio from the default mic using ffmpeg.
type Recorder struct {
	cmd  *exec.Cmd
	path string
}

// Start begins recording to a unique temp file (MP3 or WAV by backend).
func (r *Recorder) Start() error {
	if r.cmd != nil {
		return fmt.Errorf("already recording")
	}
	format := audioFormatForBackend()
	r.path = filepath.Join(os.TempDir(),
		fmt.Sprintf("suatype-%d.%s", time.Now().UnixNano(), format))
	args := ffmpegInputArgs()
	if format == "wav" {
		args = append(args, "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", r.path)
	} else {
		args = append(args, "-codec:a", "libmp3lame", "-q:a", "2", r.path)
	}
	r.cmd = exec.Command("ffmpeg", args...)
	return r.cmd.Start()
}

// Stop ends the current recording and returns the path to the audio file.
func (r *Recorder) Stop() (string, error) {
	if r.cmd == nil || r.cmd.Process == nil {
		return "", fmt.Errorf("not recording")
	}
	// SIGTERM/Interrupt lets ffmpeg flush and finalize the file properly.
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		_ = r.cmd.Process.Kill()
	}
	_ = r.cmd.Wait()
	path := r.path
	r.cmd = nil
	r.path = ""
	return path, nil
}
