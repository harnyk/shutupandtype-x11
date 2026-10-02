package main

import (
	"testing"
)

func TestAppleScriptString(t *testing.T) {
	got := appleScriptString(`say "hi"`)
	want := `"say \"hi\""`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
