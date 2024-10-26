package cloudrun

import (
	"fmt"

	"github.com/codeupify/upify/internal/config"
	"github.com/codeupify/upify/internal/handler"
	"github.com/codeupify/upify/internal/lang"
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

func AddConfig(cfg *config.Config, region string, projectID string, runtime string) error {
	cfg.GCPCloudRun = &config.GCPCloudRunConfig{
		Region:    region,
		ProjectId: projectID,
		Runtime:   runtime,
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

	return handler.AddPlatformHandler(cfg, "gcp-cloudrun", handlerCode)
}
