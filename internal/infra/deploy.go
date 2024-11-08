package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/platform"
	"github.com/joho/godotenv"
)

func PreDeployValidate(cfg *config.Config, platform platform.Platform) error {
	terraformDir := GetPlatformTerraformDir(platform)
	_, err := os.Stat(filepath.Join(terraformDir, "main.tf"))
	if os.IsNotExist(err) {
		return fmt.Errorf("couldn't find %s/main.tf, did you run `upify platform add %s`?", terraformDir, platform)
	}

	handlerPath := GetHandlerPath(cfg.Language)
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		return fmt.Errorf("%s not found in current working directory", GetHandlerFileName(cfg.Language))
	}

	return nil
}

func WriteEnvironmentVariables(platform platform.Platform) error {
	envVars, err := loadEnvironmentVariables()
	if err != nil {
		return err
	}

	envVarsPath := filepath.Join(GetPlatformTerraformDir(platform), "env.auto.tfvars")
	file, err := os.Create(envVarsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("env_vars = {\n")
	for key, value := range envVars {
		file.WriteString(fmt.Sprintf("  %s = \"%s\"\n", key, strings.ReplaceAll(value, "\"", "\\\"")))
	}
	file.WriteString("}\n")

	return nil
}

// loadEnvironmentVariables attempts to load environment variables from a file.
// It first checks for a ".env.prod" file in the ".upify" directory of the current working directory.
// If ".env.prod" does not exist, it falls back to loading ".env".
// If neither file exists, it returns an empty map without an error.
func loadEnvironmentVariables() (map[string]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	envPath := filepath.Join(cwd, ".upify", ".env.prod")
	env, err := tryLoadEnvFile(envPath)
	if err == nil {
		return env, nil
	}

	fmt.Println("Falling back to .env as .env.prod was not found or failed to load.")
	envPath = filepath.Join(cwd, ".upify", ".env")
	env, err = tryLoadEnvFile(envPath)
	if err == nil {
		return env, nil
	}

	fmt.Println("No environment file found. Returning an empty environment map.")
	return map[string]string{}, nil
}

func tryLoadEnvFile(envPath string) (map[string]string, error) {
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", envPath)
	}

	env, err := godotenv.Read(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %v", envPath, err)
	}

	return env, nil
}
