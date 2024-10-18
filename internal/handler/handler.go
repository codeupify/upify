package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
