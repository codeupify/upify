package aws

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/infra"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/platform"
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

//go:embed templates/main.tmpl
var MainTemplate string

//go:embed templates/main.module.tmpl
var MainModuleTemplate string

func AddPlatform(cfg *config.Config, region string, runtime string) error {
	fmt.Println("Adding AWS handlers...")

	var handlerCode string
	switch cfg.Language {
	case lang.Python:
		handlerCode = pythonCode
	case lang.JavaScript, lang.TypeScript:
		handlerCode = nodeCode
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	err := infra.AddPlatformHandler(cfg, "aws-lambda", handlerCode)
	if err != nil {
		return err
	}

	fmt.Println("Setting up AWS Lambda infrastructure...")

	mainContent := MainTemplate
	mainContent = strings.Replace(mainContent, "{LAMBDA_NAME}", cfg.Name, -1)
	mainContent = strings.Replace(mainContent, "{REGION}", region, -1)
	mainContent = strings.Replace(mainContent, "{RUNTIME}", runtime, -1)

	return infra.AddPlatform(platform.AWS, mainContent, MainModuleTemplate)
}
