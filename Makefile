.PHONY: build test run proto-gen install-tools

# Build the application
build: build-user build-product

build-user:
	go build -o bin/user_service cmd/user_service/main.go

build-product:
	go build -o bin/product_service cmd/product_service/main.go

# Run tests
test: test-user test-product

test-user:
	go test -v ./...

test-product:
	go test -v ./internal/product_service/...

# Run the application
run: run-user run-product

run-user:
	go run cmd/user_service/main.go -config=config.user.local.yaml

run-product:
	go run cmd/product_service/main.go -config=config.product.local.yaml

# Generate gRPC code from protobuf
proto-gen: proto-gen-user proto-gen-product

proto-gen-user:
	protoc --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       internal/user_service/adapter/controller/grpc/proto/user_service.proto

proto-gen-product:
	protoc --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       internal/product_service/adapter/controller/grpc/proto/product_service.proto

# Install required tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
