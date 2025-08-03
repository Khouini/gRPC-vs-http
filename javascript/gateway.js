const express = require('express');

const app = express();
const PORT = 3000;

// Gateway endpoint - calls microservice and processes data
app.get('/stats', async (req, res) => {
    const startTime = Date.now();

    try {
        // Call microservice
        const response = await fetch('http://localhost:3001/data');
        const data = await response.json();

        // Simple logic: count users and calculate stats
        const userCount = data.users ? data.users.length : 0;
        const activeUsers = data.users ? data.users.filter(user => user.active).length : 0;

        const processTime = Date.now() - startTime;

        // Return processed stats only
        res.json({
            processTimeMs: processTime,
            totalUsers: userCount,
            activeUsers: activeUsers,
            inactiveUsers: userCount - activeUsers,
        });
    } catch (error) {
        res.status(500).json({ error: 'Failed to fetch data from microservice' });
    }
});

app.listen(PORT, () => {
    console.log(`Gateway running on port ${PORT}`);
    console.log(`Try: http://localhost:${PORT}/stats`);
});