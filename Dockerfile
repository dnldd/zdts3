# syntax=docker/dockerfile:1

## run with: docker build .
## tag with: docker tag image-sha image-name:tag

## Builder stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy only dependency files first to leverage Docker cache
ENV GOPROXY=https://goproxy.io,direct
COPY go.mod go.sum ./
RUN go mod download 

# Copy the rest of the source code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /archiver

## Final stage
FROM scratch

# Copy SSL certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary
COPY --from=builder /archiver /archiver

# Declare the archiver binary as the entrypoint
ENTRYPOINT ["/archiver"]
