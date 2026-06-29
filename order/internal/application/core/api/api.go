package api

import (
	"log"

	"github.com/sheilallee/microservices/order/internal/application/core/domain"
	"github.com/sheilallee/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db       ports.DBPort
	payment  ports.PaymentPort
	shipping ports.ShippingPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort, shipping ports.ShippingPort) *Application {
	return &Application{
		db:       db,
		payment:  payment,
		shipping: shipping,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	// Valida se o total de itens não excede 50 
	var totalItems int32
	for _, item := range order.OrderItems {
		totalItems += item.Quantity
	}
	if totalItems > 50 {
		return domain.Order{}, status.Errorf(codes.InvalidArgument, "Order cannot have more than 50 items in total.")
	}

	// Valida se todos os códigos de produto existem no estoque
	productCodes := make([]string, 0, len(order.OrderItems))
	for _, item := range order.OrderItems {
		productCodes = append(productCodes, item.ProductCode)
	}
	stockItems, err := a.db.GetStockItemsByCodes(productCodes)
	if err != nil {
		return domain.Order{}, status.Errorf(codes.Internal, "failed to verify stock: %v", err)
	}
	stockMap := make(map[string]struct{}, len(stockItems))
	for _, si := range stockItems {
		stockMap[si.ProductCode] = struct{}{}
	}
	for _, code := range productCodes {
		if _, found := stockMap[code]; !found {
			return domain.Order{}, status.Errorf(codes.NotFound, "product_code '%s' does not exist in stock", code)
		}
	}

	// Salva o pedido no banco de dados
	if saveErr := a.db.Save(&order); saveErr != nil {
		return domain.Order{}, status.Errorf(codes.Internal, "failed to save order: %v", saveErr)
	}

	// Realiza o pagamento
	if paymentErr := a.payment.Charge(order); paymentErr != nil {
		order.Status = "Canceled"
		_ = a.db.Update(&order)
		return domain.Order{}, paymentErr
	}

	// Atualiza o status do pedido para Pago
	order.Status = "Paid"
	if updateErr := a.db.Update(&order); updateErr != nil {
		return domain.Order{}, status.Errorf(codes.Internal, "failed to update order status: %v", updateErr)
	}

	// Agenda o envio somente após o pagamento bem-sucedido
	deliveryDays, shippingErr := a.shipping.Schedule(order)
	if shippingErr != nil {
		// Pagamento já realizado — registra e retorna erro para o cliente
		log.Printf("shipping schedule failed for order_id=%d: %v", order.ID, shippingErr)
		return domain.Order{}, status.Errorf(codes.Internal, "payment succeeded but shipping scheduling failed: %v", shippingErr)
	}

	log.Printf("shipping scheduled for order_id=%d, delivery_days=%d", order.ID, deliveryDays)
	return order, nil
}
