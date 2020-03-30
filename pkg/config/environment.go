package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Environment stores all parameters needed to interact with the GOV.UK Pay API
type Environment struct {
	Name    string
	APIKey  string
	BaseURL string
}

// CreateEnvironment commits the environment into the CLIs configuration file
func (environment *Environment) CreateEnvironment() error {
	err := environment.writeEnvironment()
	if err != nil {
		return err
	}
	return nil
}

// Init populates the values of the environment based on the config file
func (environment *Environment) Init() error {
	apiKey, err := environment.GetAPIKey()
	if err != nil {
		return err
	}
	baseURL, err := environment.GetBaseURL()
	if err != nil {
		return err
	}
	environment.APIKey = apiKey
	environment.BaseURL = baseURL
	return nil
}

func (environment *Environment) GetAPIKey() (string, error) {
	// if the API key exists on the currently running process
	if environment.APIKey != "" {
		return environment.APIKey, nil
	}

	// if the API key exists in the configuration
	if err := viper.ReadInConfig(); err == nil {
		return viper.GetString(environment.GetConfigParam("api_key")), nil
	}
	return "", errors.New("The CLI has not been configured with an API key. Use `pay link` to configure")
}

func (environment *Environment) GetBaseURL() (string, error) {
	if environment.BaseURL != "" {
		return environment.BaseURL, nil
	}

	if err := viper.ReadInConfig(); err == nil {
		return viper.GetString(environment.GetConfigParam("base_url")), nil
	}
	return "", errors.New("The CLI has not been configured with an a base URL. Use `pay link` to configure")
}

func (environment *Environment) writeEnvironment() error {
	file := viper.ConfigFileUsed()

	err := makePath(file)
	if err != nil {
		return err
	}

	viper.Set(environment.GetConfigParam("api_key"), strings.TrimSpace(environment.APIKey))
	viper.Set(environment.GetConfigParam("base_url"), strings.TrimSpace(environment.BaseURL))

	viper.MergeInConfig()

	viper.SetConfigFile(file)
	viper.SetConfigType(filepath.Ext(file))

	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func makePath(path string) error {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetConfigParam returns a namespaced parameter according to the current environment
func (environment *Environment) GetConfigParam(param string) string {
	var namespace string
	if strings.TrimSpace(environment.Name) == "" {
		namespace = "default"
	} else {
		namespace = environment.Name
	}
	return namespace + "." + param
}
