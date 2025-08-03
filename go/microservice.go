package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	pb "grpc-vs-http/proto"

	"google.golang.org/grpc"
)

// Data structures matching the JSON
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	City   string `json:"city"`
	Active bool   `json:"active"`
}

type Metadata struct {
	GeneratedAt    string  `json:"generatedAt"`
	TargetSizeMB   float64 `json:"targetSizeMB"`
	EstimatedItems int     `json:"estimatedItems"`
	ActualSizeMB   float64 `json:"actualSizeMB"`
	ActualItems    int     `json:"actualItems"`
}

type DataFile struct {
	Metadata Metadata `json:"metadata"`
	Users    []User   `json:"users"`
}

// Server struct
type server struct {
	pb.UnimplementedDataServiceServer
	data *DataFile
}

// Load data at startup
func loadData() *DataFile {
	dataPath := filepath.Join("..", "data.json")
	file, err := ioutil.ReadFile(dataPath)
	if err != nil {
		log.Fatalf("Failed to read data file: %v", err)
	}

	var data DataFile
	if err := json.Unmarshal(file, &data); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Loaded %d users from data file", len(data.Users))
	return &data
}

// GetUsers implements the gRPC method
func (s *server) GetUsers(ctx context.Context, req *pb.Empty) (*pb.UsersResponse, error) {
	// Convert data to protobuf format
	pbUsers := make([]*pb.User, len(s.data.Users))
	for i, user := range s.data.Users {
		pbUsers[i] = &pb.User{
			Id:     int32(user.ID),
			Name:   user.Name,
			Email:  user.Email,
			Age:    int32(user.Age),
			City:   user.City,
			Active: user.Active,
		}
	}

	pbMetadata := &pb.Metadata{
		GeneratedAt:    s.data.Metadata.GeneratedAt,
		TargetSizeMB:   s.data.Metadata.TargetSizeMB,
		EstimatedItems: int32(s.data.Metadata.EstimatedItems),
		ActualSizeMB:   s.data.Metadata.ActualSizeMB,
		ActualItems:    int32(s.data.Metadata.ActualItems),
	}

	return &pb.UsersResponse{
		Metadata: pbMetadata,
		Users:    pbUsers,
	}, nil
}

func main() {
	// Load data once at startup
	data := loadData()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDataServiceServer(s, &server{data: data})

	log.Println("gRPC microservice running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
