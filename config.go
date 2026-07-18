package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot find home dir: %v", err)
	}

	cfgDir := filepath.Join(home, ".config", "shutupandtype")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)

	viper.SetDefault("openai_model_stt", "whisper-1")
	viper.SetDefault("timeout", "90s")
	viper.SetDefault("backend", "openai")
	viper.SetDefault("whisper_bin", "whisper-cli")
	viper.SetDefault("whisper_language", "")

	// Allow env var overrides (e.g. OPENAI_API_KEY still works).
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("config error: %v", err)
		}
	}
}

func cfgTimeout() time.Duration {
	d, err := time.ParseDuration(viper.GetString("timeout"))
	if err != nil {
		return 90 * time.Second
	}
	return d
}

func cfgBackend() string {
	b := strings.ToLower(strings.TrimSpace(viper.GetString("backend")))
	if b == "" {
		return "openai"
	}
	return b
}

func audioFormatForBackend() string {
	if cfgBackend() == "whisper" {
		return "wav"
	}
	return "mp3"
}

func expandPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func validateSTTConfig() error {
	switch cfgBackend() {
	case "openai":
		if viper.GetString("openai_api_key") == "" {
			return fmt.Errorf("openai_api_key not configured")
		}
		return nil
	case "whisper":
		bin := viper.GetString("whisper_bin")
		if bin == "" {
			return fmt.Errorf("whisper_bin not configured")
		}
		if _, err := exec.LookPath(bin); err != nil {
			if _, err2 := os.Stat(bin); err2 != nil {
				return fmt.Errorf("whisper_bin %q not found: %v", bin, err)
			}
		}
		model := expandPath(viper.GetString("whisper_model"))
		if model == "" {
			return fmt.Errorf("whisper_model not configured")
		}
		if _, err := os.Stat(model); err != nil {
			return fmt.Errorf("whisper_model %q: %w", model, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown backend %q (want openai or whisper)", cfgBackend())
	}
}
