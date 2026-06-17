# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go.mod first (go.sum will be generated if absent)
COPY go.mod ./
RUN go mod download

COPY . .
# -mod=mod allows go to update go.sum inside the container when needed
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -ldflags="-w -s" -o server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["/app/server"]
