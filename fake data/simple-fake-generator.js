#!/usr/bin/env node

/**
 * Simple Fast Fake Hotel Data Generator
 * Generates fake hotel JSON data based on number of hotels
 */

const fs = require('fs');

class SimpleFakeGenerator {
    constructor() {
        // Pre-calculate common data to avoid repeated generation
        this.hotelNames = [
            'Grand Palace Hotel', 'Ocean View Resort', 'City Center Inn', 'Mountain Lodge',
            'Sunset Beach Hotel', 'Royal Garden Hotel', 'Urban Boutique Hotel', 'Seaside Resort',
            'Metropolitan Hotel', 'Paradise Resort', 'Golden Gate Hotel', 'Riverside Inn'
        ];
        this.cities = [
            'New York', 'Los Angeles', 'Chicago', 'Houston', 'Phoenix', 'Philadelphia',
            'San Antonio', 'San Diego', 'Dallas', 'San Jose', 'Austin', 'Jacksonville'
        ];
        this.countries = ['USA', 'Canada', 'Mexico', 'UK', 'France', 'Germany', 'Spain', 'Italy'];
        this.countryCodes = ['US', 'CA', 'MX', 'GB', 'FR', 'DE', 'ES', 'IT'];
        this.zones = ['Downtown', 'Airport', 'Beach', 'Mountain', 'Suburban', 'Historic District'];
        this.currencies = ['USD', 'EUR', 'GBP', 'CAD'];
        this.boards = ['BB', 'HB', 'FB', 'AI', 'RO'];
        this.roomTypes = ['Standard', 'Deluxe', 'Suite', 'Executive', 'Presidential'];
        this.tags = ['family', 'business', 'luxury', 'budget', 'romantic', 'adventure'];
    }

    // Generate a hotel object
    generateHotel(id) {
        const countryIndex = id % this.countries.length;
        const cityIndex = id % this.cities.length;
        const rating = 1 + (id % 5); // 1-5 stars
        const score = 6 + (id % 5); // 6-10 score

        return {
            supplierId: 1000 + (id % 100),
            supplierIds: [1000 + (id % 100), 2000 + (id % 50)],
            hotelId: `HTL${String(id).padStart(6, '0')}`,
            hotelIds: [`HTL${String(id).padStart(6, '0')}`, `ALT${String(id).padStart(6, '0')}`],
            giataId: 100000 + id,
            hUid: 500000 + id,
            name: `${this.hotelNames[id % this.hotelNames.length]} ${id}`,
            rating: rating,
            address: `${100 + (id % 900)} Main Street, ${this.cities[cityIndex]}`,
            score: score + (id % 100) / 100,
            hotelChainId: 10 + (id % 20),
            accTypeId: 1 + (id % 5),
            city: this.cities[cityIndex],
            cityId: 1000 + cityIndex,
            zoneId: 1 + (id % 10),
            zone: this.zones[id % this.zones.length],
            country: this.countries[countryIndex],
            countryCode: this.countryCodes[countryIndex],
            countryId: 1 + countryIndex,
            lat: 40.0 + (id % 100) / 10,
            long: -74.0 + (id % 100) / 10,
            marketingText: `Experience luxury and comfort at ${this.hotelNames[id % this.hotelNames.length]}. Perfect for your stay.`,
            minRate: 50 + (id % 200),
            maxRate: 200 + (id % 500),
            currency: this.currencies[id % this.currencies.length],
            photos: [
                `https://example.com/hotel${id}/photo1.jpg`,
                `https://example.com/hotel${id}/photo2.jpg`
            ],
            rooms: this.generateRooms(id),
            supplements: this.generateSupplements(id),
            total: 150 + (id % 300),
            distances: {
                'airport': 5 + (id % 20),
                'city_center': 2 + (id % 10),
                'beach': 1 + (id % 15)
            },
            neighborhood: {
                name: `${this.zones[id % this.zones.length]} Area`,
                description: 'Prime location with easy access to attractions'
            },
            strength: {
                'location': (id % 2) === 0,
                'service': (id % 3) === 0,
                'facilities': (id % 4) === 0
            },
            review: {
                score: score + (id % 100) / 100,
                count: 50 + (id % 200),
                average: score + (id % 100) / 100
            },
            available: (id % 10) !== 0, // 90% availability
            boards: this.boards.slice(0, 1 + (id % 3)),
            tag: this.tags[id % this.tags.length],
            cityLat: 40.0 + (cityIndex % 10),
            cityLong: -74.0 + (cityIndex % 10),
            reviews: this.generateReviews(id),
            reviewsSubratingsAverage: {
                'cleanliness': 7 + (id % 3),
                'service': 6 + (id % 4),
                'location': 8 + (id % 2),
                'value': 7 + (id % 3)
            },
            allNRF: (id % 5) === 0,
            allRF: (id % 7) === 0,
            partialNRF: (id % 3) === 0
        };
    }

