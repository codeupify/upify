package cloudrun

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
)

var templateFS embed.FS

func AddConfig(cfg *config.Config, region string, projectID string) error {
	cfg.GCPCloudRun = &config.GCPCloudRunConfig{
		Region:    region,
		ProjectId: projectID,
	}
	return nil
}

func GenerateFiles(cfg *config.Config) error {
	if cfg.Framework != "" && cfg.Entrypoint == "" {
		return fmt.Errorf("framework or entrypoint must be specified in the configuration")
	}

	if cfg.Framework != "" && cfg.AppVar == "" {
		return fmt.Errorf("app variable is not specified in the configuration")
	}

	var dockerfileContent string

	// Choose a base Dockerfile template based on the language/framework
	switch cfg.Language {
	case config.Python:
		content, err := templateFS.ReadFile("templates/python_flask_dockerfile.template")

		if cfg.Framework == config.Flask {

			if err != nil {
				return fmt.Errorf("failed to read Python Flask Dockerfile template: %w", err)
			}
			dockerfileContent = string(content)
		} else {
			content, err := templateFS.ReadFile("templates/python_dockerfile.template")
			if err != nil {
				return fmt.Errorf("failed to read Python Dockerfile template: %w", err)
			}
			dockerfileContent = string(content)
		}
	case config.JavaScript, config.TypeScript:
		if cfg.Framework == config.Express {
			content, err := templateFS.ReadFile("templates/node_express_dockerfile.template")
			if err != nil {
				return fmt.Errorf("failed to read Node.js Express Dockerfile template: %w", err)
			}
			dockerfileContent = string(content)
		} else {
			content, err := templateFS.ReadFile("templates/node_dockerfile.template")
			if err != nil {
				return fmt.Errorf("failed to read Node.js Dockerfile template: %w", err)
			}
			dockerfileContent = string(content)
		}
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	dockerfileContent = strings.ReplaceAll(dockerfileContent, "{PORT}", fmt.Sprintf("%d", 8080))
	dockerfileContent = strings.ReplaceAll(dockerfileContent, "{ENTRYPOINT}", cfg.Entrypoint)

	dockerfilePath := filepath.Join(".upify", "Dockerfile.cloudrun")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("Generated Dockerfile at %s\n", dockerfilePath)
	return nil
}
