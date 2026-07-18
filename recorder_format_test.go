package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestAudioFormatForBackend(t *testing.T) {
	viper.Reset()
	viper.SetDefault("backend", "openai")
	if audioFormatForBackend() != "mp3" {
		t.Fatal("openai -> mp3")
	}
	viper.Set("backend", "whisper")
	if audioFormatForBackend() != "wav" {
		t.Fatal("whisper -> wav")
	}
}
