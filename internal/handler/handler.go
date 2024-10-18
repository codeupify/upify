package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/lang"
)

func GetHandlerFileName(language lang.Language) string {
	switch language {
	case lang.Python:
		return "upify_handler.py"
	case lang.JavaScript, lang.TypeScript:
		return "upify_handler.js"
	default:
		return ""
	}
}

func GetMainFileName(language lang.Language) string {
	switch language {
	case lang.Python:
		return "upify_main.py"
	case lang.JavaScript, lang.TypeScript:
		return "upify_main.js"
	default:
		return ""
	}
}

func GetHandlerPath(language lang.Language) string {
	return filepath.Join(".", GetHandlerFileName(language))
}

func GetMainPath(language lang.Language) string {
	return filepath.Join(".", GetMainFileName(language))
}

func AddHandlerSection(handlerPath string, sectionName string, sectionContent string) error {

	content, err := os.ReadFile(handlerPath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	if strings.Contains(string(content), sectionName) {
		fmt.Printf("Handler section already exists in %s\n", handlerPath)
		return nil
	}

	updatedContent := string(content) + "\n" + sectionContent
	err = os.WriteFile(handlerPath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func AddHandler(cfg *config.Config, platform string, handlerCode string) error {

	if cfg.Framework != "" && cfg.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	if cfg.Framework != "" && cfg.AppVar == "" {
		return fmt.Errorf("app variable is not specified in the configuration")
	}

	targetPath := GetHandlerPath(cfg.Language)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("upify_handler file does not exist at %s", targetPath)
	}

	appVar := "app"
	if cfg.AppVar != "" {
		appVar = cfg.AppVar
	}

	handlerCode = strings.ReplaceAll(handlerCode, "{APP_VAR}", appVar)
	err := AddHandlerSection(targetPath, "aws-lambda", handlerCode)
	if err != nil {
		return err
	}

	return nil
}
