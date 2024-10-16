const express = require('express');
const { getWeatherData } = require('./index');

const app = express();

app.get('/', async (req, res) => {
    if (!req.query.city) {
        console.log("Specify a city in the query string");
        return res.status(400).json({ error: "Specify a city in the query string" });
    }

    const city = req.query.city;
    console.log("Got a weather request for " + city);
    
    try {
        const responseData = await getWeatherData(city);
        res.status(200).json(responseData);
    } catch (error) {
        console.error("Error getting weather data:", error);
        res.status(500).json({ error: error.message });
    }
});

module.exports = app;