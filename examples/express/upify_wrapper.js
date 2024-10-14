const functions = require('@google-cloud/functions-framework');
const app = require('./index');

// Ignore App.listen() in serverless environments
// express.application.listen = function() {};

if (process.env.UPIFY_DEPLOY_PLATFORM === 'gcp-cloudrun') {
    functions.http('handler', (req, res) => {
        app(req, res);
    });
}