FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copia os stubs proto necessários pelas diretivas replace do go.mod
COPY microservices-proto/golang/shipping ./microservices-proto/golang/shipping

# Copia o código-fonte do serviço Shipping
COPY microservices/shipping ./microservices/shipping

WORKDIR /workspace/microservices/shipping
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shipping ./cmd/main.go

# Imagem final
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /workspace/microservices/shipping/shipping ./shipping
CMD ["./shipping"]
