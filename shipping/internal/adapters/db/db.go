package db

import (
	"fmt"

	"github.com/sheilallee/microservices/shipping/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Shipping struct {
	gorm.Model
	OrderID      int64
	DeliveryDays int32
	Items        []ShippingItem
}

type ShippingItem struct {
	gorm.Model
	ProductCode string
	Quantity    int32
	ShippingID  uint
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceUrl string) (*Adapter, error) {
	db, openErr := gorm.Open(mysql.Open(dataSourceUrl), &gorm.Config{})
	if openErr != nil {
		return nil, fmt.Errorf("db connection error: %v", openErr)
	}
	err := db.AutoMigrate(&Shipping{}, &ShippingItem{})
	if err != nil {
		return nil, fmt.Errorf("db migration error: %v", err)
	}
	return &Adapter{db: db}, nil
}

func (a Adapter) Get(id string) (domain.Shipping, error) {
	var shippingEntity Shipping
	res := a.db.Preload("Items").First(&shippingEntity, id)
	if res.Error != nil {
		return domain.Shipping{}, res.Error
	}
	var items []domain.ShippingItem
	for _, item := range shippingEntity.Items {
		items = append(items, domain.ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}
	return domain.Shipping{
		ID:           int64(shippingEntity.ID),
		OrderID:      shippingEntity.OrderID,
		DeliveryDays: shippingEntity.DeliveryDays,
		Items:        items,
		CreatedAt:    shippingEntity.CreatedAt.UnixNano(),
	}, nil
}

func (a Adapter) Save(shipping *domain.Shipping) error {
	var items []ShippingItem
	for _, item := range shipping.Items {
		items = append(items, ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}
	shippingModel := Shipping{
		OrderID:      shipping.OrderID,
		DeliveryDays: shipping.DeliveryDays,
		Items:        items,
	}
	res := a.db.Create(&shippingModel)
	if res.Error == nil {
		shipping.ID = int64(shippingModel.ID)
	}
	return res.Error
}
