package main

import (
	"context"
	"log"
	"net/http"
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

// setupRoutes configures the HTTP routes
func (g *GatewayServer) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Original endpoint (loads all data at once)
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

	log.Println("Gateway running on port 8080")
	log.Println("Endpoints:")
	log.Println("  - GET /stats (hotel statistics)")
	log.Println("  - GET /health (health check)")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
