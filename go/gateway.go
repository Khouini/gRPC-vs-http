package main

import (
	"context"
	"log"
	"net/http"
	"time"

	pb "grpc-vs-http/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StatsResponse struct {
	TotalUsers    int     `json:"totalUsers"`
	ActiveUsers   int     `json:"activeUsers"`
	InactiveUsers int     `json:"inactiveUsers"`
	DataSize      float64 `json:"dataSize"`
	ProcessTimeMs int64   `json:"processTimeMs"`
}

func main() {
	// Connect to gRPC microservice
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewDataServiceClient(conn)

	// Setup Gin
	r := gin.Default()

	r.GET("/stats", func(c *gin.Context) {
		startTime := time.Now()

		// Call gRPC microservice
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.GetUsers(ctx, &pb.Empty{})
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
		stats := StatsResponse{
			TotalUsers:    totalUsers,
			ActiveUsers:   activeUsers,
			InactiveUsers: totalUsers - activeUsers,
			DataSize:      resp.Metadata.ActualSizeMB,
			ProcessTimeMs: processTime,
		}

		c.JSON(http.StatusOK, stats)
	})

	log.Println("Gateway running on port 8080")
	log.Println("Try: http://localhost:8080/stats")
	r.Run(":8080")
}
