FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build optimized binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api /app/api
# Create uploads directory
RUN mkdir -p /app/uploads
EXPOSE 8000
CMD ["/app/api"]
