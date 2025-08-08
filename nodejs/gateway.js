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

// Concurrent stats endpoint - makes multiple concurrent calls to microservice
app.get('/concurrent-stats', async (req, res) => {
    const startTime = Date.now();

    // Get number of concurrent calls from query parameter, default to 10
    const concurrentCalls = Math.min(parseInt(req.query.calls) || 10, 100); // Limit to 100 max

    console.log(`Processing ${concurrentCalls} concurrent stats calls`);

    try {
        // Create array of promises for concurrent calls
        const promises = Array(concurrentCalls).fill(null).map(async () => {
            const callStartTime = Date.now();

            try {
                // Call microservice
                const response = await fetch('http://localhost:3001/data');
                const data = await response.json();

                // Count hotels and calculate stats
                const hotelCount = data.hotels ? data.hotels.length : 0;
                const availableHotels = data.hotels ? data.hotels.filter(hotel => hotel.available).length : 0;

                const callProcessTime = Date.now() - callStartTime;

                return {
                    processTimeMs: callProcessTime,
                    totalHotels: hotelCount,
                    availableHotels: availableHotels,
                    dataSize: data.metadata ? data.metadata.actualSizeMB : 0,
                };
            } catch (error) {
                throw new Error(`Call failed: ${error.message}`);
            }
        });

        // Wait for all promises to resolve
        const results = await Promise.allSettled(promises);

        // Separate successful and failed results
        const successfulResults = results
            .filter(result => result.status === 'fulfilled')
            .map(result => result.value);

        const failedResults = results
            .filter(result => result.status === 'rejected')
            .map(result => result.reason.message);

        const totalTime = Date.now() - startTime;

        // Calculate statistics
        let minTime = Number.MAX_SAFE_INTEGER;
        let maxTime = 0;
        let totalProcessTime = 0;

        successfulResults.forEach(result => {
            if (result.processTimeMs < minTime) {
                minTime = result.processTimeMs;
            }
            if (result.processTimeMs > maxTime) {
                maxTime = result.processTimeMs;
            }
            totalProcessTime += result.processTimeMs;
        });

        const averageTime = successfulResults.length > 0 ? totalProcessTime / successfulResults.length : 0;

        // Return comprehensive stats
        res.json({
            totalTimeMs: totalTime,
            concurrentCalls: concurrentCalls,
            successfulCalls: successfulResults.length,
            failedCalls: failedResults.length,
            averageTimeMs: averageTime,
            minTimeMs: successfulResults.length > 0 ? minTime : 0,
            maxTimeMs: maxTime,
            results: successfulResults,
            errors: failedResults
        });
    } catch (error) {
        res.status(500).json({ error: 'Failed to process concurrent requests' });
    }
});

app.listen(PORT, () => {
    console.log(`Hotel Gateway running on port ${PORT}`);
    console.log(`Endpoints:`);
    console.log(`  - GET /stats (hotel statistics)`);
    console.log(`  - GET /concurrent-stats?calls=<num> (concurrent hotel statistics, default: 10 calls)`);
});