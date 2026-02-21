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
FROM alpine:3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN adduser -D worker

COPY --from=builder /src/worker /worker

USER worker
WORKDIR /

CMD ["/worker"]