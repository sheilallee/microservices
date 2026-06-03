package api

import (
	"context"

	"github.com/sheilallee/microservices/payment/internal/application/core/domain"
	"github.com/sheilallee/microservices/payment/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db ports.DBPort
}

func NewApplication(db ports.DBPort) *Application {
	return &Application{
		db: db,
	}
}

func (a Application) Charge(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	err := a.db.Save(ctx, &payment)
	if err != nil {
		return domain.Payment{}, status.Errorf(codes.Internal, "failed to save payment: %v", err)
	}
	return payment, nil
}
