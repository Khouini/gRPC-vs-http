package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"time"

	"grpc-vs-http/internal/types"
	pb "grpc-vs-http/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Server implements the gRPC DataService
type Server struct {
	pb.UnimplementedDataServiceServer
	data        *types.DataFile
	pbUsers     []*pb.User   // Pre-converted protobuf users
	pbMetadata  *pb.Metadata // Pre-converted protobuf metadata
	activeCount int32        // Pre-calculated active user count
}

// NewServer creates a new server instance with loaded data
func NewServer() *Server {
	data := loadData()

	// Pre-convert data to protobuf format once at startup
	pbUsers := make([]*pb.User, len(data.Users))
	activeCount := int32(0)

	for i, user := range data.Users {
		pbUsers[i] = &pb.User{
			Id:     int32(user.ID),
			Name:   user.Name,
			Email:  user.Email,
			Age:    int32(user.Age),
			City:   user.City,
			Active: user.Active,
		}
		if user.Active {
			activeCount++
		}
	}

	pbMetadata := &pb.Metadata{
		GeneratedAt:    data.Metadata.GeneratedAt,
		TargetSizeMB:   data.Metadata.TargetSizeMB,
		EstimatedItems: int32(data.Metadata.EstimatedItems),
		ActualSizeMB:   data.Metadata.ActualSizeMB,
		ActualItems:    int32(data.Metadata.ActualItems),
	}

	log.Printf("Pre-converted %d users to protobuf format (%d active)", len(pbUsers), activeCount)

	return &Server{
		data:        data,
		pbUsers:     pbUsers,
		pbMetadata:  pbMetadata,
		activeCount: activeCount,
	}
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

// GetUsers implements the gRPC method (original non-streaming)
func (s *Server) GetUsers(ctx context.Context, req *pb.Empty) (*pb.UsersResponse, error) {
	// Return pre-converted data - no conversion overhead!
	return &pb.UsersResponse{
		Metadata: s.pbMetadata,
		Users:    s.pbUsers,
	}, nil
}

// GetUsersStreaming implements streaming method with chunked data
func (s *Server) GetUsersStreaming(req *pb.StreamRequest, stream pb.DataService_GetUsersStreamingServer) error {
	chunkSize := int(req.ChunkSize)
	if chunkSize <= 0 {
		chunkSize = 1000 // Default chunk size
	}

	totalUsers := len(s.pbUsers)
	totalChunks := (totalUsers + chunkSize - 1) / chunkSize // Ceiling division

	log.Printf("Streaming %d users in %d chunks of size %d", totalUsers, totalChunks, chunkSize)

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > totalUsers {
			end = totalUsers
		}

		chunk := &pb.UserChunk{
			Users:       s.pbUsers[start:end],
			ChunkIndex:  int32(i),
			TotalChunks: int32(totalChunks),
			IsLast:      i == totalChunks-1,
		}

		// Include metadata only in the first chunk
		if i == 0 {
			chunk.Metadata = s.pbMetadata
		}

		// Send chunk through stream
		if err := stream.Send(chunk); err != nil {
			log.Printf("Error sending chunk %d: %v", i, err)
			return err
		}

		// Small delay to demonstrate streaming (remove in production)
		// time.Sleep(10 * time.Millisecond)
	}

	log.Printf("Completed streaming %d chunks", totalChunks)
	return nil
}

// GetStatsOnly returns only statistics without user data (ultra-fast)
func (s *Server) GetStatsOnly(ctx context.Context, req *pb.Empty) (*pb.StatsResponse, error) {
	return &pb.StatsResponse{
		TotalUsers:  int32(len(s.pbUsers)),
		ActiveUsers: s.activeCount,
		DataSizeMB:  s.pbMetadata.ActualSizeMB,
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

	log.Println("gRPC microservice running on port 50051 with optimizations")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
