package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() (map[string]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	envPath := filepath.Join(cwd, ".upify", ".env")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println("No .upify/.env file found, not adding environment variables")
		return make(map[string]string), nil
	}

	env, err := godotenv.Read(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .env file: %v", err)
	}

	return env, nil
}

func VerifyWrapperExists() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	wrapperPath := filepath.Join(cwd, "upify_wrapper.py")
	if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
		return fmt.Errorf("upify_wrapper.py not found in current working directory")
	}

	return nil
}
