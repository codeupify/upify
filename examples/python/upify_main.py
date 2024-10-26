from flask import Flask, request, jsonify
from main import get_weather_data

app = Flask(__name__, instance_path='/tmp')

@app.route('/')
def handler():
    city = request.args.get("city")
    if city is None:
        return jsonify({"error": "Specify a city in the query string"}), 400

    try:
        response_data = get_weather_data(city)
        return jsonify(response_data), 200
    except Exception as e:
        return jsonify({"error": f"Failed to get weather data: {repr(e)}"}), 500