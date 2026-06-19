package api

import (
	"github.com/sheilallee/microservices/order/internal/application/core/domain"
	"github.com/sheilallee/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db      ports.DBPort
	payment ports.PaymentPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort) *Application {
	return &Application{
		db:      db,
		payment: payment,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	// Validate total items
	var totalItems int32
	for _, item := range order.OrderItems {
		totalItems += item.Quantity
	}
	if totalItems > 50 {
		return domain.Order{}, status.Errorf(codes.InvalidArgument, "Order cannot have more than 50 items in total.")
	}

	err := a.db.Save(&order)
	if err != nil {
		return domain.Order{}, status.Errorf(codes.Internal, "failed to save order: %v", err)
	}

	paymentErr := a.payment.Charge(order)
	if paymentErr != nil {
		// Update order status to "Canceled" due to payment error
		order.Status = "Canceled"
		a.db.Update(&order)
		return domain.Order{}, paymentErr
	}

	// Update order status to "Paid" after successful payment
	order.Status = "Paid"
	err = a.db.Update(&order)
	if err != nil {
		return domain.Order{}, status.Errorf(codes.Internal, "failed to update order status: %v", err)
	}

	return order, nil
}
