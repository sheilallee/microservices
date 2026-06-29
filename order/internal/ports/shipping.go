package ports

import "github.com/sheilallee/microservices/order/internal/application/core/domain"

type ShippingPort interface {
	Schedule(order domain.Order) (int32, error)
}
