package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	"grpc-vs-http/internal/types"
	pb "grpc-vs-http/proto"

	"google.golang.org/grpc"
)

// Server implements the gRPC DataService
type Server struct {
	pb.UnimplementedDataServiceServer
	data *types.DataFile
}

// NewServer creates a new server instance with loaded data
func NewServer() *Server {
	data := loadData()
	return &Server{data: data}
}

// loadData reads and parses the data.json file
func loadData() *types.DataFile {
	// Try multiple possible paths for the data file
	possiblePaths := []string{
		"../../../data.json", // When running from cmd/microservice/
		"../../data.json",    // When running from go/
		"data.json",          // When running from project root
		"../data.json",       // Alternative path
	}

	var dataPath string
	var file []byte
	var err error

	for _, path := range possiblePaths {
		dataPath = filepath.Join(path)
		file, err = ioutil.ReadFile(dataPath)
		if err == nil {
			log.Printf("Found data file at: %s", dataPath)
			break
		}
	}

	if err != nil {
		log.Fatalf("Failed to read data file from any location: %v", err)
	}

	var data types.DataFile
	if err := json.Unmarshal(file, &data); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Loaded %d users from data file", len(data.Users))
	return &data
}

// GetUsers implements the gRPC method
func (s *Server) GetUsers(ctx context.Context, req *pb.Empty) (*pb.UsersResponse, error) {
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
	// Create server with loaded data
	server := NewServer()

	// Start gRPC server with larger message limits
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Set max message size to 100MB for both send and receive
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(100 * 1024 * 1024), // 100MB
		grpc.MaxSendMsgSize(100 * 1024 * 1024), // 100MB
	}

	s := grpc.NewServer(opts...)
	pb.RegisterDataServiceServer(s, server)

	log.Println("gRPC microservice running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
