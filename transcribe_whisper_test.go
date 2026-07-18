package main

import (
	"strings"
	"testing"
)

func TestBuildWhisperArgs_withLanguage(t *testing.T) {
	args := buildWhisperArgs("/m.bin", "/a.wav", "ru")
	want := []string{"-m", "/m.bin", "-f", "/a.wav", "-l", "ru", "-nt", "-np"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", args, want)
	}
}

func TestBuildWhisperArgs_emptyLanguageUsesAuto(t *testing.T) {
	args := buildWhisperArgs("/m.bin", "/a.wav", "")
	want := []string{"-m", "/m.bin", "-f", "/a.wav", "-l", "auto", "-nt", "-np"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", args, want)
	}
}

func TestParseWhisperStdout(t *testing.T) {
	if got := parseWhisperStdout("  hello \n"); got != "hello" {
		t.Fatalf("got %q", got)
	}
	if got := parseWhisperStdout("   \n"); got != "" {
		t.Fatalf("empty expected, got %q", got)
	}
}
