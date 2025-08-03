const express = require('express');
const fs = require('fs');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3001;

// Middleware
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
    res.json({ status: 'ok', service: 'microservice', timestamp: new Date().toISOString() });
});

// Get data endpoint - returns the raw data.json file
app.get('/data', async (req, res) => {
    try {
        const startTime = Date.now();
        const dataPath = path.join(__dirname, '..', 'data.json');
        
        console.log(`[${new Date().toISOString()}] Reading data from: ${dataPath}`);
        
        // Check if file exists
        if (!fs.existsSync(dataPath)) {
            return res.status(404).json({ 
                error: 'Data file not found',
                path: dataPath 
            });
        }

        // Get file stats
        const stats = fs.statSync(dataPath);
        const fileSizeMB = (stats.size / (1024 * 1024)).toFixed(2);
        
        console.log(`[${new Date().toISOString()}] File size: ${fileSizeMB}MB`);

        // Read file asynchronously for better performance
        const data = await new Promise((resolve, reject) => {
            fs.readFile(dataPath, 'utf8', (err, content) => {
                if (err) reject(err);
                else resolve(content);
            });
        });

        // Parse JSON
        const jsonData = JSON.parse(data);
        const processingTime = Date.now() - startTime;

        console.log(`[${new Date().toISOString()}] Data read successfully in ${processingTime}ms`);

        res.json({
            success: true,
            data: jsonData,
            metadata: {
                fileSizeMB: parseFloat(fileSizeMB),
                processingTimeMs: processingTime,
                timestamp: new Date().toISOString(),
                service: 'microservice'
            }
        });

    } catch (error) {
        console.error(`[${new Date().toISOString()}] Error reading data:`, error.message);
        
        res.status(500).json({
            success: false,
            error: error.message,
            timestamp: new Date().toISOString(),
            service: 'microservice'
        });
    }
});

// Get data stream endpoint - for large files, stream the response
app.get('/data/stream', (req, res) => {
    try {
        const dataPath = path.join(__dirname, '..', 'data.json');
        
        console.log(`[${new Date().toISOString()}] Streaming data from: ${dataPath}`);
        
        if (!fs.existsSync(dataPath)) {
            return res.status(404).json({ 
                error: 'Data file not found',
                path: dataPath 
            });
        }

        const stats = fs.statSync(dataPath);
        const fileSizeMB = (stats.size / (1024 * 1024)).toFixed(2);
        
        res.setHeader('Content-Type', 'application/json');
        res.setHeader('Content-Length', stats.size);
        res.setHeader('X-File-Size-MB', fileSizeMB);
        res.setHeader('X-Service', 'microservice');
        
        const readStream = fs.createReadStream(dataPath);
        
        readStream.on('error', (error) => {
            console.error(`[${new Date().toISOString()}] Stream error:`, error.message);
            if (!res.headersSent) {
                res.status(500).json({ error: error.message });
            }
        });

        readStream.pipe(res);
        
        console.log(`[${new Date().toISOString()}] Started streaming ${fileSizeMB}MB file`);

    } catch (error) {
        console.error(`[${new Date().toISOString()}] Error streaming data:`, error.message);
        res.status(500).json({ error: error.message });
    }
});

// Error handling middleware
app.use((error, req, res, next) => {
    console.error('Unhandled error:', error);
    res.status(500).json({ 
        error: 'Internal server error', 
        message: error.message,
        service: 'microservice'
    });
});

app.listen(PORT, () => {
    console.log(`🚀 Microservice running on port ${PORT}`);
    console.log(`📊 Data endpoint: http://localhost:${PORT}/data`);
    console.log(`🌊 Stream endpoint: http://localhost:${PORT}/data/stream`);
    console.log(`❤️  Health check: http://localhost:${PORT}/health`);
});

module.exports = app;
