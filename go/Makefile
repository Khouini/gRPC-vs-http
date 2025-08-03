.PHONY: proto deps run-micro run-gateway test clean

# Generate protobuf files
proto:
	mkdir -p proto
	protoc --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative data.proto

# Install dependencies
deps:
	go mod tidy

# Setup everything
setup: proto deps

# Run microservice
run-micro:
	go run microservice.go

# Run gateway
run-gateway:
	go run gateway.go

# Test the API
test:
	curl http://localhost:8080/stats

# Clean generated files
clean:
	rm -rf proto/
