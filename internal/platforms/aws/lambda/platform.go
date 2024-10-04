package lambda

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeupify/upify/internal/config"
)

const (
	FlaskLambdaHandlerTemplate = `from apig_wsgi import make_lambda_handler
import %s

# Modify '%s.flask_app' to match your Flask app
handler = make_lambda_handler(%s.flask_app)
`

	NoneLambdaHandlerTemplate = `import %s

""" Event payload format is API Gateway format 2 
https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html#http-api-develop-integrations-lambda.proxy-format """

def handler(event, context):
    """ Wrap your python code to make it handle web requests and return web responses. 
    Look at the example code at https://.... """
    pass
`
	ExpressLambdaHandlerTemplate = `const serverless = require('serverless-http');

const express_app = require('./%s');

module.exports.handler = serverless(express_app);
`
)

func AddConfig(cfg *config.Config) error {
	if cfg.AWSLambda != nil {
		return fmt.Errorf("aws-lambda configuration already exists")
	}

	runtime := determineRuntime(cfg.Language)
	if runtime == "" {
		return fmt.Errorf("unable to determine appropriate runtime for language: %s and framework: %s", cfg.Language, cfg.Framework)
	}

	cfg.AWSLambda = &config.AWSLambdaConfig{
		Region:   "us-east-1",
		RoleName: fmt.Sprintf("%s-role", cfg.Name),
		Runtime:  runtime,
	}
	return nil
}

func determineRuntime(language string) string {
	switch language {
	case "python":
		return "python3.12"
	case "javascript":
		return "nodejs20.x"
	case "typescript":
		return "nodejs20.x"
	}

	return ""
}

func GenerateLambdaHandler(cfg *config.Config) error {
	if cfg.Entrypoint == "" {
		return fmt.Errorf("entrypoint is not specified in the configuration")
	}

	upifyDir := ".upify"
	if err := os.MkdirAll(upifyDir, 0755); err != nil {
		return fmt.Errorf("failed to create .upify directory: %w", err)
	}

	moduleName := strings.TrimSuffix(filepath.Base(cfg.Entrypoint), filepath.Ext(cfg.Entrypoint))

	var (
		content    string
		handlerExt string
	)

	switch cfg.Framework {
	case "flask":
		content = fmt.Sprintf(FlaskLambdaHandlerTemplate, moduleName, moduleName, moduleName)
		handlerExt = ".py"
	case "express":
		content = fmt.Sprintf(ExpressLambdaHandlerTemplate, moduleName)
		handlerExt = ".js"
	case "none":
		if cfg.Language == "python" {
			content = fmt.Sprintf(NoneLambdaHandlerTemplate, moduleName)
			handlerExt = ".py"
		} else if cfg.Language == "javascript" {
			content = fmt.Sprintf(ExpressLambdaHandlerTemplate, moduleName)
			handlerExt = ".js"
		} else {
			return fmt.Errorf("unsupported language for 'none' framework: %s", cfg.Language)
		}
	default:
		return fmt.Errorf("unsupported framework for lambda handler generation: %s", cfg.Framework)
	}

	handlerPath := filepath.Join(upifyDir, "lambda_handler"+handlerExt)

	if err := os.WriteFile(handlerPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write lambda handler file: %w", err)
	}

	fmt.Printf("Generated lambda handler at %s\n", handlerPath)
	return nil
}
