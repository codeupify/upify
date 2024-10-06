// Note: this template is for CommonJS modules, for ES modules use this template: TODO
const importedModule = require('./index');

/**
 * Event payload format is API Gateway format 2
 * @see https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html#http-api-develop-integrations-lambda.proxy-format
 */

exports.handler = async (event, context) => {
    if (!event.queryStringParameters) {
        console.log("Specify a city in the query string");
        return {
            statusCode: 400,
            body: JSON.stringify({ error: "Specify a city in the query string" })
        };
    }

    const city = event.queryStringParameters.city;
    console.log("Got a weather request for " + city);
    
    try {
        const responseData = await main.getWeatherData(city);
        return {
            statusCode: 200,
            body: JSON.stringify(responseData)
        };
    } catch (error) {
        console.error("Error getting weather data:", error);
        return {
            statusCode: 500,
            body: JSON.stringify({ error: "Failed to get weather data" })
        };
    }
};