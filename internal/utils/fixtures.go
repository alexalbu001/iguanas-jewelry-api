package utils

import (
	"time"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
)

// Shared test IDs that multiple tests can reference
var KnownUserID = uuid.New().String()
var KnownCartID = uuid.New().String()
var KnownCartItemID = uuid.New().String()
var SecondUserID = uuid.New().String()
var SecondCartID = uuid.New().String()
var AdminUserID = uuid.New().String()
var KnownProductID = uuid.New().String()

// Order IDs
var KnownOrderID = uuid.New().String()
var KnownOrderID2 = uuid.New().String()
var SecondOrderID = uuid.New().String()
var AdminOrderID = uuid.New().String()

// Favorite IDs
var KnownFavoriteId = uuid.New().String()
var KnownFavoriteId2 = uuid.New().String()

// Product IDs - keeping them consistent across all fixtures
const (
	GoldRingID        = "prod-gold-ring-001"
	DiamondEarringsID = "prod-diamond-earrings-002"
	PearlNecklaceID   = "prod-pearl-necklace-003"
	SilverBraceletID  = "prod-silver-bracelet-004"
	SapphireRingID    = "prod-sapphire-ring-005"
	RubyEarringsID    = "prod-ruby-earrings-006"
	EmeraldNecklaceID = "prod-emerald-necklace-007"
)

// CreateJewelryProducts returns a complete set of jewelry products for testing
func CreateJewelryProducts() []models.Product {
	baseTime := time.Now()

	return []models.Product{
		{
			ID:            GoldRingID,
			Name:          "Classic Gold Wedding Ring",
			Price:         899.99,
			Description:   "18k gold classic wedding band with polished finish",
			Category:      "rings",
			StockQuantity: 15,
			CreatedAt:     baseTime.Add(-30 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-5 * 24 * time.Hour),
		},
		{
			ID:            DiamondEarringsID,
			Name:          "Diamond Stud Earrings",
			Price:         1299.50,
			Description:   "1 carat total weight diamond studs in 14k white gold",
			Category:      "earrings",
			StockQuantity: 8,
			CreatedAt:     baseTime.Add(-25 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-3 * 24 * time.Hour),
		},
		{
			ID:            PearlNecklaceID,
			Name:          "Cultured Pearl Necklace",
			Price:         650.00,
			Description:   "18-inch strand of AAA cultured freshwater pearls",
			Category:      "necklaces",
			StockQuantity: 12,
			CreatedAt:     baseTime.Add(-20 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-2 * 24 * time.Hour),
		},
		{
			ID:            SilverBraceletID,
			Name:          "Sterling Silver Chain Bracelet",
			Price:         245.75,
			Description:   "Elegant sterling silver link bracelet with lobster clasp",
			Category:      "bracelets",
			StockQuantity: 20,
			CreatedAt:     baseTime.Add(-15 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-1 * 24 * time.Hour),
		},
		{
			ID:            SapphireRingID,
			Name:          "Blue Sapphire Engagement Ring",
			Price:         2150.00,
			Description:   "2 carat blue sapphire with diamond accents in platinum",
			Category:      "rings",
			StockQuantity: 3, // Low stock for testing edge cases
			CreatedAt:     baseTime.Add(-10 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-6 * time.Hour),
		},
		{
			ID:            RubyEarringsID,
			Name:          "Ruby Drop Earrings",
			Price:         875.25,
			Description:   "Elegant ruby drop earrings with gold accents",
			Category:      "earrings",
			StockQuantity: 6,
			CreatedAt:     baseTime.Add(-8 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-4 * time.Hour),
		},
		{
			ID:            EmeraldNecklaceID,
			Name:          "Emerald Tennis Necklace",
			Price:         3200.00,
			Description:   "Premium emerald tennis necklace in 18k gold setting",
			Category:      "necklaces",
			StockQuantity: 2, // Very low stock for testing
			CreatedAt:     baseTime.Add(-5 * 24 * time.Hour),
			UpdatedAt:     baseTime.Add(-2 * time.Hour),
		},
	}
}

// CreateTestUsers returns a set of users for testing different scenarios
func CreateTestUsers() []models.User {
	baseTime := time.Now()

	return []models.User{
		{
			ID:        KnownUserID,
			GoogleID:  "google-auth-123456789",
			Email:     "sarah.johnson@example.com",
			Name:      "Sarah Johnson",
			Role:      "customer",
			CreatedAt: baseTime.Add(-60 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-7 * 24 * time.Hour),
		},
		{
			ID:        SecondUserID,
			GoogleID:  "google-auth-987654321",
			Email:     "michael.chen@example.com",
			Name:      "Michael Chen",
			Role:      "customer",
			CreatedAt: baseTime.Add(-45 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-2 * 24 * time.Hour),
		},
		{
			ID:        AdminUserID,
			GoogleID:  "google-auth-admin-001",
			Email:     "admin@iguanas-jewelry.com",
			Name:      "Emma Rodriguez",
			Role:      "admin",
			CreatedAt: baseTime.Add(-90 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-1 * 24 * time.Hour),
		},
	}
}

