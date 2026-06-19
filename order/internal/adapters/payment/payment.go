package payment_adapter

import (
	"context"
	"log"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sheilallee/microservices-proto/golang/payment"
	"github.com/sheilallee/microservices/order/internal/application/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Adapter struct {
	payment payment.PaymentClient
}

func NewAdapter(paymentServiceURL string) (*Adapter, error) {
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
			grpc_retry.WithMax(5),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
		)))
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(paymentServiceURL, opts...)
	if err != nil {
		log.Printf("failed to connect to payment service at %s: %v", paymentServiceURL, err)
		return nil, err
	}
	client := payment.NewPaymentClient(conn)
	log.Printf("payment client initialized for %s", paymentServiceURL)
	return &Adapter{payment: client}, nil
}

func (a *Adapter) Charge(order domain.Order) error {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := a.payment.Create(ctx, &payment.CreatePaymentRequest{
		UserId:     order.CustomerID,
		OrderId:    order.ID,
		TotalPrice: order.TotalPrice(),
	})
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			log.Printf("payment request timeout for order_id=%d customer_id=%d", order.ID, order.CustomerID)
		}
		log.Printf("payment charge failed for order_id=%d customer_id=%d: %v", order.ID, order.CustomerID, err)
	}
	return err
}
