package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return !os.IsNotExist(err)
}

func CreateDefaultConfig(configPath string) error {
	// Set default values
	viper.SetDefault("key", "default_value")

	// Write default config to file
	err := viper.WriteConfigAs(configPath)
	if err != nil {
		return fmt.Errorf("unable to write default config to file: %w", err)
	}
	return nil
}
