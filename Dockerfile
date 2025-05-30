FROM golang:alpine AS builder

# Install git and necessary packages
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application statically
RUN CGO_ENABLED=0 GOOS=linux go build -o prometheus-fe2-exporter .

# Stage 2: Runtime
FROM scratch 

# Use a non-root user
USER 1000:1000

# Copy the statically linked binary
COPY --from=builder /app/prometheus-fe2-exporter /

# Expose the metrics port
EXPOSE 9865

# Start the exporter
ENTRYPOINT ["/prometheus-fe2-exporter"]
