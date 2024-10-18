package lambda

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/framework"
	"github.com/codeupify/upify/internal/handler"
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

	if cfg.Framework != "" && cfg.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	if cfg.Framework != "" && cfg.AppVar == "" {
		return fmt.Errorf("app variable is not specified in the configuration")
	}

	targetPath := handler.GetHandlerPath(cfg.Language)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("upify_handler file does not exist at %s", targetPath)
	}

	var handlerCode string
	switch cfg.Language {
	case lang.Python:
		handlerCode = pythonCode
	case lang.JavaScript, lang.TypeScript:
		handlerCode = nodeCode
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	handlerCode = strings.ReplaceAll(handlerCode, "{APP_VAR}", cfg.AppVar)
	err := handler.AddHandlerSection(targetPath, "aws-lambda", handlerCode)
	if err != nil {
		return err
	}

	var mainCode string
	switch cfg.Framework {
	case framework.Flask:
		mainCode = template.PythonMainTemplate
	case framework.Express:
		mainCode = template.NodeMainTemplate
	default:
		mainCode = ""
	}

	mainPath := handler.GetMainPath(cfg.Language)
	_, err = os.Stat(mainPath)
	if os.IsNotExist(err) {
		fmt.Println("Saving upify_main to:", mainPath)
		if err := os.WriteFile(mainPath, []byte(mainCode), 0644); err != nil {
			return err
		}
	} else if err == nil {
		fmt.Printf("upify_main already exists at %s", mainPath)
	} else {
		return err
	}

	return nil
}
