package domain

import "time"

type ShippingItem struct {
	ProductCode string `json:"product_code"`
	Quantity    int32  `json:"quantity"`
}

type Shipping struct {
	ID           int64          `json:"id"`
	OrderID      int64          `json:"order_id"`
	DeliveryDays int32          `json:"delivery_days"`
	Items        []ShippingItem `json:"items"`
	CreatedAt    int64          `json:"created_at"`
}

func NewShipping(orderID int64, items []ShippingItem) Shipping {
	return Shipping{
		OrderID:   orderID,
		Items:     items,
		CreatedAt: time.Now().Unix(),
	}
}

func (s *Shipping) CalculateDeliveryDays() int32 {
	var totalUnits int32
	for _, item := range s.Items {
		totalUnits += item.Quantity
	}
	return 1 + totalUnits/5
}
