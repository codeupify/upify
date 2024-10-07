package lambda

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
)

//go:embed templates/*
var templateFS embed.FS

func AddConfig(cfg *config.Config, region string, runtime string) error {

	cfg.AWSLambda = &config.AWSLambdaConfig{
		Region:   region,
		RoleName: fmt.Sprintf("%s-role", cfg.Name),
		Runtime:  runtime,
	}
	return nil
}

func GenerateLambdaHandler(cfg *config.Config) error {

	var (
		handlerContent string
		handlerPath    string
	)

	if cfg.Framework != "" && cfg.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	if cfg.Framework != "" && cfg.AppVar == "" {
		return fmt.Errorf("app variable is not specified in the configuration")
	}

	if cfg.Framework == config.Express {
		content, err := templateFS.ReadFile("templates/express_lambda_handler_template.js")
		if err != nil {
			return fmt.Errorf("failed to read Express template file: %w", err)
		}

		moduleName := strings.TrimSuffix(filepath.Base(cfg.Entrypoint), filepath.Ext(cfg.Entrypoint))
		handlerContent = strings.ReplaceAll(string(content), "{MODULE_NAME}", moduleName)
		handlerContent = strings.ReplaceAll(handlerContent, "{APP_VAR}", cfg.AppVar)
		handlerPath = filepath.Join(".", "lambda_handler.js")
	} else if cfg.Framework == config.Flask {
		content, err := templateFS.ReadFile("templates/flask_lambda_handler_template.py")
		if err != nil {
			return fmt.Errorf("failed to read Flask template file: %w", err)
		}

		moduleName := strings.TrimSuffix(filepath.Base(cfg.Entrypoint), filepath.Ext(cfg.Entrypoint))
		handlerContent = strings.ReplaceAll(string(content), "{MODULE_NAME}", moduleName)
		handlerContent = strings.ReplaceAll(handlerContent, "{APP_VAR}", cfg.AppVar)
		handlerPath = filepath.Join(".", "lambda_handler.py")
	} else {
		switch cfg.Language {
		case config.Python:
			content, err := templateFS.ReadFile("templates/flask_lambda_handler_template.py")
			if err != nil {
				return fmt.Errorf("failed to read Flask template file: %w", err)
			}

			handlerContent = string(content)
			handlerPath = filepath.Join(".", "lambda_handler.py")
		case config.JavaScript, config.TypeScript:
			content, err := templateFS.ReadFile("templates/node_lambda_handler_template.js")
			if err != nil {
				return fmt.Errorf("failed to read Node.js template file: %w", err)
			}
			handlerContent = string(content)
			handlerPath = filepath.Join(".", "lambda_handler.js")
		}
	}

	fmt.Println("Saving AWS Lambda handler to:", handlerPath)
	if err := os.WriteFile(handlerPath, []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to write lambda handler file: %w", err)
	}

	fmt.Printf("Generated lambda handler at %s\n", handlerPath)
	return nil
}
