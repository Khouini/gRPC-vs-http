.PHONY: proto deps build run-micro run-gateway test clean

# Generate protobuf files
proto:
	mkdir -p proto
	protoc --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative data.proto

# Install dependencies
deps:
	go mod tidy

# Build binaries
build: proto deps
	go build -o bin/microservice ./cmd/microservice
	go build -o bin/gateway ./cmd/gateway

# Setup everything
setup: proto deps

# Run microservice
run-micro:
	cd cmd/microservice && go run main.go

# Run gateway
run-gateway:
	cd cmd/gateway && go run main.go

# Test the API
test:
	curl http://localhost:8080/stats

# Clean generated files
clean:
	rm -rf proto/ bin/
