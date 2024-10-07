const serverless = require('serverless-http');
const importedModule = require('./{MODULE_NAME}');

let expressApp = importedModule;

if (importedModule && importedModule['{APP_VAR}']) {
  expressApp = importedModule['{APP_VAR}'];
}

module.exports.handler = serverless(expressApp);