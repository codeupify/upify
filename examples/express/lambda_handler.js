const serverless = require('serverless-http');
const importedModule = require('./index');

let expressApp = importedModule;

if (importedModule && importedModule['app']) {
  expressApp = importedModule['app'];
}

module.exports.handler = serverless(expressApp);