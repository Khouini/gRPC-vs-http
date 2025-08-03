#!/usr/bin/env node

/**
 * Simple Fast Fake JSON Data Generator
 * Generates basic fake JSON data quickly based on target size in MB
 */

const fs = require('fs');

class SimpleFakeGenerator {
    constructor() {
        // Pre-calculate common strings to avoid repeated generation
        this.names = ['John', 'Jane', 'Mike', 'Sara', 'Alex', 'Emma', 'Chris', 'Lisa'];
        this.emails = ['gmail.com', 'yahoo.com', 'hotmail.com'];
        this.cities = ['NYC', 'LA', 'Chicago', 'Houston', 'Phoenix'];
    }

    // Generate a simple user object
    generateUser(id) {
        return {
            id: id,
            name: this.names[id % this.names.length],
            email: `user${id}@${this.emails[id % this.emails.length]}`,
            age: 20 + (id % 50),
            city: this.cities[id % this.cities.length],
            active: id % 2 === 0
        };
    }

    // Generate data to approximate target size
    generateData(targetSizeMB) {
        const targetBytes = targetSizeMB * 1024 * 1024;

        // Estimate: each user object is roughly 100-120 bytes when JSON stringified
        const estimatedItemSize = 110;
        const estimatedItemCount = Math.floor(targetBytes / estimatedItemSize);

        console.log(`Generating approximately ${estimatedItemCount} items for ${targetSizeMB}MB...`);

        const data = {
            metadata: {
                generatedAt: new Date().toISOString(),
                targetSizeMB: targetSizeMB,
                estimatedItems: estimatedItemCount
            },
            users: []
        };

        // Generate items in batches for better performance
        const batchSize = 1000;
        let generated = 0;

        while (generated < estimatedItemCount) {
            const batch = [];
            const batchEnd = Math.min(generated + batchSize, estimatedItemCount);

            for (let i = generated; i < batchEnd; i++) {
                batch.push(this.generateUser(i + 1));
            }

            data.users.push(...batch);
            generated = batchEnd;

            // Show progress every 10k items
            if (generated % 10000 === 0) {
                console.log(`Generated ${generated.toLocaleString()} items...`);
            }
        }

        // Calculate actual size
        const jsonString = JSON.stringify(data);
        const actualSizeMB = Buffer.byteLength(jsonString, 'utf8') / (1024 * 1024);

        data.metadata.actualSizeMB = parseFloat(actualSizeMB.toFixed(2));
        data.metadata.actualItems = data.users.length;

        return { data, jsonString };
    }

    saveToFile(jsonString, filename) {
        fs.writeFileSync(filename, jsonString, 'utf8');
        const fileSizeMB = fs.statSync(filename).size / (1024 * 1024);
        console.log(`\nSaved to: ${filename}`);
        console.log(`File size: ${fileSizeMB.toFixed(2)}MB`);
    }
}

// CLI functionality
function main() {
    const args = process.argv.slice(2);

    if (args.length === 0 || args.includes('--help')) {
        console.log(`
Simple Fast Fake JSON Data Generator

Usage: node simple-fake-generator.js <size_in_mb> [output_file]

Examples:
  node simple-fake-generator.js 5
  node simple-fake-generator.js 10 my-data.json
        `);
        return;
    }

    const sizeMB = parseFloat(args[0]);
    if (isNaN(sizeMB) || sizeMB <= 0) {
        console.error('Please provide a valid size in MB');
        return;
    }

    const outputFile = args[1] || `simple-data-${sizeMB}mb.json`;
    const generator = new SimpleFakeGenerator();

    console.time('Generation Time');

    try {
        const { data, jsonString } = generator.generateData(sizeMB);
        generator.saveToFile(jsonString, outputFile);

        console.timeEnd('Generation Time');
        console.log(`Target: ${sizeMB}MB, Actual: ${data.metadata.actualSizeMB}MB`);
        console.log(`Items: ${data.metadata.actualItems.toLocaleString()}`);

    } catch (error) {
        console.error('Error:', error.message);
    }
}

if (require.main === module) {
    main();
}

module.exports = SimpleFakeGenerator;
