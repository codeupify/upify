package lambda

import (
	_ "embed"
	"fmt"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/handler"
	"github.com/codeupify/upify/internal/lang"
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

	var handlerCode string
	switch cfg.Language {
	case lang.Python:
		handlerCode = pythonCode
	case lang.JavaScript, lang.TypeScript:
		handlerCode = nodeCode
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	return handler.AddPlatformHandler(cfg, "aws-lambda", handlerCode)
}
