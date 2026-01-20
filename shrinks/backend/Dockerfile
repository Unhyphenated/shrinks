FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary named 'server' from the cmd directory
RUN go build -o server ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .
# Copy the SQL schema so the app can find it if needed
# COPY --from=builder /app/internal/storage/migrations ./migrations

# Command to run the executable
CMD ["./server"]