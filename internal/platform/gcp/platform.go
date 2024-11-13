package gcp

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/infra"
	"github.com/codeupify/upify/internal/lang"
	"github.com/codeupify/upify/internal/platform"
)

const pythonCode = `if os.getenv("UPIFY_DEPLOY_PLATFORM") == "gcp-cloudrun":
    import functions_framework

    @functions_framework.http
    def flask_function(request):
        with app.request_context(request.environ):
            return app.full_dispatch_request()

    handler = flask_function`

const nodeCode = `if (process.env.UPIFY_DEPLOY_PLATFORM === 'gcp-cloudrun') {
    const functions = require('@google-cloud/functions-framework');
    functions.http('handler', (req, res) => {
        app(req, res);
    });
}`

//go:embed templates/main.tmpl
var MainTemplate string

//go:embed templates/main.module.tmpl
var MainModuleTemplate string

func AddPlatform(cfg *config.Config, region string, runtime string, projectId string) error {
	fmt.Println("Adding GCP handlers...")

	var handlerCode string
	switch cfg.Language {
	case lang.Python:
		handlerCode = pythonCode
	case lang.JavaScript, lang.TypeScript:
		handlerCode = nodeCode
	default:
		return fmt.Errorf("unsupported language: %s", cfg.Language)
	}

	err := infra.AddPlatformHandler(cfg, "gcp-cloudrun", handlerCode)
	if err != nil {
		return err
	}

	fmt.Println("Setting up GCP Cloud Run infrastructure...")

	mainContent := MainTemplate
	mainContent = strings.Replace(mainContent, "{FUNCTION_NAME}", cfg.Name, -1)
	mainContent = strings.Replace(mainContent, "{REGION}", region, -1)
	mainContent = strings.Replace(mainContent, "{RUNTIME}", runtime, -1)
	mainContent = strings.Replace(mainContent, "{PROJECT_ID}", projectId, -1)

	return infra.AddPlatform(platform.GCP, mainContent, MainModuleTemplate)
}
