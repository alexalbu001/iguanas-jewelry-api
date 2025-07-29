package handlers

import (
	"github.com/alexalbu001/iguanas-jewelry/internal/responses"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
)

func convertToCartResponse(summary service.CartSummary) responses.CartResponse {
	var responseItems []responses.CartItemResponse
	for _, item := range summary.Items {
		responseItems = append(responseItems, responses.CartItemResponse{
			CartItemID:  item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	return responses.CartResponse{
		CartID: summary.CartID,
		Items:  responseItems,
		Total:  summary.Total,
	}
}
