package ports

import "github.com/sheilallee/microservices/shipping/internal/application/core/domain"

type APIPort interface {
	CreateShipping(shipping domain.Shipping) (domain.Shipping, error)
}
