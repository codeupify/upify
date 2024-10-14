import os
from main import app
import functions_framework

handler = None

if os.getenv("UPIFY_DEPLOY_PLATFORM") == "gcp-cloudrun":
    @functions_framework.http
    def flask_function(request):
        with app.request_context(request.environ):
            return app.full_dispatch_request()

    handler = flask_function

# elif os.getenv("DEPLOY_PLATFORM") == "aws-lambda":
#     from apig_wsgi import make_lambda_handler
    # handler = make_lambda_handler(app)

