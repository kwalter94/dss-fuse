package dssapi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Instances       map[string]InstanceParams `json:"dss_instances"`
	DefaultInstance *string                   `json:"default_instance"`
}

type InstanceParams struct {
	Url    string `json:"url"`
	ApiKey string `json:"api_key"`
}

func LoadConfig(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadUserConfig() (*Config, error) {
	homePath, found := os.LookupEnv("HOME")
	if !found {
		log.Fatalf("Environment variable HOME not found!")
	}

	configPath := fmt.Sprintf("%s/.dataiku/config.json", homePath)
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	return LoadConfig(configPath)
}

func (config *Config) GetDefaultInstance() *InstanceParams {
	if config.DefaultInstance == nil || config.Instances == nil {
		return nil
	}

	instance := config.Instances[*config.DefaultInstance]

	return &instance
}
