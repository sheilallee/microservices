package api

import (
	"github.com/sheilallee/microservices/shipping/internal/application/core/domain"
	"github.com/sheilallee/microservices/shipping/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db ports.DBPort
}

func NewApplication(db ports.DBPort) *Application {
	return &Application{db: db}
}

func (a Application) CreateShipping(shipping domain.Shipping) (domain.Shipping, error) {
	shipping.DeliveryDays = shipping.CalculateDeliveryDays()

	err := a.db.Save(&shipping)
	if err != nil {
		return domain.Shipping{}, status.Errorf(codes.Internal, "failed to save shipping: %v", err)
	}
	return shipping, nil
}
