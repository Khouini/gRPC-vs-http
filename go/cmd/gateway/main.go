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
)

// GatewayServer handles HTTP requests and calls gRPC microservice
type GatewayServer struct {
	client pb.DataServiceClient
}

// NewGatewayServer creates a new gateway server
func NewGatewayServer(client pb.DataServiceClient) *GatewayServer {
	return &GatewayServer{client: client}
}

// handleStats processes the /stats endpoint
func (g *GatewayServer) handleStats(c *gin.Context) {
	startTime := time.Now()

	// Call gRPC microservice
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := g.client.GetUsers(ctx, &pb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from microservice"})
		return
	}

	// Process data - count users
	totalUsers := len(resp.Users)
	activeUsers := 0
	for _, user := range resp.Users {
		if user.Active {
			activeUsers++
		}
	}

	processTime := time.Since(startTime).Milliseconds()

	// Return processed stats
	stats := types.StatsResponse{
		TotalUsers:    totalUsers,
		ActiveUsers:   activeUsers,
		InactiveUsers: totalUsers - activeUsers,
		DataSize:      resp.Metadata.ActualSizeMB,
		ProcessTimeMs: processTime,
	}

	c.JSON(http.StatusOK, stats)
}

// setupRoutes configures the HTTP routes
func (g *GatewayServer) setupRoutes() *gin.Engine {
	r := gin.Default()
	r.GET("/stats", g.handleStats)
	return r
}

func main() {
	// Connect to gRPC microservice with larger message limits
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
	log.Println("Try: http://localhost:8080/stats")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
