build:
	@go build -o ./bin/vinom-matchmaking ./main.go

test:
	go test -v ./...

run: build
	@./bin/vinom-matchmaking

# Variables
PROTO_DIRS  = $(shell find . -name '*.proto' -exec dirname {} \; | sort -u) # Find unique directories containing .proto files

# Rule to generate Go code
genpb: 
	@echo "Generating Go code from .proto files..."
	@for dir in $(PROTO_DIRS); do \
		echo "Processing directory: $$dir"; \
		protoc -I $$dir --go_out=$$dir $$dir/*.proto \
								--go-grpc_out=$$dir; \
	done
	@echo "Protobuf generation complete!"
