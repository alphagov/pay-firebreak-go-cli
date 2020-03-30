package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Init ensures a default environment is set and configures viper configuration
func Init() {
	configFile := filepath.Join(getConfigFolder(), "config.toml")
	viper.SetConfigType("toml")
	viper.SetConfigFile(configFile)
	viper.SetConfigPermissions(os.FileMode(0600))
	viper.ReadInConfig()
}

func getConfigFolder() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configPath := filepath.Join(homeDirectory, ".config", "pay")
	return configPath
}
