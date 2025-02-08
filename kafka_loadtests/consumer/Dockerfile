# Step 1: Builder stage
FROM --platform=linux/amd64 golang:1.23 AS builder

# Install required tools for building
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc-multilib \
    g++-multilib \
    libc6-dev \
    ca-certificates \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application for Linux amd64 architecture
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main .

# Step 2: Minimal runtime stage
FROM --platform=linux/amd64 debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates  && apt-get install -y librdkafka1 && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]


