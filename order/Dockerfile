# Build stage
# NOTE: Build context must be the project root (ProjetogRPC/) so that
# both microservices/ and microservices-proto/ are available to resolve
# the local replace directives in go.mod.
FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copy proto stubs required by replace directives in go.mod
COPY microservices-proto/golang/order ./microservices-proto/golang/order
COPY microservices-proto/golang/payment ./microservices-proto/golang/payment
COPY microservices-proto/golang/shipping ./microservices-proto/golang/shipping

# Copy order service source
COPY microservices/order ./microservices/order

WORKDIR /workspace/microservices/order
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order ./cmd/main.go

# Final image
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /workspace/microservices/order/order ./order
CMD ["./order"]
