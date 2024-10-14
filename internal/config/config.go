package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	ConfigDir      = ".upify"
	ConfigFileName = "config.yaml"
)

type Config struct {
	Name           string             `yaml:"name"`
	Framework      Framework          `yaml:"framework,omitempty"`
	Language       Language           `yaml:"language"`
	RuntimeVersion string             `yaml:"runtime_version"`
	PackageManager PackageManager     `yaml:"package_manager"`
	Entrypoint     string             `yaml:"entrypoint,omitempty"`
	AppVar         string             `yaml:"app_var,omitempty"`
	AWSLambda      *AWSLambdaConfig   `yaml:"aws-lambda,omitempty"`
	GCPCloudRun    *GCPCloudRunConfig `yaml:"gcp-cloudrun,omitempty"`
}

type AWSLambdaConfig struct {
	Region   string `yaml:"region"`
	RoleName string `yaml:"role_name"`
	Runtime  string `yaml:"runtime"`
	Memory   int    `yaml:"memory,omitempty"`
}

type GCPCloudRunConfig struct {
	Region      string `yaml:"region"`
	ServiceName string `yaml:"service_name"`
	ProjectId   string `yaml:"project_id"`
	ImageURL    string `yaml:"image_url"`
	Runtime     string `yaml:"runtime"`
	Memory      int    `yaml:"memory,omitempty"`
}

func GetConfigFilePath() string {
	return filepath.Join(ConfigDir, ConfigFileName)
}

func LoadConfig() (*Config, error) {
	configFilePath := GetConfigFilePath()
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return nil, err
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if _, err := os.Stat(ConfigDir); os.IsNotExist(err) {
		os.Mkdir(ConfigDir, os.ModePerm)
	}

	configFilePath := GetConfigFilePath()
	err = os.WriteFile(configFilePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ConfigExists() bool {
	configFilePath := GetConfigFilePath()
	_, err := os.Stat(configFilePath)
	return !os.IsNotExist(err)
}
