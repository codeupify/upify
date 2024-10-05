from apig_wsgi import make_lambda_handler
import {MODULE_NAME}

# Modify '{MODULE_NAME}.flask_app' to match your Flask app
handler = make_lambda_handler({MODULE_NAME}.flask_app)