# Stage 1: Build
FROM golang:1-alpine3@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder
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
FROM alpine:3@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
RUN adduser -D worker

COPY --from=builder /src/worker /worker

USER worker
WORKDIR /

CMD ["/worker"]