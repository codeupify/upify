import requests
import os

YOUR_OPENWEATHERMAP_API_KEY = os.getenv("YOUR_OPENWEATHERMAP_API_KEY")

def get_weather_data(city):
    base_url = "http://api.openweathermap.org/data/2.5/weather"

    params = {
        "q": city,
        "appid": YOUR_OPENWEATHERMAP_API_KEY,
        "units": "metric"
    }

    response = requests.get(base_url, params=params)
    response.raise_for_status()  # Raises an HTTPError for bad responses
    weather_data = response.json()

    return weather_data

if __name__ == "__main__":

    if len(sys.argv) < 2:
        print("Usage: python main.py <city>")
        sys.exit(1)

    city = sys.argv[1]
    weather_data = get_weather_data(city)
    print(weather_data)