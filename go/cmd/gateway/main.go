package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"grpc-vs-http/internal/types"
	pb "grpc-vs-http/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GatewayServer handles HTTP requests and calls gRPC microservice
type GatewayServer struct {
	client pb.DataServiceClient
}

// NewGatewayServer creates a new gateway server
func NewGatewayServer(client pb.DataServiceClient) *GatewayServer {
	return &GatewayServer{client: client}
}

// handleStats processes the /stats endpoint (original method)
func (g *GatewayServer) handleStats(c *gin.Context) {
	startTime := time.Now()

	// Call gRPC microservice
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := g.client.GetHotels(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from microservice"})
		return
	}

	// Count available hotels
	totalHotels := len(resp.Hotels)
	availableHotels := 0

	for _, hotel := range resp.Hotels {
		if hotel.Available != nil && *hotel.Available {
			availableHotels++
		}
	}

	processTime := time.Since(startTime).Milliseconds()

	stats := types.StatsResponse{
		TotalHotels:     totalHotels,
		AvailableHotels: availableHotels,
		DataSize:        resp.Metadata.ActualSizeMB,
		ProcessTimeMs:   processTime,
	}

	c.JSON(http.StatusOK, stats)
}

// handleStatsStreaming processes the /stats-streaming endpoint
func (g *GatewayServer) handleStatsStreaming(c *gin.Context) {
	startTime := time.Now()

	// Get chunk size from query parameter (default: 100)
	chunkSizeStr := c.DefaultQuery("chunkSize", "100")
	chunkSize, err := strconv.Atoi(chunkSizeStr)
	if err != nil || chunkSize <= 0 {
		chunkSize = 100
	}

	// Call streaming gRPC method
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := g.client.GetHotelsStreaming(ctx, &pb.StreamRequest{
		ChunkSize: int32(chunkSize),
	})
	if err != nil {
		log.Printf("gRPC streaming call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start streaming from microservice"})
		return
	}

	// Process streaming chunks
	totalHotels := 0
	availableHotels := 0
	chunksReceived := 0
	var metadata *pb.Metadata

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break // End of stream
		}
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error receiving data chunk"})
			return
		}

		// Process chunk
		chunksReceived++
		totalHotels += len(chunk.Hotels)

		for _, hotel := range chunk.Hotels {
			if hotel.Available != nil && *hotel.Available {
				availableHotels++
			}
		}

		// Get metadata from first chunk
		if chunk.Metadata != nil {
			metadata = chunk.Metadata
		}

		log.Printf("Processed chunk %d/%d with %d hotels", chunk.ChunkIndex+1, chunk.TotalChunks, len(chunk.Hotels))
	}

	processTime := time.Since(startTime).Milliseconds()

	stats := types.StatsResponse{
		TotalHotels:     totalHotels,
		AvailableHotels: availableHotels,
		DataSize:        metadata.ActualSizeMB,
		ProcessTimeMs:   processTime,
	}

	// Add streaming info
	response := gin.H{
		"stats":          stats,
		"chunksReceived": chunksReceived,
		"chunkSize":      chunkSize,
		"streamingTime":  processTime,
	}

	c.JSON(http.StatusOK, response)
}

// setupRoutes configures the HTTP routes
func (g *GatewayServer) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Original endpoint (loads all data at once)
	r.GET("/stats", g.handleStats)

	// Streaming endpoint (processes data in chunks)
	r.GET("/stats-streaming", g.handleStatsStreaming)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return r
}

func main() {
	// Connect to gRPC microservice with optimized settings
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // Send keepalive pings every 10 seconds
		Timeout:             time.Second,      // Wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // Send pings even without active streams
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024), // 100MB
			grpc.MaxCallSendMsgSize(100*1024*1024), // 100MB
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

	log.Println("Gateway running on port 8080 with streaming support")
	log.Println("Endpoints:")
	log.Println("  - GET /stats (original)")
	log.Println("  - GET /stats-streaming?chunkSize=100 (chunked streaming)")
	log.Println("  - GET /health (health check)")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
