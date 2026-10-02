package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func shutupandtypeConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "shutupandtype", "config.yaml")
}

func persistViperConfig() error {
	path := viper.ConfigFileUsed()
	if path == "" {
		path = shutupandtypeConfigPath()
		if path == "" {
			return os.ErrInvalid
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		viper.SetConfigFile(path)
	}
	if viper.ConfigFileUsed() != "" {
		return viper.WriteConfig()
	}
	return viper.SafeWriteConfigAs(path)
}
