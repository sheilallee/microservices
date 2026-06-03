package payment_adapter

import (
	"context"
	"log"

	"github.com/sheilallee/microservices-proto/golang/payment"
	"github.com/sheilallee/microservices/order/internal/application/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Adapter struct {
	payment payment.PaymentClient
}

func NewAdapter(paymentServiceURL string) (*Adapter, error) {
	var opts []grpc.DialOption
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

func (a *Adapter) Charge(order *domain.Order) error {
	_, err := a.payment.Create(context.Background(), &payment.CreatePaymentRequest{
		UserId:     order.CustomerID,
		OrderId:    order.ID,
		TotalPrice: order.TotalPrice(),
	})
	if err != nil {
		log.Printf("payment charge failed for order_id=%d customer_id=%d: %v", order.ID, order.CustomerID, err)
	}
	return err
}
