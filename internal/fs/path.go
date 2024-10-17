package fs

import (
	"path/filepath"

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

func GetHandlerPath(language lang.Language) string {
	return filepath.Join(".", GetHandlerFileName(language))
}
