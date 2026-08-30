package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type whisperTranscriber struct {
	bin, model, lang string
}

func buildWhisperArgs(model, audio, lang string) []string {
	if lang == "" {
		lang = "auto"
	}
	return []string{"-m", model, "-f", audio, "-l", lang, "-nt", "-np"}
}

func parseWhisperStdout(s string) string {
	return strings.TrimSpace(s)
}

func (t *whisperTranscriber) Transcribe(audioPath string) (string, error) {
	cmd := exec.Command(t.bin, buildWhisperArgs(t.model, audioPath, t.lang)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("whisper-cli: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	text := parseWhisperStdout(stdout.String())
	if text == "" {
		return "", fmt.Errorf("whisper-cli returned empty transcript")
	}
	return text, nil
}
