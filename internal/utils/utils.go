package utils

import (
	"fmt"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
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
	// Return empty slice instead of error when no order items exist
	if len(orderItems) == 0 {
		return []string{}, nil
	}

	var productIDs []string
	for _, orderItem := range orderItems {
		productIDs = append(productIDs, orderItem.ProductID)
	}
	return productIDs, nil
}

func ExtractOrderIDs(orders []models.Order) ([]string, error) {
	// Return empty slice instead of error when no orders exist
	if len(orders) == 0 {
		return []string{}, nil
	}

	var orderIDs []string
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}
	return orderIDs, nil
}
