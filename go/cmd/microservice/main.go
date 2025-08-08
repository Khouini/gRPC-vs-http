package main

import (
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

// DataFile represents the structure of the data.json file
type DataFile struct {
	Metadata json.RawMessage `json:"metadata"`
	Hotels   json.RawMessage `json:"hotels"`
}

// Server implements the gRPC DataService
type Server struct {
	pb.UnimplementedDataServiceServer
	pbHotels   []*pb.Hotel  // Pre-converted protobuf hotels
	pbMetadata *pb.Metadata // Pre-converted protobuf metadata
}

// NewServer creates a new server instance with loaded data
func NewServer() *Server {
	pbHotels, pbMetadata := loadData()

	log.Printf("Loaded %d hotels from data file", len(pbHotels))

	return &Server{
		pbHotels:   pbHotels,
		pbMetadata: pbMetadata,
	}
}

// loadData reads and parses the data.json file directly into protobuf types
func loadData() ([]*pb.Hotel, *pb.Metadata) {
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

	var data DataFile
	if err := json.Unmarshal(file, &data); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Parse metadata
	var metadata pb.Metadata
	if err := json.Unmarshal(data.Metadata, &metadata); err != nil {
		log.Fatalf("Failed to parse metadata: %v", err)
	}

	// Parse hotels array
	var hotels []*pb.Hotel
	if err := json.Unmarshal(data.Hotels, &hotels); err != nil {
		log.Fatalf("Failed to parse hotels: %v", err)
	}

	return hotels, &metadata
}

// GetHotelsStreaming implements the streaming gRPC method
func (s *Server) GetHotelsStreaming(req *pb.StreamRequest, stream pb.DataService_GetHotelsStreamingServer) error {
	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 100 // Default chunk size
	}

	totalHotels := len(s.pbHotels)
	totalChunks := (totalHotels + int(chunkSize) - 1) / int(chunkSize) // Ceiling division

	for i := 0; i < totalHotels; i += int(chunkSize) {
		end := i + int(chunkSize)
		if end > totalHotels {
			end = totalHotels
		}

		chunk := &pb.HotelChunk{
			Hotels:      s.pbHotels[i:end],
			ChunkIndex:  int32(i / int(chunkSize)),
			TotalChunks: int32(totalChunks),
			IsLast:      end == totalHotels,
		}

		// Include metadata only in the first chunk
		if i == 0 {
			chunk.Metadata = s.pbMetadata
		}

		if err := stream.Send(chunk); err != nil {
			return err
		}
	}

	return nil
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
		grpc.MaxRecvMsgSize(1000 * 1024 * 1024), // 100MB
		grpc.MaxSendMsgSize(1000 * 1024 * 1024), // 100MB
		grpc.MaxConcurrentStreams(1000),         // Allow up to 1000 concurrent streams
	}

	s := grpc.NewServer(opts...)
	pb.RegisterDataServiceServer(s, server)

	log.Println("gRPC microservice running on port 50051 with optimizations")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
