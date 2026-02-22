package main

import (
	"log"
	"os"
	"path/filepath"
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
