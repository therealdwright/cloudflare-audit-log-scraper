# Stage 1: build the Go project
FROM golang:1.24 AS build

WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the module dependencies
RUN go mod download

# Copy the project source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o cloudflare-audit-log-scraper .

# Stage 2: create the final image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/cloudflare-audit-log-scraper .

# Run the binary
CMD ["./cloudflare-audit-log-scraper"]
