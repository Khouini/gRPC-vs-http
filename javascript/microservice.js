const express = require('express');
const fs = require('fs');
const path = require('path');

// Load hotel data once at startup
const dataPath = path.join(__dirname, '..', 'data.json');
const DATA = JSON.parse(fs.readFileSync(dataPath, 'utf8'));

const app = express();
const PORT = 3001;

// Simple endpoint to return hotel data
app.get('/data', (req, res) => {
    res.json(DATA);
});

app.listen(PORT, () => {
    console.log(`Hotel Microservice running on port ${PORT}`);
    console.log(`Loaded ${DATA.hotels ? DATA.hotels.length : 0} hotels`);
});