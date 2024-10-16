package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeupify/upify/internal/config"
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

func VerifyWrapperExists(language config.Language) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	fileName := ""
	switch language {
	case config.Python:
		fileName = "upify_handler.py"
	case config.JavaScript, config.TypeScript:
		fileName = "upify_handler.js"
	}

	wrapperPath := filepath.Join(cwd, fileName)
	if _, err := os.Stat(wrapperPath); os.IsNotExist(err) {
		return fmt.Errorf("%s not found in current working directory", fileName)
	}

	return nil
}
