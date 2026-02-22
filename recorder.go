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

// Start begins recording to a unique temp MP3 file.
func (r *Recorder) Start() error {
	if r.cmd != nil {
		return fmt.Errorf("already recording")
	}
	r.path = filepath.Join(os.TempDir(),
		fmt.Sprintf("suatype-%d.mp3", time.Now().UnixNano()))
	r.cmd = exec.Command("ffmpeg",
		"-f", "alsa", "-i", "default",
		"-codec:a", "libmp3lame", "-q:a", "2",
		r.path,
	)
	return r.cmd.Start()
}

// Stop ends the current recording and returns the path to the MP3 file.
func (r *Recorder) Stop() (string, error) {
	if r.cmd == nil || r.cmd.Process == nil {
		return "", fmt.Errorf("not recording")
	}
	// SIGTERM lets ffmpeg flush and finalize the file properly.
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		_ = r.cmd.Process.Kill()
	}
	_ = r.cmd.Wait()
	path := r.path
	r.cmd = nil
	r.path = ""
	return path, nil
}
