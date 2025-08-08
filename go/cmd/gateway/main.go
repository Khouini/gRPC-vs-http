package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	pb "grpc-vs-http/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// StatsResponse represents the response from the gateway
type StatsResponse struct {
	ProcessTimeMs   int64 `json:"processTimeMs"`
	TotalHotels     int   `json:"totalHotels"`
	AvailableHotels int   `json:"availableHotels"`
}

// ConcurrentStatsResponse represents the response from concurrent stats testing
type ConcurrentStatsResponse struct {
	TotalTimeMs     int64           `json:"totalTimeMs"`
	ConcurrentCalls int             `json:"concurrentCalls"`
	SuccessfulCalls int             `json:"successfulCalls"`
	FailedCalls     int             `json:"failedCalls"`
	AverageTimeMs   float64         `json:"averageTimeMs"`
	MinTimeMs       int64           `json:"minTimeMs"`
	MaxTimeMs       int64           `json:"maxTimeMs"`
	Results         []StatsResponse `json:"results"`
}

// GatewayServer handles HTTP requests and calls gRPC microservice
type GatewayServer struct {
	client pb.DataServiceClient
}

// NewGatewayServer creates a new gateway server
func NewGatewayServer(client pb.DataServiceClient) *GatewayServer {
	return &GatewayServer{client: client}
}

// handleStats processes the /stats endpoint using streaming
func (g *GatewayServer) handleStats(c *gin.Context) {
	startTime := time.Now()

	// Get chunk size from query parameter, default to 100
	chunkSize := int32(100)
	if chunkParam := c.Query("chunkSize"); chunkParam != "" {
		if parsed, err := strconv.ParseInt(chunkParam, 10, 32); err == nil && parsed > 0 {
			chunkSize = int32(parsed)
		}
	}

	log.Printf("Processing stats with chunk size: %d", chunkSize)

	// Call gRPC microservice using streaming
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := g.client.GetHotelsStreaming(ctx, &pb.StreamRequest{ChunkSize: chunkSize})
	if err != nil {
		log.Printf("gRPC streaming call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from microservice"})
		return
	}

	var totalHotels int
	var availableHotels int

	// Receive all chunks and process them
	for {
		chunk, err := stream.Recv()
		if err != nil {
			// if err.Error() == "EOF" {
			if err == io.EOF {
				break
			}
			log.Printf("gRPC stream receive failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive data from microservice"})
			return
		}

		// Count hotels in this chunk
		totalHotels += len(chunk.Hotels)
		for _, hotel := range chunk.Hotels {
			if hotel.Available != nil && *hotel.Available {
				availableHotels++
			}
		}
	}

	processTime := time.Since(startTime).Milliseconds()

	stats := StatsResponse{
		ProcessTimeMs:   processTime,
		TotalHotels:     totalHotels,
		AvailableHotels: availableHotels,
	}

	c.JSON(http.StatusOK, stats)
}

// handleConcurrentStats processes multiple concurrent calls to the stats endpoint
func (g *GatewayServer) handleConcurrentStats(c *gin.Context) {
	startTime := time.Now()

	// Get number of concurrent calls from query parameter, default to 10
	concurrentCalls := 10
	if callsParam := c.Query("calls"); callsParam != "" {
		if parsed, err := strconv.Atoi(callsParam); err == nil && parsed > 0 && parsed <= 100 {
			concurrentCalls = parsed
		}
	}

	// Get chunk size from query parameter, default to 100
	chunkSize := int32(100)
	if chunkParam := c.Query("chunkSize"); chunkParam != "" {
		if parsed, err := strconv.ParseInt(chunkParam, 10, 32); err == nil && parsed > 0 {
			chunkSize = int32(parsed)
		}
	}

	log.Printf("Processing %d concurrent stats calls with chunk size: %d", concurrentCalls, chunkSize)

	// Create channels for collecting results
	resultsChan := make(chan StatsResponse, concurrentCalls)
	errorsChan := make(chan error, concurrentCalls)

	// Launch concurrent goroutines
	var wg sync.WaitGroup
	for i := 0; i < concurrentCalls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Call gRPC microservice using streaming
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			callStartTime := time.Now()

			stream, err := g.client.GetHotelsStreaming(ctx, &pb.StreamRequest{ChunkSize: chunkSize})
			if err != nil {
				errorsChan <- err
				return
			}

			var totalHotels int
			var availableHotels int

			// Receive all chunks and process them
			for {
				chunk, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					errorsChan <- err
					return
				}

				// Count hotels in this chunk
				totalHotels += len(chunk.Hotels)
				for _, hotel := range chunk.Hotels {
					if hotel.Available != nil && *hotel.Available {
						availableHotels++
					}
				}
			}

			callProcessTime := time.Since(callStartTime).Milliseconds()

			result := StatsResponse{
				ProcessTimeMs:   callProcessTime,
				TotalHotels:     totalHotels,
				AvailableHotels: availableHotels,
			}

			resultsChan <- result
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Collect results
	var results []StatsResponse
	var errors []error

	for result := range resultsChan {
		results = append(results, result)
	}

	for err := range errorsChan {
		errors = append(errors, err)
	}

	totalTime := time.Since(startTime).Milliseconds()

	// Calculate statistics
	var minTime int64 = 999999
	var maxTime int64 = 0
	var totalProcessTime int64 = 0

	for _, result := range results {
		if result.ProcessTimeMs < minTime {
			minTime = result.ProcessTimeMs
		}
		if result.ProcessTimeMs > maxTime {
			maxTime = result.ProcessTimeMs
		}
		totalProcessTime += result.ProcessTimeMs
	}

	averageTime := float64(0)
	if len(results) > 0 {
		averageTime = float64(totalProcessTime) / float64(len(results))
	}

	response := ConcurrentStatsResponse{
		TotalTimeMs:     totalTime,
		ConcurrentCalls: concurrentCalls,
		SuccessfulCalls: len(results),
		FailedCalls:     len(errors),
		AverageTimeMs:   averageTime,
		MinTimeMs:       minTime,
		MaxTimeMs:       maxTime,
		Results:         results,
	}

	c.JSON(http.StatusOK, response)
}

// setupRoutes configures the HTTP routes
func (g *GatewayServer) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Streaming endpoint
	r.GET("/stats", g.handleStats)

	// Concurrent stats endpoint
	r.GET("/concurrent-stats", g.handleConcurrentStats)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return r
}

func main() {
	// Connect to gRPC microservice with optimized settings
	kacp := keepalive.ClientParameters{
		Time:                2 * time.Minute,
		Timeout:             20 * time.Second,
		PermitWithoutStream: true,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1000*1024*1024), // 100MB
			grpc.MaxCallSendMsgSize(1000*1024*1024), // 100MB
		),
	}

	conn, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewDataServiceClient(conn)
	gateway := NewGatewayServer(client)

	// Setup routes
	router := gateway.setupRoutes()

	log.Println("Gateway running on port 8080")
	log.Println("Endpoints:")
	log.Println("  - GET /stats?chunkSize=<size> (hotel statistics with configurable chunk size, default: 100)")
	log.Println("  - GET /concurrent-stats?calls=<num>&chunkSize=<size> (concurrent hotel statistics, default: 10 calls)")
	log.Println("  - GET /health (health check)")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
