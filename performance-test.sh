#!/bin/bash

echo "=== Performance Comparison: gRPC vs HTTP ==="
echo ""

# Function to test API performance
test_api() {
    local url=$1
    local name=$2
    local requests=${3:-100}
    
    echo "Testing $name ($requests requests)..."
    
    # Use curl with timing and parallel requests
    time_output=$(curl -s -w "%{time_total}" -o /dev/null "$url")
    
    echo "  Single request time: ${time_output}s"
    
    # Test with multiple concurrent requests using xargs
    echo "  Testing $requests concurrent requests..."
    start_time=$(date +%s.%N)
    
    seq 1 $requests | xargs -n1 -P10 -I{} curl -s -o /dev/null "$url"
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    rps=$(echo "scale=2; $requests / $duration" | bc)
    
    echo "  Total time: ${duration}s"
    echo "  Requests per second: $rps"
    echo ""
}

# Check if services are running
echo "Checking services..."

# Check Node.js services
if curl -s http://localhost:3000/stats > /dev/null 2>&1; then
    echo "✓ Node.js services are running"
    NODE_RUNNING=true
else
    echo "✗ Node.js services not running"
    NODE_RUNNING=false
fi

# Check Go services  
if curl -s http://localhost:8080/stats > /dev/null 2>&1; then
    echo "✓ Go services are running"
    GO_RUNNING=true
else
    echo "✗ Go services not running"
    GO_RUNNING=false
fi

echo ""

# Run performance tests
if [ "$NODE_RUNNING" = true ]; then
    test_api "http://localhost:3000/stats" "Node.js HTTP" 100
fi

if [ "$GO_RUNNING" = true ]; then
    test_api "http://localhost:8080/stats" "Go gRPC->HTTP" 100
fi

# Memory usage comparison
echo "=== Memory Usage ==="
if [ "$NODE_RUNNING" = true ]; then
    echo "Node.js processes:"
    ps aux | grep "node.*microservice\|node.*gateway" | grep -v grep
fi

if [ "$GO_RUNNING" = true ]; then
    echo "Go processes:"
    ps aux | grep "go.*run\|main" | grep -v grep
fi

echo ""
echo "=== Summary ==="
echo "Optimizations applied:"
echo "- Pre-converted protobuf data (eliminates conversion overhead)"
echo "- Connection keep-alive and pooling"
echo "- Increased message size limits"
echo "- Concurrent stream limits"
echo "- Optimized timeout settings"
