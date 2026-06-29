# Build stage
# NOTE: Build context must be the project root (ProjetogRPC/) so that
# both microservices/ and microservices-proto/ are available.
FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copy proto stubs required by replace directives in go.mod
COPY microservices-proto/golang/payment ./microservices-proto/golang/payment

# Copy payment service source
COPY microservices/payment ./microservices/payment

WORKDIR /workspace/microservices/payment
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payment ./cmd/main.go

# Final image
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /workspace/microservices/payment/payment ./payment
CMD ["./payment"]