    generateRooms(hotelId) {
        const numRooms = 20 + (hotelId % 31); // 20-50 rooms per hotel
        const rooms = [];

        for (let i = 0; i < numRooms; i++) {
            const roomTypeIndex = i % this.roomTypes.length;
            const floorNumber = Math.floor(i / 20) + 1; // 20 rooms per floor
            const roomNumber = (i % 20) + 1;

            rooms.push({
                code: `RM${hotelId}${String(i).padStart(3, '0')}`,
                codes: [`RM${hotelId}${String(i).padStart(3, '0')}`, `ALT${hotelId}${String(i).padStart(3, '0')}`],
                name: `${this.roomTypes[roomTypeIndex]} Room ${floorNumber}${String(roomNumber).padStart(2, '0')}`,
                names: [`${this.roomTypes[roomTypeIndex]} Room ${floorNumber}${String(roomNumber).padStart(2, '0')}`],
                rates: this.generateRates(hotelId, i),
                category: this.roomTypes[roomTypeIndex],
                total: 100 + (hotelId % 200) + (roomTypeIndex * 50) + (i % 100),
                originalCode: `ORIG${hotelId}${String(i).padStart(3, '0')}`,
                originalName: `Original ${this.roomTypes[roomTypeIndex]} ${floorNumber}${String(roomNumber).padStart(2, '0')}`
            });
        }

        return rooms;
    } generateRates(hotelId, roomIndex) {
        const numRates = 1 + (hotelId % 2); // 1-2 rates per room
        const rates = [];

        for (let i = 0; i < numRates; i++) {
            rates.push({
                rateKey: `RK${hotelId}${roomIndex}${i}`,
                rateClass: 'NOR',
                contractId: 1000 + (hotelId % 100),
                rateType: 'BOOKABLE',
                paymentType: 'AT_HOTEL',
                allotment: 5 + (hotelId % 10),
                availability: 'OK',
                amount: 80 + (hotelId % 150) + (i * 20),
                currency: this.currencies[hotelId % this.currencies.length],
                boardCode: this.boards[i % this.boards.length],
                boardName: `${this.boards[i % this.boards.length]} Board`,
                nrf: (hotelId % 3) === 0,
                rooms: 1 + (hotelId % 3),
                adults: '2',
                children: '0',
                infant: '0',
                childrenAges: '',
                rateComments: 'Standard rate conditions apply',
                packaging: false,
                total: 100 + (hotelId % 200) + (i * 30),
                purchasePrice: 90 + (hotelId % 180) + (i * 25),
                cancellationPolicies: [{
                    amount: 50,
                    from: '2024-12-01',
                    realFrom: '2024-12-01T00:00:00',
                    name: 'Non-refundable',
                    purchasePrice: 45
                }],
                taxes: [{
                    name: 'City Tax',
                    amount: 5,
                    currency: this.currencies[hotelId % this.currencies.length],
                    included: false,
                    type: 'LOCAL'
                }]
            });
        }

        return rates;
    }

    generateSupplements(hotelId) {
        return [
            {
                name: 'Breakfast',
                amount: 15 + (hotelId % 20),
                currency: this.currencies[hotelId % this.currencies.length],
                included: (hotelId % 3) === 0
            },
            {
                name: 'Parking',
                amount: 10 + (hotelId % 15),
                currency: this.currencies[hotelId % this.currencies.length],
                included: (hotelId % 4) === 0
            }
        ];
    }

    generateReviews(hotelId) {
        const numReviews = 1 + (hotelId % 3); // 1-3 reviews per hotel
        const reviews = [];

        for (let i = 0; i < numReviews; i++) {
            reviews.push({
                id: `REV${hotelId}${i}`,
                rating: 7 + (hotelId % 3),
                comment: `Great stay at hotel ${hotelId}. Excellent service and facilities.`,
                author: `Guest${hotelId}${i}`,
                date: '2024-01-01',
                subratings: {
                    'cleanliness': 7 + (hotelId % 3),
                    'service': 6 + (hotelId % 4),
                    'location': 8 + (hotelId % 2),
                    'value': 7 + (hotelId % 3)
                }
            });
        }

        return reviews;
    }

    // Generate data with specified number of hotels
    generateData(numberOfHotels) {
        console.log(`Generating ${numberOfHotels} hotels...`);

        const data = {
            metadata: {
                generatedAt: new Date().toISOString(),
                totalHotels: numberOfHotels,
                generatedBy: 'SimpleFakeHotelGenerator'
            },
            hotels: []
        };

        // Generate hotels in batches for better performance
        const batchSize = 50; // Increased batch size due to smaller hotel objects with 20-50 rooms
        let generated = 0;

        while (generated < numberOfHotels) {
            const batch = [];
            const batchEnd = Math.min(generated + batchSize, numberOfHotels);

            for (let i = generated; i < batchEnd; i++) {
                batch.push(this.generateHotel(i + 1));
            }

            data.hotels.push(...batch);
            generated = batchEnd;

            // Show progress every 100 items
            if (generated % 100 === 0) {
                console.log(`Generated ${generated.toLocaleString()} hotels...`);
            }
        }

        // Calculate actual size
        const jsonString = JSON.stringify(data);
        const actualSizeMB = Buffer.byteLength(jsonString, 'utf8') / (1024 * 1024);

        data.metadata.actualSizeMB = parseFloat(actualSizeMB.toFixed(2));
        data.metadata.actualHotels = data.hotels.length;

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
Simple Fast Fake Hotel Data Generator

Usage: node simple-fake-generator.js <number_of_hotels> [output_file]

Examples:
  node simple-fake-generator.js 1000
  node simple-fake-generator.js 5000 hotels-data.json
        `);
        return;
    }

    const numberOfHotels = parseInt(args[0]);
    if (isNaN(numberOfHotels) || numberOfHotels <= 0) {
        console.error('Please provide a valid number of hotels');
        return;
    }

    const outputFile = args[1] || `hotels-${numberOfHotels}.json`;
    const generator = new SimpleFakeGenerator();

    console.time('Generation Time');

    try {
        const { data, jsonString } = generator.generateData(numberOfHotels);
        generator.saveToFile(jsonString, outputFile);

        console.timeEnd('Generation Time');
        console.log(`Generated: ${data.metadata.actualHotels} hotels`);
        console.log(`File size: ${data.metadata.actualSizeMB}MB`);

    } catch (error) {
        console.error('Error:', error.message);
    }
}

if (require.main === module) {
    main();
}

module.exports = SimpleFakeGenerator;
