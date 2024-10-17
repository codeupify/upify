package lambda

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/lang"
	template "github.com/codeupify/upify/internal/template"
)

const pythonCode = `if os.getenv("UPIFY_DEPLOY_PLATFORM") == "aws-lambda":
    from apig_wsgi import make_lambda_handler
    handler = make_lambda_handler({APP_VAR})`

const nodeCode = `if (process.env.UPIFY_DEPLOY_PLATFORM === 'aws-lambda') {
    const serverless = require('serverless-http');
    let expressApp = {APP_VAR};
    if ({APP_VAR} && {APP_VAR}['app']) {
        expressApp = {APP_VAR}['app'];
      }
      module.exports.handler = serverless(expressApp);
}`

func AddConfig(cfg *config.Config, region string, runtime string) error {

	cfg.AWSLambda = &config.AWSLambdaConfig{
		Region:   region,
		RoleName: fmt.Sprintf("%s-role", cfg.Name),
		Runtime:  runtime,
	}
	return nil
}

func AddHandler(cfg *config.Config) error {

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

	extension := ""
	switch cfg.Language {
	case lang.Python:
		extension = ".py"
	case lang.JavaScript, lang.TypeScript:
		extension = ".js"
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	targetPath := filepath.Join(".", "upify_handler"+extension)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("upify_handler file does not exist at %s", targetPath)
	}

	if cfg.Language == lang.Python {
		handlerTemplate = template.PythonHandlerTemplate
		handlerOutputPath = filepath.Join(".", "upify_handler.py")

		if cfg.Framework != "" {
			mainTemplate = template.PythonMainTemplate
			mainOutputPath = filepath.Join(".", "upify_main.py")
		}
	} else if cfg.Language == lang.JavaScript || cfg.Language == lang.TypeScript {
		handlerTemplate = template.NodeHandlerTemplate
		handlerOutputPath = filepath.Join(".", "upify_handler.js")

		if cfg.Framework != "" {
			mainTemplate = template.NodeMainTemplate
			mainOutputPath = filepath.Join(".", "upify_main.js")
		}
	}

	handlerContent := strings.ReplaceAll(handlerTemplate, "{ENTRYPOINT}", filepath.Base(cfg.Entrypoint))
	handlerContent = strings.ReplaceAll(handlerContent, "{APP_VAR}", cfg.AppVar)

	handlerPath := filepath.Join(".", handlerOutputPath)
	_, err := os.Stat(handlerPath)

	if err == "" {
		return fmt.Printf("handler file already exists at %s", handlerPath)
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
