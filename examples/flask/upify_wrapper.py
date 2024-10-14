import os
from main import app
import functions_framework


# handler = None

# if os.getenv("DEPLOY_PLATFORM") == "gcp-cloudrun":
#     handler = app

# elif os.getenv("DEPLOY_PLATFORM") == "aws-lambda":
#     from apig_wsgi import make_lambda_handler
    # handler = make_lambda_handler(app)
@functions_framework.http
def handler(request):
    return app(request.environ, lambda x, y: None)
