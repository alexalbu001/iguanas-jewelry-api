package utils

import (
	"fmt"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
)

func ExtractProductIDs(cartItems []models.CartItems) ([]string, error) {
	if len(cartItems) == 0 {
		return nil, fmt.Errorf("No cart items")
	}
	var productIDs []string
	for _, item := range cartItems {
		productIDs = append(productIDs, item.ProductID)
	}
	return productIDs, nil
}

func ExtractProductIDsFromOrderItems(orderItems []models.OrderItem) ([]string, error) {
	if len(orderItems) == 0 {
		return nil, fmt.Errorf("No order items")
	}

	var productIDs []string
	for _, orderItem := range orderItems {
		productIDs = append(productIDs, orderItem.ProductID)
	}
	return productIDs, nil
}

func ExtractOrderIDs(orders []models.Order) ([]string, error) {
	if len(orders) == 0 {
		return nil, fmt.Errorf("No orders available")
	}

	var orderIDs []string
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}
	return orderIDs, nil
}
