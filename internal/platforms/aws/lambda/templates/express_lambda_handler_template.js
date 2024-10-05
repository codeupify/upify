const serverless = require('serverless-http');

const express_app = require('./{MODULE_NAME}');

module.exports.handler = serverless(express_app);