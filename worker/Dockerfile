# Stage 1: Build
#v1.25-alpine
FROM golang@sha256:ac09a5f469f307e5da71e766b0bd59c9c49ea460a528cc3e6686513d64a6f1fb AS builder
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
#v3.23.2
FROM alpine@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62
RUN adduser -D worker

COPY --from=builder /src/worker /worker

USER worker
WORKDIR /

CMD ["/worker"]