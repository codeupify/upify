const app = require('./upify_main');

if (process.env.UPIFY_DEPLOY_PLATFORM === 'aws-lambda') {
    const serverless = require('serverless-http');
    let expressApp = app;
    if (app && app['app']) {
        expressApp = app['app'];
      }
      module.exports.handler = serverless(expressApp);
}

if (process.env.UPIFY_DEPLOY_PLATFORM === 'gcp-cloudrun') {
    const functions = require('@google-cloud/functions-framework');
    functions.http('handler', (req, res) => {
        app(req, res);
    });
}