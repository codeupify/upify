package lambda

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/codeupify/upify/internal/config"
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

	var (
		handlerCode string
		mainCode    string
	)

	switch cfg.Language {
	case lang.Python:
		handlerCode = pythonCode
		mainCode = template.PythonMainTemplate
	case lang.JavaScript, lang.TypeScript:
		handlerCode = nodeCode
		mainCode = template.NodeMainTemplate
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	appVar := "app"
	if cfg.AppVar != "" {
		appVar = cfg.AppVar
	}

	handlerCode = strings.ReplaceAll(handlerCode, "{APP_VAR}", appVar)
	err := handler.AddHandlerSection(targetPath, "aws-lambda", handlerCode)
	if err != nil {
		return err
	}

	return nil
}
