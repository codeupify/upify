import axios from 'axios';

const YOUR_OPENWEATHERMAP_API_KEY = process.env.YOUR_OPENWEATHERMAP_API_KEY;

interface WeatherData {
    main: {
        temp: number;
        feels_like: number;
        humidity: number;
    };
    weather: Array<{
        description: string;
        icon: string;
    }>;
    name: string;
}

export async function getWeatherData(city: string): Promise<WeatherData> {
    const baseUrl = "http://api.openweathermap.org/data/2.5/weather";

    const params = {
        q: city,
        appid: YOUR_OPENWEATHERMAP_API_KEY,
        units: "metric"
    };

    try {
        const response = await axios.get<WeatherData>(baseUrl, { params });
        return response.data;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            throw new Error(`Failed to fetch weather data: ${error.message}`);
        } else {
            throw error;
        }
    }
}

async function main() {
    if (process.argv.length < 3) {
        console.log("Usage: ts-node main.ts <city>");
        process.exit(1);
    }

    const city = process.argv[2];
    try {
        const weatherData = await getWeatherData(city);
        console.log(weatherData);
    } catch (error) {
        console.error(error);
    }
}
