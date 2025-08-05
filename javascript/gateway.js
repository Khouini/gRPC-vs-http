const express = require('express');

const app = express();
const PORT = 3000;

// Gateway endpoint - calls microservice and processes hotel data
app.get('/stats', async (req, res) => {
    const startTime = Date.now();

    try {
        // Call microservice
        const response = await fetch('http://localhost:3001/data');
        const data = await response.json();

        // Count hotels and calculate stats
        const hotelCount = data.hotels ? data.hotels.length : 0;
        const availableHotels = data.hotels ? data.hotels.filter(hotel => hotel.available).length : 0;

        const processTime = Date.now() - startTime;

        // Return processed stats only
        res.json({
            processTimeMs: processTime,
            totalHotels: hotelCount,
            availableHotels: availableHotels,
            dataSize: data.metadata ? data.metadata.actualSizeMB : 0,
        });
    } catch (error) {
        res.status(500).json({ error: 'Failed to fetch data from microservice' });
    }
});

app.listen(PORT, () => {
    console.log(`Hotel Gateway running on port ${PORT}`);
    console.log(`Try: http://localhost:${PORT}/stats`);
});