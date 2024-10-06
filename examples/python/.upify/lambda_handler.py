import main

"""
Event payload format is API Gateway format 2
https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html#http-api-develop-integrations-lambda.proxy-format
"""

def handler(event, context):
    if event["queryStringParameters"] is None:
        print("Specify a city in the query string")
        return

    city = event["queryStringParameters"]["city"]
    print("Got a weather request for " + city)
    response_data = main.get_weather_data(event["queryStringParameters"]["city"])
    return response_data