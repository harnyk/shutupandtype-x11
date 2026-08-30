package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Transcriber interface {
	Transcribe(audioPath string) (string, error)
}

func newTranscriber() (Transcriber, error) {
	switch cfgBackend() {
	case "openai":
		return &openaiTranscriber{}, nil
	case "whisper":
		return &whisperTranscriber{
			bin:   viper.GetString("whisper_bin"),
			model: expandPath(viper.GetString("whisper_model")),
			lang:  viper.GetString("whisper_language"),
		}, nil
	default:
		return nil, fmt.Errorf("unknown backend %q", cfgBackend())
	}
}

func transcribe(audioPath string) (string, error) {
	t, err := newTranscriber()
	if err != nil {
		return "", err
	}
	return t.Transcribe(audioPath)
}
