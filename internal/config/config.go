package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	ConfigDir      = ".upify"
	ConfigFileName = "config.yaml"
)

var (
	validFrameworks      = map[string]bool{"none": true, "flask": true, "express": true}
	validLanguages       = map[string]bool{"python": true, "javascript": true, "typescript": true}
	validPackageManagers = map[string]bool{"pip": true, "npm": true, "yarn": true}
)

type Config struct {
	Framework      string             `yaml:"framework"`
	Language       string             `yaml:"language"`
	PackageManager string             `yaml:"package_manager"`
	Entrypoint     string             `yaml:"entrypoint"`
	Name           string             `yaml:"name"`
	AWSLambda      *AWSLambdaConfig   `yaml:"aws-lambda,omitempty"`
	GCPCloudRun    *GCPCloudRunConfig `yaml:"gcp-cloudrun,omitempty"`
}

type AWSLambdaConfig struct {
	Region   string `yaml:"region"`
	RoleName string `yaml:"role_name"`
	Runtime  string `yaml:"runtime"`
}

type GCPCloudRunConfig struct {
	// TODO
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

func (c *Config) Validate() error {
	if c.Framework == "" {
		return fmt.Errorf("framework is not specified in the configuration")
	}
	if !validFrameworks[strings.ToLower(c.Framework)] {
		return fmt.Errorf("invalid framework: %s. Must be one of: none, flask, express", c.Framework)
	}

	if c.Language == "" {
		return fmt.Errorf("language is not specified in the configuration")
	}
	if !validLanguages[strings.ToLower(c.Language)] {
		return fmt.Errorf("invalid language: %s. Must be one of: python, javascript, typescript", c.Language)
	}

	if c.PackageManager == "" {
		return fmt.Errorf("package manager is not specified in the configuration")
	}
	if !validPackageManagers[strings.ToLower(c.PackageManager)] {
		return fmt.Errorf("invalid package manager: %s. Must be one of: pip, npm, yarn", c.PackageManager)
	}

	if c.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	if c.Name == "" {
		return fmt.Errorf("name is not specified in the configuration")
	}

	if c.AWSLambda != nil {
		if err := c.validateAWSLambda(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateAWSLambda() error {
	if c.AWSLambda.Region == "" {
		return fmt.Errorf("AWS Lambda region is not specified")
	}

	if c.AWSLambda.RoleName == "" {
		return fmt.Errorf("AWS Lambda role name is not specified")
	}

	if c.AWSLambda.Runtime == "" {
		return fmt.Errorf("AWS Lambda runtime is not specified")
	}

	if c.Language == "python" && !strings.HasPrefix(c.AWSLambda.Runtime, "python") {
		return fmt.Errorf("invalid runtime for Python: %s. Must start with 'python'", c.AWSLambda.Runtime)
	}
	if (c.Language == "javascript" || c.Language == "typescript") && !strings.HasPrefix(c.AWSLambda.Runtime, "nodejs") {
		return fmt.Errorf("invalid runtime for JavaScript/TypeScript: %s. Must start with 'nodejs'", c.AWSLambda.Runtime)
	}

	return nil
}
