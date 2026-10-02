package main

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

var viperMu sync.RWMutex

func shutupandtypeConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "shutupandtype", "config.yaml")
}

func persistViperConfigUnlocked() error {
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

func persistViperConfig() error {
	viperMu.Lock()
	defer viperMu.Unlock()
	return persistViperConfigUnlocked()
}

func setSmartModeAndPersist(enable bool) error {
	viperMu.Lock()
	defer viperMu.Unlock()
	viper.Set("smart_mode", enable)
	return persistViperConfigUnlocked()
}