// CreateTestCarts returns carts for different test scenarios
func CreateTestCarts() []models.Cart {
	baseTime := time.Now()

	return []models.Cart{
		{
			ID:        KnownCartID,
			UserID:    KnownUserID,
			CreatedAt: baseTime.Add(-7 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-2 * time.Hour),
		},
		{
			ID:        SecondCartID,
			UserID:    SecondUserID,
			CreatedAt: baseTime.Add(-3 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-1 * 24 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			UserID:    AdminUserID,
			CreatedAt: baseTime.Add(-1 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-30 * time.Minute),
		},
	}
}

// CreateTestCartItems returns cart items that match the products and carts
func CreateTestCartItems() []models.CartItems {
	baseTime := time.Now()

	return []models.CartItems{
		// Items for KnownUserID's cart
		{
			ID:        KnownCartItemID,
			ProductID: GoldRingID, // $899.99 x 2 = $1799.98
			CartID:    KnownCartID,
			Quantity:  2,
			CreatedAt: baseTime.Add(-6 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-3 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			ProductID: DiamondEarringsID, // $1299.50 x 1 = $1299.50
			CartID:    KnownCartID,
			Quantity:  1,
			CreatedAt: baseTime.Add(-5 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-5 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			ProductID: PearlNecklaceID, // $650.00 x 1 = $650.00
			CartID:    KnownCartID,
			Quantity:  1,
			CreatedAt: baseTime.Add(-4 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-4 * time.Hour),
		},
		// Total for KnownCartID: $3749.48

		// Items for SecondUserID's cart
		{
			ID:        uuid.New().String(),
			ProductID: SilverBraceletID, // $245.75 x 3 = $737.25
			CartID:    SecondCartID,
			Quantity:  3,
			CreatedAt: baseTime.Add(-2 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-1 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			ProductID: RubyEarringsID, // $875.25 x 1 = $875.25
			CartID:    SecondCartID,
			Quantity:  1,
			CreatedAt: baseTime.Add(-1 * 24 * time.Hour),
			UpdatedAt: baseTime.Add(-30 * time.Minute),
		},
		// Total for SecondCartID: $1612.50
	}
}

func CreateTestFavorites() []models.UserFavorites {
	baseTime := time.Now()

	return []models.UserFavorites{
		{
			ID:        KnownFavoriteId,
			UserID:    KnownUserID,
			ProductID: GoldRingID,
			CreatedAt: baseTime.Add(-1 * time.Hour),
		},
		{
			ID:        uuid.NewString(),
			UserID:    KnownUserID,
			ProductID: SilverBraceletID,
			CreatedAt: baseTime.Add(-5 * time.Hour),
		},
		{
			ID:        KnownFavoriteId2,
			UserID:    AdminUserID,
			ProductID: SapphireRingID,
			CreatedAt: baseTime.Add(-10 * time.Hour),
		},
	}
}

// CreateTestOrders returns orders for different test scenarios
func CreateTestOrders() []models.Order {
	baseTime := time.Now()

	return []models.Order{
		{
			ID:          KnownOrderID,
			UserID:      KnownUserID,
			TotalAmount: 3749.48, // Same as cart total - order created from cart
			Status:      "pending",

			// Shipping information
			ShippingName:         "Sarah Johnson",
			ShippingEmail:        "sarah.johnson@example.com",
			ShippingPhone:        "+1-555-0123",
			ShippingAddressLine1: "123 Jewelry Lane",
			ShippingAddressLine2: "Apt 4B",
			ShippingCity:         "New York",
			ShippingState:        "NY",
			ShippingPostalCode:   "10001",
			ShippingCountry:      "USA",

			CreatedAt: baseTime.Add(-24 * time.Hour), // 1 day ago
			UpdatedAt: baseTime.Add(-23 * time.Hour), // Updated 1 hour after creation
		},
		{
			ID:          SecondOrderID,
			UserID:      SecondUserID,
			TotalAmount: 1612.50, // From second user's cart
			Status:      "paid",

			// Shipping information
			ShippingName:         "Michael Chen",
			ShippingEmail:        "michael.chen@example.com",
			ShippingPhone:        "+1-555-0456",
			ShippingAddressLine1: "456 Diamond Street",
			ShippingAddressLine2: "",
			ShippingCity:         "Los Angeles",
			ShippingState:        "CA",
			ShippingPostalCode:   "90210",
			ShippingCountry:      "USA",

			CreatedAt: baseTime.Add(-48 * time.Hour), // 2 days ago
			UpdatedAt: baseTime.Add(-36 * time.Hour), // Updated 12 hours later
		},
		{
			ID:          AdminOrderID,
			UserID:      AdminUserID,
			TotalAmount: 5350.00, // High-value admin order
			Status:      "delivered",

			// Shipping information
			ShippingName:         "Emma Rodriguez",
			ShippingEmail:        "admin@iguanas-jewelry.com",
			ShippingPhone:        "+1-555-0789",
			ShippingAddressLine1: "789 Executive Plaza",
			ShippingAddressLine2: "Suite 100",
			ShippingCity:         "Miami",
			ShippingState:        "FL",
			ShippingPostalCode:   "33101",
			ShippingCountry:      "USA",

			CreatedAt: baseTime.Add(-72 * time.Hour), // 3 days ago
			UpdatedAt: baseTime.Add(-12 * time.Hour), // Recently updated
		},
		{
			ID:          KnownOrderID2,
			UserID:      KnownUserID,
			TotalAmount: 500.48, // Same as cart total - order created from cart
			Status:      "delivered",

			// Shipping information
			ShippingName:         "Sarah Johnson",
			ShippingEmail:        "sarah.johnson@example.com",
			ShippingPhone:        "+1-555-0123",
			ShippingAddressLine1: "123 Jewelry Lane",
			ShippingAddressLine2: "Apt 4B",
			ShippingCity:         "New York",
			ShippingState:        "NY",
			ShippingPostalCode:   "10001",
			ShippingCountry:      "USA",

			CreatedAt: baseTime.Add(-48 * time.Hour), // 1 day ago
			UpdatedAt: baseTime.Add(-47 * time.Hour), // Updated 1 hour after creation
		},
	}
}

// CreateTestOrderItems returns order items that match the orders and products
func CreateTestOrderItems() []models.OrderItem {
	baseTime := time.Now()

	return []models.OrderItem{
		// Items for KnownUserID's order (matches their cart)
		{
			ID:        uuid.New().String(),
			OrderID:   KnownOrderID,
			ProductID: GoldRingID, // Gold Ring
			Quantity:  2,
			Price:     899.99, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-24 * time.Hour),
			UpdatedAt: baseTime.Add(-24 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			OrderID:   KnownOrderID,
			ProductID: DiamondEarringsID,
			Quantity:  1,
			Price:     1299.50, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-24 * time.Hour),
			UpdatedAt: baseTime.Add(-24 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			OrderID:   KnownOrderID,
			ProductID: PearlNecklaceID,
			Quantity:  1,
			Price:     650.00, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-24 * time.Hour),
			UpdatedAt: baseTime.Add(-24 * time.Hour),
		},
		// Total for KnownOrderID: $3749.48

		// Items for SecondUserID's order
		{
			ID:        uuid.New().String(),
			OrderID:   SecondOrderID,
			ProductID: SilverBraceletID,
			Quantity:  3,
			Price:     245.75, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-48 * time.Hour),
			UpdatedAt: baseTime.Add(-48 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			OrderID:   SecondOrderID,
			ProductID: RubyEarringsID,
			Quantity:  1,
			Price:     875.25, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-48 * time.Hour),
			UpdatedAt: baseTime.Add(-48 * time.Hour),
		},
		// Total for SecondOrderID: $1612.50

		// Items for AdminUserID's order (high-value items)
		{
			ID:        uuid.New().String(),
			OrderID:   AdminOrderID,
			ProductID: SapphireRingID,
			Quantity:  1,
			Price:     2150.00, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-72 * time.Hour),
			UpdatedAt: baseTime.Add(-72 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			OrderID:   AdminOrderID,
			ProductID: EmeraldNecklaceID,
			Quantity:  1,
			Price:     3200.00, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-72 * time.Hour),
			UpdatedAt: baseTime.Add(-72 * time.Hour),
		},
		// Total for AdminOrderID: $5350.00
		{
			ID:        uuid.New().String(),
			OrderID:   KnownOrderID2,
			ProductID: EmeraldNecklaceID,
			Quantity:  1,
			Price:     3200.00, // Price snapshot at time of purchase
			CreatedAt: baseTime.Add(-72 * time.Hour),
			UpdatedAt: baseTime.Add(-72 * time.Hour),
		},
	}
}
