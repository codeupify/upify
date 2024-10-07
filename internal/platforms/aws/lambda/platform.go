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

func getPythonTemplate() (string, string, error) {
	content, err := templateFS.ReadFile("templates/flask_lambda_handler_template.py")
	if err != nil {
		return "", "", fmt.Errorf("failed to read Flask template file: %w", err)
	}
	handlerPath := filepath.Join(".", "lambda_handler.py")

	return string(content), handlerPath, nil
}

func getJavascriptTemplate(moduleSystem config.ModuleSystem) (string, string, error) {
	content, err := templateFS.ReadFile(fmt.Sprintf("templates/node_lambda_handler_template_%s.js", string(moduleSystem)))
	if err != nil {
		return "", "", fmt.Errorf("failed to read Node.js template file: %w", err)
	}
	handlerPath := filepath.Join(".", "lambda_handler.js")
	return string(content), handlerPath, nil
}

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
		err            error
	)

	if cfg.Framework == "" {
		switch cfg.Language {
		case config.Python:
			handlerContent, handlerPath, err = getPythonTemplate()
		case config.JavaScript, config.TypeScript:
			handlerContent, handlerPath, err = getJavascriptTemplate(cfg.ModuleSystem)
		default:
			return fmt.Errorf("CLI doesn't support the specified language yet for AWS Lambda: %s", cfg.Language)
		}

		if err != nil {
			return err
		}
	} else {
		if cfg.Entrypoint == "" {
			return fmt.Errorf("entrypoint is not specified in the configuration")
		}
		if cfg.AppVar == "" {
			return fmt.Errorf("app variable is not specified in the configuration")
		}

		moduleName := strings.TrimSuffix(filepath.Base(cfg.Entrypoint), filepath.Ext(cfg.Entrypoint))

		if cfg.Framework == "flask" {

			content, err := templateFS.ReadFile("templates/flask_lambda_handler_template.py")
			if err != nil {
				return fmt.Errorf("failed to read Flask template file: %w", err)
			}

			handlerContent = strings.ReplaceAll(string(content), "{MODULE_NAME}", moduleName)
			handlerContent = strings.ReplaceAll(handlerContent, "{APP_VAR}", cfg.AppVar)
			handlerPath = filepath.Join(".", "lambda_handler.py")
		} else if cfg.Framework == "express" {

			content, err := templateFS.ReadFile("templates/express_lambda_handler_template.js")
			if err != nil {
				return fmt.Errorf("failed to read Express template file: %w", err)
			}

			handlerContent = strings.ReplaceAll(string(content), "{MODULE_NAME}", moduleName)
			handlerContent = strings.ReplaceAll(handlerContent, "{APP_VAR}", cfg.AppVar)
			handlerPath = filepath.Join(".", "lambda_handler.js")
		} else {
			return fmt.Errorf("CLI doesn't support the specified framework yet for AWS Lambda: %s", cfg.Framework)
		}
	}

	if err := os.WriteFile(handlerPath, []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to write lambda handler file: %w", err)
	}

	fmt.Printf("Generated lambda handler at %s\n", handlerPath)
	return nil
}
