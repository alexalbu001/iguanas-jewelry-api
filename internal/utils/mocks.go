package utils

import (
	"context"
	"fmt"
	"sort"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TX

type MockTxManager struct{}

func (m *MockTxManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	// For tests, just execute the function without a real transaction
	return fn(nil)
}

// Products

type MockProductStore struct {
	Store []models.Product
}

func (m *MockProductStore) GetAll(ctx context.Context) ([]models.Product, error) {
	return m.Store, nil
}

func (m *MockProductStore) GetByID(ctx context.Context, id string) (models.Product, error) {

	for _, product := range m.Store {
		if id == product.ID {
			return product, nil
		}
	}
	return models.Product{}, fmt.Errorf("Product not found: %s", id)
}

func (m *MockProductStore) GetByIDBatch(ctx context.Context, productIDs []string) (map[string]models.Product, error) {
	productMap := make(map[string]models.Product) // create return map

	storeMap := make(map[string]models.Product) //create store map
	for _, product := range m.Store {
		storeMap[product.ID] = product
	}

	for _, productID := range productIDs {
		if product, exists := storeMap[productID]; exists {
			productMap[productID] = product
		}
	}
	return productMap, nil
}

func (m *MockProductStore) Add(ctx context.Context, product models.Product) (models.Product, error) {
	product.ID = uuid.NewString()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	m.Store = append(m.Store, product)
	return product, nil
}

func (m *MockProductStore) AddTx(ctx context.Context, product models.Product, tx pgx.Tx) (models.Product, error) {
	product.ID = uuid.NewString()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	m.Store = append(m.Store, product)
	return product, nil
}

func (m *MockProductStore) Update(ctx context.Context, id string, product models.Product) (models.Product, error) {
	for index, value := range m.Store {
		if id == value.ID {
			product.ID = id
			product.CreatedAt = value.CreatedAt
			product.UpdatedAt = time.Now()
			m.Store[index] = product
			return product, nil
		}
	}
	return models.Product{}, fmt.Errorf("ID: %s not found", id)
}
func (m *MockProductStore) Delete(ctx context.Context, id string) error {
	for i, value := range m.Store {
		if value.ID == id {
			m.Store = append(m.Store[:i], m.Store[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("ID: %s not found", id)
}

func (m *MockProductStore) DeleteTx(ctx context.Context, id string, tx pgx.Tx) error {
	for i, value := range m.Store {
		if value.ID == id {
			m.Store = append(m.Store[:i], m.Store[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("ID: %s not found", id)
}

func (m *MockProductStore) UpdateStock(ctx context.Context, productID string, stockChange int) error {
	for i := range m.Store {
		if productID == m.Store[i].ID {
			if newQuantity := m.Store[i].StockQuantity + stockChange; newQuantity >= 0 {
				m.Store[i].StockQuantity = newQuantity
				m.Store[i].UpdatedAt = time.Now()
			}
		}
	}
	return nil
}

func (m *MockProductStore) UpdateStockTx(ctx context.Context, productID string, stockChange int, tx pgx.Tx) error {
	for index, product := range m.Store {
		if productID == product.ID {
			if newQuantity := product.StockQuantity + stockChange; newQuantity >= 0 {
				m.Store[index].StockQuantity = newQuantity
				m.Store[index].UpdatedAt = time.Now()
				return nil
			}
		}
	}
	return fmt.Errorf("Product not found with id %s", productID)
}

func (m *MockProductStore) GetAllIncludingDeleted(ctx context.Context) ([]models.Product, error) {
	return m.Store, nil
}

func (m *MockProductStore) Restore(ctx context.Context, id string) error {
	// For mock purposes, just return nil as if restore was successful
	return nil
}

/////////////////////////////////////
// Carts

type MockCartsStore struct {
	CartsStore     []models.Cart
	CartItemsStore []models.CartItems
}

func (m *MockCartsStore) GetOrCreateCartByUserID(ctx context.Context, userID string) (models.Cart, error) {
	// for _, cart := range m.cartsStore {
	// 	if userID == cart.UserID {
	// 		return cart, nil
	// 	}
	// }
	cartsMap := make(map[string]models.Cart)
	for _, cart := range m.CartsStore {
		cartsMap[cart.UserID] = cart
	}
	if foundCart, exists := cartsMap[userID]; exists {
		return foundCart, nil
	}

	newCart := models.Cart{
		ID:        uuid.NewString(),
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.CartsStore = append(m.CartsStore, newCart)

	return newCart, nil
}

func (m *MockCartsStore) EmptyCart(ctx context.Context, userID string) error {
	var targetCartID string
	for _, cart := range m.CartsStore {
		if userID == cart.UserID {
			targetCartID = cart.ID
			break
		}
	}
	if targetCartID == "" {
		return nil
	}

	var newCartItems []models.CartItems
	for _, cartItem := range m.CartItemsStore {
		if cartItem.CartID != targetCartID {
			newCartItems = append(newCartItems, cartItem)
		}
	}

	m.CartItemsStore = newCartItems
	return nil
}

func (m *MockCartsStore) EmptyCartTx(ctx context.Context, userID string, tx pgx.Tx) error {
	var targetCartID string
	for _, cart := range m.CartsStore {
		if userID == cart.UserID {
			targetCartID = cart.ID
			break
		}
	}
	if targetCartID == "" {
		return nil
	}

	var newCartItems []models.CartItems
	for _, cartItem := range m.CartItemsStore {
		if cartItem.CartID != targetCartID {
			newCartItems = append(newCartItems, cartItem)
		}
	}

	m.CartItemsStore = newCartItems
	return nil
}

func (m *MockCartsStore) GetCartItemByID(ctx context.Context, id string) (models.CartItems, error) {
	cartItemsMap := make(map[string]models.CartItems)
	for _, cartItem := range m.CartItemsStore {
		cartItemsMap[cartItem.ID] = cartItem
	}

	if foundCartItem, exists := cartItemsMap[id]; exists {
		return foundCartItem, nil
	}
	return models.CartItems{}, &customerrors.ErrCartItemNotFound
}

func (m *MockCartsStore) AddItemToCart(ctx context.Context, cartID, productID string, quantity int) (models.CartItems, error) {
	for i, cartItem := range m.CartItemsStore {
		if cartItem.CartID == cartID && cartItem.ProductID == productID {
			m.CartItemsStore[i].Quantity += quantity
			m.CartItemsStore[i].CreatedAt = time.Now()
			return m.CartItemsStore[i], nil
		}
	}
	newCartItem := models.CartItems{
		ID:        uuid.NewString(),
		ProductID: productID,
		CartID:    cartID,
		Quantity:  quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.CartItemsStore = append(m.CartItemsStore, newCartItem)
	return newCartItem, nil
}

func (m *MockCartsStore) GetCartItems(ctx context.Context, cartID string) ([]models.CartItems, error) {
	var returnedCartItems []models.CartItems
	for _, cartItem := range m.CartItemsStore {
		if cartID == cartItem.CartID {
			returnedCartItems = append(returnedCartItems, cartItem)
		}

	}

	sort.Slice(returnedCartItems, func(i, j int) bool {
		return returnedCartItems[i].CreatedAt.Before(returnedCartItems[j].CreatedAt)
	})

	return returnedCartItems, nil
}

func (m *MockCartsStore) UpdateCartItemQuantity(ctx context.Context, itemID string, newQuantity int) error {
	for i, cartItem := range m.CartItemsStore {
		if itemID == cartItem.ID {
			m.CartItemsStore[i].Quantity = newQuantity
			m.CartItemsStore[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return &customerrors.ErrCartItemNotFound
}

func (m *MockCartsStore) DeleteCartItem(ctx context.Context, cartItemID string) error {
	var foundCartItemID bool

	for i, cartItem := range m.CartItemsStore {
		if cartItemID == cartItem.ID {
			m.CartItemsStore = append(m.CartItemsStore[:i], m.CartItemsStore[i+1:]...)
			foundCartItemID = true
		}

	}
	if !foundCartItemID {
		return &customerrors.ErrCartItemNotFound
	}
	return nil
}

func (m *MockCartsStore) GetCartItemByProductID(ctx context.Context, productID, cartID string) (models.CartItems, error) {

	for _, cartItem := range m.CartItemsStore {
		if productID == cartItem.ProductID && cartID == cartItem.CartID {
			return cartItem, nil
		}
	}
	return models.CartItems{}, &customerrors.ErrCartItemNotFound
}

//////////////////////////////////////////////
// Users

type MockUsersStore struct {
	Users []models.User
}

// Implement the one you need
func (m *MockUsersStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	for _, user := range m.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return models.User{}, fmt.Errorf("user not found: %s", id)
}

// Stub all the others
func (m *MockUsersStore) GetUsers(ctx context.Context) ([]models.User, error) {
	panic("GetUsers not implemented in mock")
}

func (m *MockUsersStore) GetUserByGoogleID(ctx context.Context, googleID string) (models.User, error) {
	panic("GetUserByGoogleID not implemented in mock")
}

func (m *MockUsersStore) AddUser(ctx context.Context, user models.User) (models.User, error) {
	panic("AddUser not implemented in mock")
}

func (m *MockUsersStore) DeleteUser(ctx context.Context, id string) error {
	panic("DeleteUser not implemented in mock")
}

func (m *MockUsersStore) UpdateUser(ctx context.Context, id string, user models.User) (models.User, error) {
	panic("UpdateUser not implemented in mock")
}

func (m *MockUsersStore) UpdateUserRole(ctx context.Context, id string, role string) error {
	panic("UpdateUserRole not implemented in mock")
}

//////////////////////////////////////////////
// Orders

type MockOrderStore struct {
	OrderStore     []models.Order
	OrderItemStore []models.OrderItem
}

func (m *MockOrderStore) InsertOrder(ctx context.Context, order models.Order) error {
	// order.ID = uuid.NewString()
	// order.Status = "pending"
	// order.CreatedAt = time.Now().UTC()
	// order.UpdatedAt = time.Now().UTC() This is done by service
	m.OrderStore = append(m.OrderStore, order)
	return nil
}

func (m *MockOrderStore) InsertOrderTx(ctx context.Context, order models.Order, tx pgx.Tx) error {
	// order.ID = uuid.NewString()
	// order.Status = "pending"
	// order.CreatedAt = time.Now().UTC()
	// order.UpdatedAt = time.Now().UTC()
	m.OrderStore = append(m.OrderStore, order)
	return nil
}

func (m *MockOrderStore) InsertOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	m.OrderItemStore = append(m.OrderItemStore, orderItem)
	return nil
}

func (m *MockOrderStore) InsertOrderItemTx(ctx context.Context, orderItem models.OrderItem, tx pgx.Tx) error {
	m.OrderItemStore = append(m.OrderItemStore, orderItem)
	return nil
}

func (m *MockOrderStore) InsertOrderItemBulk(ctx context.Context, orderItems []models.OrderItem) error {
	for _, orderItem := range orderItems {
		err := m.InsertOrderItem(ctx, orderItem)
		if err != nil {
			return fmt.Errorf("Order item in bulk could not be created, %w", err)
		}
	}
	return nil
}

func (m *MockOrderStore) InsertOrderItemBulkTx(ctx context.Context, orderItems []models.OrderItem, tx pgx.Tx) error {
	for _, orderItem := range orderItems {
		err := m.InsertOrderItem(ctx, orderItem)
		if err != nil {
			return fmt.Errorf("Order item in bulk could not be created, %w", err)
		}
	}
	return nil
}

func (m *MockOrderStore) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	for _, order := range m.OrderStore {
		if orderID == order.ID {
			return order, nil
		}
	}
	return models.Order{}, fmt.Errorf("Order not found with id : %s", orderID)
}

func (m *MockOrderStore) GetOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	var orderItems []models.OrderItem
	for _, orderItem := range m.OrderItemStore {
		if orderID == orderItem.OrderID {
			orderItems = append(orderItems, orderItem)
		}
	}

	return orderItems, nil
}

func (m *MockOrderStore) GetOrderItemsBatch(ctx context.Context, orderIDs []string) (map[string][]models.OrderItem, error) {
	newOrderItemStore := make(map[string][]models.OrderItem)
	orderItemsMap := make(map[string][]models.OrderItem)

	for _, orderItem := range m.OrderItemStore {
		newOrderItemStore[orderItem.OrderID] = append(newOrderItemStore[orderItem.OrderID], orderItem)
	}

	for _, orderID := range orderIDs {
		if orderItems, exists := newOrderItemStore[orderID]; exists {
			orderItemsMap[orderID] = append(orderItemsMap[orderID], orderItems...)
		}
	}
	return orderItemsMap, nil
}

func (m *MockOrderStore) GetUsersOrders(ctx context.Context, userID string) ([]models.Order, error) {
	var userOrders []models.Order
	for _, order := range m.OrderStore {
		if userID == order.UserID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders, nil
}

func (m *MockOrderStore) UpdateOrderStatus(ctx context.Context, status, orderID string) error {
	for i := range m.OrderStore {
		if orderID == m.OrderStore[i].ID {
			m.OrderStore[i].Status = status
			m.OrderStore[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return &customerrors.ErrOrderNotFound
}

func (m *MockOrderStore) UpdateOrderStatusTx(ctx context.Context, status, orderID string, tx pgx.Tx) error {
	for i := range m.OrderStore {
		if orderID == m.OrderStore[i].ID {
			m.OrderStore[i].Status = status
			m.OrderStore[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return &customerrors.ErrOrderNotFound
}

func (m *MockOrderStore) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	return m.OrderStore, nil
}

func (m *MockOrderStore) GetOrdersByStatus(ctx context.Context, status string) ([]models.Order, error) {
	var returnedOrders []models.Order
	for _, order := range m.OrderStore {
		if order.Status == status {
			returnedOrders = append(returnedOrders, order)
		}
	}
	return returnedOrders, nil
}

//////////////////////////
////Favorites

type MockUserFavoritesStore struct {
	UserFavorites []models.UserFavorites
}

func (m *MockUserFavoritesStore) GetUserFavorites(ctx context.Context, userID string) ([]models.UserFavorites, error) {
	var returnedFavorites []models.UserFavorites
	for _, userFavorite := range m.UserFavorites {
		if userID == userFavorite.UserID {
			returnedFavorites = append(returnedFavorites, userFavorite)
		}
	}

	return returnedFavorites, nil
}

func (m *MockUserFavoritesStore) AddUserFavorite(ctx context.Context, userID string, productID string) error {
	newFavorite := models.UserFavorites{
		ID:        uuid.NewString(),
		UserID:    userID,
		ProductID: productID,
		CreatedAt: time.Now(),
	}
	m.UserFavorites = append(m.UserFavorites, newFavorite)
	return nil
}

func (m *MockUserFavoritesStore) RemoveUserFavorite(ctx context.Context, userID string, productID string) error {

	for i, userFavorite := range m.UserFavorites {
		if userFavorite.UserID == userID && userFavorite.ProductID == productID {
			m.UserFavorites = append(m.UserFavorites[:i], m.UserFavorites[i+1:]...)
		}
	}
	return nil
}

func (m *MockUserFavoritesStore) ClearUserFavorites(ctx context.Context, userID string) error {
	for i, userFavorite := range m.UserFavorites {
		if userFavorite.UserID == userID {
			m.UserFavorites = append(m.UserFavorites[:i], m.UserFavorites[i+1:]...)
		}
	}
	return nil
}
