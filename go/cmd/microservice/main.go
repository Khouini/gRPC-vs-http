package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	pb "grpc-vs-http/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Server implements the gRPC DataService
type Server struct {
	pb.UnimplementedDataServiceServer
	pbHotels   []*pb.Hotel  // Pre-loaded protobuf hotels
	pbMetadata *pb.Metadata // Pre-loaded protobuf metadata
}

// NewServer creates a new server instance with loaded data
func NewServer() *Server {
	pbHotels, pbMetadata := loadDataAsProtobuf()

	log.Printf("Loaded %d hotels as protobuf format", len(pbHotels))

	return &Server{
		pbHotels:   pbHotels,
		pbMetadata: pbMetadata,
	}
}

// loadDataAsProtobuf reads and parses the data.json file directly into protobuf types
func loadDataAsProtobuf() ([]*pb.Hotel, *pb.Metadata) {
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
		file, err = os.ReadFile(dataPath)
		if err == nil {
			log.Printf("Found data file at: %s", dataPath)
			break
		}
	}

	if err != nil {
		log.Fatalf("Failed to read data file from any location: %v", err)
	}

	// Parse JSON directly into a temporary structure
	var jsonData struct {
		Metadata struct {
			GeneratedAt  string  `json:"generatedAt"`
			TotalHotels  int     `json:"totalHotels"`
			GeneratedBy  string  `json:"generatedBy"`
			ActualSizeMB float64 `json:"actualSizeMB"`
			ActualHotels int     `json:"actualHotels"`
		} `json:"metadata"`
		Hotels []json.RawMessage `json:"hotels"`
	}

	if err := json.Unmarshal(file, &jsonData); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Convert metadata
	pbMetadata := &pb.Metadata{
		GeneratedAt:  jsonData.Metadata.GeneratedAt,
		TotalHotels:  int32(jsonData.Metadata.TotalHotels),
		GeneratedBy:  jsonData.Metadata.GeneratedBy,
		ActualSizeMB: jsonData.Metadata.ActualSizeMB,
		ActualHotels: int32(jsonData.Metadata.ActualHotels),
	}

	// Convert hotels - we'll parse each hotel as JSON and then unmarshal into protobuf
	pbHotels := make([]*pb.Hotel, len(jsonData.Hotels))
	for i, hotelJSON := range jsonData.Hotels {
		var hotel pb.Hotel
		if err := json.Unmarshal(hotelJSON, &hotel); err != nil {
			log.Printf("Warning: Failed to parse hotel %d: %v", i, err)
			continue
		}
		pbHotels[i] = &hotel
	}

	log.Printf("Loaded %d hotels from data file", len(pbHotels))
	return pbHotels, pbMetadata
}

// GetHotels implements the gRPC method (original non-streaming)
func (s *Server) GetHotels(ctx context.Context, req *pb.Empty) (*pb.HotelsResponse, error) {
	// Return pre-converted data - no conversion overhead!
	return &pb.HotelsResponse{
		Metadata: s.pbMetadata,
		Hotels:   s.pbHotels,
	}, nil
}

func main() {
	// Create server with loaded data
	server := NewServer()

	// Start gRPC server with optimized settings
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Optimized server options
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.MaxRecvMsgSize(100 * 1024 * 1024), // 100MB
		grpc.MaxSendMsgSize(100 * 1024 * 1024), // 100MB
		grpc.MaxConcurrentStreams(1000),        // Allow up to 1000 concurrent streams
	}

	s := grpc.NewServer(opts...)
	pb.RegisterDataServiceServer(s, server)

	log.Println("gRPC microservice running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
