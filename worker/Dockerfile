# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git

WORKDIR /src

# Copy go.mod and go.sum for the worker module
COPY src/go.mod .
COPY src/go.sum .

# Download dependencies
RUN go mod download

# Copy worker source
COPY src/ .

# Build binary
RUN go build -o worker .

# Stage 2: Minimal image
FROM alpine:3.18
RUN adduser -D worker

COPY --from=builder /src/worker /worker

USER worker
WORKDIR /

CMD ["/worker"]