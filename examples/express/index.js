const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.send('Hello, World!');
});

app.get('/env', (req, res) => {
  const envVars = Object.keys(process.env).map(key => ({ [key]: process.env[key] }));
  res.send(JSON.stringify(envVars));
});

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});

module.exports = app