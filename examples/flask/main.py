import os
import json
from flask import Flask

app = Flask(__name__, instance_path='/tmp')


@app.route('/')
def hello_world():
    return 'Hello, World!'

@app.route('/env')
def env():
    return json.dumps(dict(os.environ))

if __name__ == '__main__':
    app.run()