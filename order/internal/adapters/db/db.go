package db

import (
	"fmt"

	"github.com/sheilallee/microservices/order/internal/application/core/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	CustomerID int64
	Status     string
	OrderItems []OrderItem
}

type OrderItem struct {
	gorm.Model
	ProductCode string
	UnitPrice   float32
	Quantity    int32
	OrderID     uint
}

// StockItem representa um produto existente na tabela de estoque.
type StockItem struct {
	gorm.Model
	ProductCode string `gorm:"uniqueIndex;not null;size:100"`
	Name        string
	UnitPrice   float32
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(dataSourceUrl string) (*Adapter, error) {
	db, openErr := gorm.Open(mysql.Open(dataSourceUrl), &gorm.Config{})
	if openErr != nil {
		return nil, fmt.Errorf("db connection error: %v", openErr)
	}
	err := db.AutoMigrate(&Order{}, &OrderItem{}, &StockItem{})
	if err != nil {
		return nil, fmt.Errorf("db migration error: %v", err)
	}
	return &Adapter{db: db}, nil
}

func (a Adapter) Get(id string) (domain.Order, error) {
	var orderEntity Order
	res := a.db.Preload("OrderItems").First(&orderEntity, id)
	var orderItems []domain.OrderItem
	for _, orderItem := range orderEntity.OrderItems {
		orderItems = append(orderItems, domain.OrderItem{
			ProductCode: orderItem.ProductCode,
			UnitPrice:   orderItem.UnitPrice,
			Quantity:    orderItem.Quantity,
		})
	}
	order := domain.Order{
		ID:         int64(orderEntity.ID),
		CustomerID: orderEntity.CustomerID,
		Status:     orderEntity.Status,
		OrderItems: orderItems,
		CreatedAt:  orderEntity.CreatedAt.UnixNano(),
	}
	return order, res.Error
}

func (a Adapter) Save(order *domain.Order) error {
	var orderItems []OrderItem
	for _, orderItem := range order.OrderItems {
		orderItems = append(orderItems, OrderItem{
			ProductCode: orderItem.ProductCode,
			UnitPrice:   orderItem.UnitPrice,
			Quantity:    orderItem.Quantity,
		})
	}
	orderModel := Order{
		CustomerID: order.CustomerID,
		Status:     order.Status,
		OrderItems: orderItems,
	}
	res := a.db.Create(&orderModel)
	if res.Error == nil {
		order.ID = int64(orderModel.ID)
	}
	return res.Error
}

func (a Adapter) Update(order *domain.Order) error {
	res := a.db.Model(&Order{}).Where("id = ?", order.ID).Update("status", order.Status)
	return res.Error
}

// GetStockItemsByCodes retorna os itens de estoque que correspondem aos códigos de produto fornecidos.
func (a Adapter) GetStockItemsByCodes(codes []string) ([]domain.StockItem, error) {
	var stockItems []StockItem
	res := a.db.Where("product_code IN ?", codes).Find(&stockItems)
	if res.Error != nil {
		return nil, res.Error
	}
	var result []domain.StockItem
	for _, si := range stockItems {
		result = append(result, domain.StockItem{
			ProductCode: si.ProductCode,
			Name:        si.Name,
			UnitPrice:   si.UnitPrice,
		})
	}
	return result, nil
}
