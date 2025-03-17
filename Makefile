.PHONY: build test run proto-gen

# Build the application
build:
	go build -o bin/user_service cmd/user_service/main.go

# Run tests
test:
	go test -v ./...

# Run the application
run:
	go run cmd/user_service/main.go

# Generate gRPC code from protobuf
proto-gen:
	protoc --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       internal/user_service/adapter/controller/grpc/proto/user_service.proto

# Install required tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0