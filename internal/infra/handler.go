package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/lang/node"
	"github.com/codeupify/upify/internal/lang/python"
)

const pythonHandlerCode = `import os
from {ENTRYPOINT} import app

handler = None
`

const nodeHandlerCode = `const app = require('./{ENTRYPOINT}');`

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

	updatedContent := string(content) + "\n\n" + sectionContent

	fmt.Printf("Adding handler section to %s...\n", handlerPath)
	err = os.WriteFile(handlerPath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func AddPlatformHandler(cfg *config.Config, platform string, handlerCode string) error {

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
	err := AddHandlerSection(targetPath, platform, handlerCode)
	if err != nil {
		return err
	}

	return nil
}

func AddHandlerFile(cfg *config.Config) error {
	var handlerCode string

	switch cfg.Language {
	case lang.Python:
		handlerCode = pythonHandlerCode
	case lang.JavaScript, lang.TypeScript:
		handlerCode = nodeHandlerCode
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	targetPath := GetHandlerPath(cfg.Language)
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("Handler file already exists at %s\n", targetPath)
		return nil
	}

	entrypoint := cfg.Entrypoint
	if entrypoint == "" {
		entrypoint = "upify_main"
	} else {
		entrypoint = strings.TrimSuffix(entrypoint, filepath.Ext(entrypoint))
	}

	if cfg.Language == lang.Python {
		entrypoint = strings.ReplaceAll(entrypoint, "/", ".")
	}

	fmt.Printf("Adding handler file at %s...\n", targetPath)
	handlerCode = strings.ReplaceAll(handlerCode, "{ENTRYPOINT}", entrypoint)
	err := os.WriteFile(targetPath, []byte(handlerCode), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func AddMainFile(cfg *config.Config) error {
	var mainCode string
	switch cfg.Language {
	case lang.Python:
		mainCode = python.PythonMainTemplate
	case lang.JavaScript, lang.TypeScript:
		mainCode = node.NodeMainTemplate
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	mainPath := GetMainPath(cfg.Language)
	if _, err := os.Stat(mainPath); err == nil {
		fmt.Printf("Main file already exists at %s\n", mainPath)
	} else {

		fmt.Printf("Adding a main file at %s...\n", mainPath)
		err = os.WriteFile(mainPath, []byte(mainCode), 0644)
		if err != nil {
			return fmt.Errorf("error writing file: %v", err)
		}
	}

	return nil
}
