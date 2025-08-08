package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
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
			if err.Error() == "EOF" {
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

// setupRoutes configures the HTTP routes
func (g *GatewayServer) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Streaming endpoint
	r.GET("/stats", g.handleStats)

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
	log.Println("  - GET /health (health check)")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
