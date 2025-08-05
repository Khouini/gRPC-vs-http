package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
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
	data       *types.DataFile
	pbHotels   []*pb.Hotel  // Pre-converted protobuf hotels
	pbMetadata *pb.Metadata // Pre-converted protobuf metadata
}

// NewServer creates a new server instance with loaded data
func NewServer() *Server {
	data := loadData()

	// Pre-convert data to protobuf format once at startup
	pbHotels := make([]*pb.Hotel, len(data.Hotels))

	for i, hotel := range data.Hotels {
		pbHotels[i] = convertHotelToProto(hotel)
	}

	pbMetadata := &pb.Metadata{
		GeneratedAt:  data.Metadata.GeneratedAt,
		TotalHotels:  int32(data.Metadata.TotalHotels),
		GeneratedBy:  data.Metadata.GeneratedBy,
		ActualSizeMB: data.Metadata.ActualSizeMB,
		ActualHotels: int32(data.Metadata.ActualHotels),
	}

	log.Printf("Pre-converted %d hotels to protobuf format", len(pbHotels))

	return &Server{
		data:       data,
		pbHotels:   pbHotels,
		pbMetadata: pbMetadata,
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
		file, err = os.ReadFile(dataPath)
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

	log.Printf("Loaded %d hotels from data file", len(data.Hotels))
	return &data
}

// convertHotelToProto converts a types.Hotel to pb.Hotel
func convertHotelToProto(hotel types.Hotel) *pb.Hotel {
	pbHotel := &pb.Hotel{
		ZoneId: hotel.ZoneId,
		Zone:   hotel.Zone,
	}

	// Handle optional fields
	if hotel.SupplierId != nil {
		pbHotel.SupplierId = hotel.SupplierId
	}
	if hotel.SupplierIds != nil {
		pbHotel.SupplierIds = hotel.SupplierIds
	}
	if hotel.HotelId != nil {
		pbHotel.HotelId = hotel.HotelId
	}
	if hotel.HotelIds != nil {
		pbHotel.HotelIds = hotel.HotelIds
	}
	if hotel.GiataId != nil {
		pbHotel.GiataId = hotel.GiataId
	}
	if hotel.HUid != nil {
		pbHotel.HUid = hotel.HUid
	}
	if hotel.Name != nil {
		pbHotel.Name = hotel.Name
	}
	if hotel.Rating != nil {
		pbHotel.Rating = hotel.Rating
	}
	if hotel.Address != nil {
		pbHotel.Address = hotel.Address
	}
	if hotel.Score != nil {
		pbHotel.Score = hotel.Score
	}
	if hotel.HotelChainId != nil {
		pbHotel.HotelChainId = hotel.HotelChainId
	}
	if hotel.AccTypeId != nil {
		pbHotel.AccTypeId = hotel.AccTypeId
	}
	if hotel.City != nil {
		pbHotel.City = hotel.City
	}
	if hotel.CityId != nil {
		pbHotel.CityId = hotel.CityId
	}
	if hotel.Country != nil {
		pbHotel.Country = hotel.Country
	}
	if hotel.CountryCode != nil {
		pbHotel.CountryCode = hotel.CountryCode
	}
	if hotel.CountryId != nil {
		pbHotel.CountryId = hotel.CountryId
	}
	if hotel.Lat != nil {
		pbHotel.Lat = hotel.Lat
	}
	if hotel.Long != nil {
		pbHotel.Long = hotel.Long
	}
	if hotel.MarketingText != nil {
		pbHotel.MarketingText = hotel.MarketingText
	}
	if hotel.MinRate != nil {
		pbHotel.MinRate = hotel.MinRate
	}
	if hotel.MaxRate != nil {
		pbHotel.MaxRate = hotel.MaxRate
	}
	if hotel.Currency != nil {
		pbHotel.Currency = hotel.Currency
	}
	if hotel.Photos != nil {
		pbHotel.Photos = hotel.Photos
	}
	if hotel.Total != nil {
		pbHotel.Total = hotel.Total
	}
	if hotel.Distances != nil {
		pbHotel.Distances = hotel.Distances
	}
	if hotel.Strength != nil {
		pbHotel.Strength = hotel.Strength
	}
	if hotel.Available != nil {
		pbHotel.Available = hotel.Available
	}
	if hotel.Boards != nil {
		pbHotel.Boards = hotel.Boards
	}
	if hotel.Tag != nil {
		pbHotel.Tag = hotel.Tag
	}
	if hotel.CityLat != nil {
		pbHotel.CityLat = hotel.CityLat
	}
	if hotel.CityLong != nil {
		pbHotel.CityLong = hotel.CityLong
	}
	if hotel.ReviewsSubratingsAverage != nil {
		pbHotel.ReviewsSubratingsAverage = hotel.ReviewsSubratingsAverage
	}
	if hotel.AllNRF != nil {
		pbHotel.AllNRF = hotel.AllNRF
	}
	if hotel.AllRF != nil {
		pbHotel.AllRF = hotel.AllRF
	}
	if hotel.PartialNRF != nil {
		pbHotel.PartialNRF = hotel.PartialNRF
	}

	// Convert complex nested structures (simplified for now)
	if hotel.Neighborhood != nil {
		pbHotel.Neighborhood = &pb.Neighborhood{
			Name:        hotel.Neighborhood.Name,
			Description: hotel.Neighborhood.Description,
		}
	}

	if hotel.Review != nil {
		pbHotel.Review = &pb.Review{
			Score:   hotel.Review.Score,
			Count:   hotel.Review.Count,
			Average: hotel.Review.Average,
		}
	}

	return pbHotel
}

// GetHotels implements the gRPC method (original non-streaming)
func (s *Server) GetHotels(ctx context.Context, req *pb.Empty) (*pb.HotelsResponse, error) {
	// Return pre-converted data - no conversion overhead!
	return &pb.HotelsResponse{
		Metadata: s.pbMetadata,
		Hotels:   s.pbHotels,
	}, nil
}

// GetHotelsStreaming implements streaming method with chunked data
func (s *Server) GetHotelsStreaming(req *pb.StreamRequest, stream pb.DataService_GetHotelsStreamingServer) error {
	chunkSize := int(req.ChunkSize)
	if chunkSize <= 0 {
		chunkSize = 100 // Default chunk size for hotels
	}

	totalHotels := len(s.pbHotels)
	totalChunks := (totalHotels + chunkSize - 1) / chunkSize // Ceiling division

	log.Printf("Streaming %d hotels in %d chunks of size %d", totalHotels, totalChunks, chunkSize)

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > totalHotels {
			end = totalHotels
		}

		chunk := &pb.HotelChunk{
			Hotels:      s.pbHotels[start:end],
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
	}

	log.Printf("Completed streaming %d chunks", totalChunks)
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
