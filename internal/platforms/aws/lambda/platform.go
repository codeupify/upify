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

func AddConfig(cfg *config.Config) error {
	if cfg.AWSLambda != nil {
		return fmt.Errorf("aws-lambda configuration already exists")
	}

	runtime := determineRuntime(cfg.Language)
	if runtime == "" {
		return fmt.Errorf("unable to determine appropriate runtime for language: %s and framework: %s", cfg.Language, cfg.Framework)
	}

	cfg.AWSLambda = &config.AWSLambdaConfig{
		Region:   "us-east-1",
		RoleName: fmt.Sprintf("%s-role", cfg.Name),
		Runtime:  runtime,
	}
	return nil
}

func determineRuntime(language string) string {
	switch language {
	case "python":
		return "python3.12"
	case "javascript":
		return "nodejs20.x"
	case "typescript":
		return "nodejs20.x"
	}

	return ""
}

func GenerateLambdaHandler(cfg *config.Config) error {
	if cfg.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	upifyDir := ".upify"
	if err := os.MkdirAll(upifyDir, 0755); err != nil {
		return fmt.Errorf("failed to create .upify directory: %w", err)
	}

	moduleName := strings.TrimSuffix(filepath.Base(cfg.Entrypoint), filepath.Ext(cfg.Entrypoint))

	var (
		templateName string
		handlerExt   string
	)

	switch cfg.Framework {
	case "flask":
		templateName = "flask_lambda_handler_template.py"
		handlerExt = ".py"
	case "express":
		templateName = "express_lambda_handler_template.js"
		handlerExt = ".js"
	case "none":
		if cfg.Language == "python" {
			templateName = "python_lambda_handler_template.py"
			handlerExt = ".py"
		} else if cfg.Language == "javascript" || cfg.Language == "typescript" {
			templateName = "javascript_lambda_handler_template.js"
			handlerExt = ".js"
		} else {
			return fmt.Errorf("unsupported language for 'none' framework: %s", cfg.Language)
		}
	default:
		return fmt.Errorf("unsupported framework for lambda handler generation: %s", cfg.Framework)
	}

	content, err := templateFS.ReadFile(filepath.Join("templates", templateName))
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	handlerContent := strings.ReplaceAll(string(content), "{MODULE_NAME}", moduleName)

	handlerPath := filepath.Join(upifyDir, "lambda_handler"+handlerExt)

	if err := os.WriteFile(handlerPath, []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to write lambda handler file: %w", err)
	}

	fmt.Printf("Generated lambda handler at %s\n", handlerPath)
	return nil
}
