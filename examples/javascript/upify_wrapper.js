const functions = require('@google-cloud/functions-framework');
const app = require('./request_wrapper');

if (process.env.UPIFY_DEPLOY_PLATFORM === 'gcp-cloudrun') {
    functions.http('handler', (req, res) => {
        app(req, res);
    });
}