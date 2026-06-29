package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/sheilallee/microservices-proto/golang/shipping"
	"github.com/sheilallee/microservices/shipping/config"
	"github.com/sheilallee/microservices/shipping/internal/application/core/domain"
	"github.com/sheilallee/microservices/shipping/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Adapter struct {
	api  ports.APIPort
	port int
	shipping.UnimplementedShippingServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

func (a Adapter) Create(ctx context.Context, request *shipping.CreateShippingRequest) (*shipping.CreateShippingResponse, error) {
	var items []domain.ShippingItem
	for _, item := range request.Items {
		items = append(items, domain.ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}

	newShipping := domain.NewShipping(request.OrderId, items)
	result, err := a.api.CreateShipping(newShipping)
	if err != nil {
		return nil, err
	}

	return &shipping.CreateShippingResponse{
		OrderId:      result.OrderID,
		DeliveryDays: result.DeliveryDays,
	}, nil
}

func (a Adapter) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}

	grpcServer := grpc.NewServer()
	shipping.RegisterShippingServer(grpcServer, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcServer)
	}

	log.Printf("starting shipping service on port %d ...", a.port)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve grpc on port %d", a.port)
	}
}
