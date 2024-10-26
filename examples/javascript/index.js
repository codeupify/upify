const axios = require('axios');

const YOUR_OPENWEATHERMAP_API_KEY = process.env.YOUR_OPENWEATHERMAP_API_KEY;

async function getWeatherData(city) {
    const baseUrl = "http://api.openweathermap.org/data/2.5/weather";

    const params = {
        q: city,
        appid: YOUR_OPENWEATHERMAP_API_KEY,
        units: "metric"
    };

    try {
        const response = await axios.get(baseUrl, { params });
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
        console.log("Usage: node main.js <city>");
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

module.exports = {
    getWeatherData
};