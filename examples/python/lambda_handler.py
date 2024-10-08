from main import get_weather_data
import json

"""
Event payload format is API Gateway format 2
https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html#http-api-develop-integrations-lambda.proxy-format
"""

def handler(event, context):

    event_query_params = event.get("queryStringParameters")
    city = event_query_params.get("city") if event_query_params else None
    if event_query_params is None or city is None:
        return {
            "statusCode": 400,
            "body": json.dumps({"error": "Specify a city in the query string"})
        };

    print("Got a weather request for " + city)
    try:
        response_data = get_weather_data(city)
        return {
            "statusCode": 200,
            "body": json.dumps(response_data)
        };
    except Exception as e:
        return {
            "statusCode": 500,
            "body": json.dumps({"error": f"Failed to get weather data: {repr(e)}"})
        };