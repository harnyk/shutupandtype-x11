package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureGUIPath prepends common Homebrew locations so GUI-launched apps
// (Finder / open / Launchpad) can find whisper-cli and ffmpeg.
func ensureGUIPath() {
	// Force UTF-8 so subprocesses / CoreFoundation don't fall back to MacRoman.
	if os.Getenv("LANG") == "" {
		os.Setenv("LANG", "en_US.UTF-8")
	}
	if os.Getenv("LC_ALL") == "" {
		os.Setenv("LC_ALL", "en_US.UTF-8")
	}
	os.Setenv("__CF_USER_TEXT_ENCODING", "0x0:0x8000100:0")

	extras := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
	}
	path := os.Getenv("PATH")
	parts := strings.Split(path, string(os.PathListSeparator))
	seen := map[string]bool{}
	for _, p := range parts {
		seen[p] = true
	}
	var prefix []string
	for _, dir := range extras {
		if seen[dir] {
			continue
		}
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			prefix = append(prefix, dir)
			seen[dir] = true
		}
	}
	if len(prefix) == 0 {
		return
	}
	os.Setenv("PATH", strings.Join(append(prefix, path), string(os.PathListSeparator)))
}

// resolveBin returns an absolute path if bin is a bare name found on PATH;
// otherwise returns bin unchanged.
func resolveBin(bin string) string {
	bin = strings.TrimSpace(bin)
	if bin == "" || filepath.IsAbs(bin) || strings.Contains(bin, string(os.PathSeparator)) {
		return bin
	}
	if p, err := exec.LookPath(bin); err == nil {
		return p
	}
	return bin
}
