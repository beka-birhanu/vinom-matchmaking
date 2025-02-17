# Stage prod-stage
FROM golang:1.24.0 AS prod-stage 

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the application
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./main.go


# Command to run the application binary
ENTRYPOINT ["/api"]



# Stage dev-stage: Development environment with air
FROM golang:1.24.0 AS dev-stage

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y curl git && apt-get clean

# Install air
RUN go install github.com/air-verse/air@latest

# Copy the application files
COPY . .

# Command to run air for development
CMD ["air", "-c", ".air.toml"]
