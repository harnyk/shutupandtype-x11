package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Recorder captures audio from the default mic using arecord.
type Recorder struct {
	cmd  *exec.Cmd
	path string
}

// Start begins recording to a unique temp file. Returns an error if arecord
// cannot be launched or a recording is already in progress.
func (r *Recorder) Start() error {
	if r.cmd != nil {
		return fmt.Errorf("already recording")
	}
	r.path = filepath.Join(os.TempDir(),
		fmt.Sprintf("suatype-%d.wav", time.Now().UnixNano()))
	r.cmd = exec.Command("arecord", "-f", "cd", "-t", "wav", r.path)
	return r.cmd.Start()
}

// Stop ends the current recording and returns the path to the WAV file.
func (r *Recorder) Stop() (string, error) {
	if r.cmd == nil || r.cmd.Process == nil {
		return "", fmt.Errorf("not recording")
	}
	// SIGTERM lets arecord flush and close the WAV header properly.
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		_ = r.cmd.Process.Kill()
	}
	_ = r.cmd.Wait()
	path := r.path
	r.cmd = nil
	r.path = ""
	return path, nil
}
