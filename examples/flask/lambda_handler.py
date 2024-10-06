from apig_wsgi import make_lambda_handler
from main import flask_app

handler = make_lambda_handler(flask_app)