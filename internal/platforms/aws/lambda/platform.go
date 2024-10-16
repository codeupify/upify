package lambda

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
	template "github.com/codeupify/upify/internal/templates"
)

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
		handlerTemplate   string
		handlerOutputPath string
		mainTemplate      string
		mainOutputPath    string
	)

	if cfg.Framework != "" && cfg.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	if cfg.Framework != "" && cfg.AppVar == "" {
		return fmt.Errorf("app variable is not specified in the configuration")
	}

	if cfg.Language == config.Python {
		handlerTemplate = template.PythonHandlerTemplate
		handlerOutputPath = filepath.Join(".", "upify_handler.py")

		if cfg.Framework != "" {
			mainTemplate = template.PythonMainTemplate
			mainOutputPath = filepath.Join(".", "upify_main.py")
		}
	} else if cfg.Language == config.JavaScript || cfg.Language == config.TypeScript {
		handlerTemplate = template.NodeHandlerTemplate
		handlerOutputPath = filepath.Join(".", "upify_handler.js")

		if cfg.Framework != "" {
			mainTemplate = template.NodeMainTemplate
			mainOutputPath = filepath.Join(".", "upify_main.js")
		}
	}

	handlerContent := strings.ReplaceAll(handlerTemplate, "{ENTRYPOINT}", filepath.Base(cfg.Entrypoint))
	handlerContent = strings.ReplaceAll(handlerContent, "{ENTRYPOINT}", filepath.Base(cfg.Entrypoint))

	handlerPath := filepath.Join(".", handlerOutputPath)
	if _, err := os.Stat(handlerPath); err == nil {
		return fmt.Errorf("handler file already exists at %s", handlerPath)
	}

	fmt.Println("Saving upify_handler to:", handlerPath)
	if err := os.WriteFile(handlerPath, []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to write lambda handler file: %w", err)
	}

	if mainOutputPath != "" {
		if _, err := os.Stat(mainOutputPath); err == nil {
			return fmt.Errorf("main file already exists at %s", mainOutputPath)
		}

		fmt.Println("Saving upify_main to:", mainOutputPath)
		err := os.WriteFile(mainOutputPath, []byte(mainTemplate), 0644)
		if err != nil {
			return fmt.Errorf("failed to write lambda main file: %w", err)
		}

		fmt.Printf("\nAdd your code to %s to wrap your script with request/response\n", mainOutputPath)
	}

	return nil
}
