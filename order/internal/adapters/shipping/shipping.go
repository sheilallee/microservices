package shipping_adapter

import (
	"context"
	"log"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sheilallee/microservices-proto/golang/shipping"
	"github.com/sheilallee/microservices/order/internal/application/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Adapter struct {
	shipping shipping.ShippingClient
}

func NewAdapter(shippingServiceURL string) (*Adapter, error) {
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
			grpc_retry.WithMax(5),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
		)))
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(shippingServiceURL, opts...)
	if err != nil {
		log.Printf("failed to connect to shipping service at %s: %v", shippingServiceURL, err)
		return nil, err
	}
	client := shipping.NewShippingClient(conn)
	log.Printf("shipping client initialized for %s", shippingServiceURL)
	return &Adapter{shipping: client}, nil
}

func (a *Adapter) Schedule(order domain.Order) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var items []*shipping.ShippingItem
	for _, item := range order.OrderItems {
		items = append(items, &shipping.ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}

	resp, err := a.shipping.Create(ctx, &shipping.CreateShippingRequest{
		OrderId: order.ID,
		Items:   items,
	})
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			log.Printf("shipping request timeout for order_id=%d", order.ID)
		}
		log.Printf("shipping schedule failed for order_id=%d: %v", order.ID, err)
		return 0, err
	}
	return resp.DeliveryDays, nil
}
