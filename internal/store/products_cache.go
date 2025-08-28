package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type CachedProductsStore struct {
	store       service.ProductsStore
	redisClient *redis.Client
	ttl         time.Duration
}

func NewCachedProductsStore(store service.ProductsStore, redisClient *redis.Client) *CachedProductsStore {
	return &CachedProductsStore{
		store:       store,
		redisClient: redisClient,
		ttl:         60 * time.Minute,
	}
}

func (c *CachedProductsStore) GetAll(ctx context.Context) ([]models.Product, error) {
	cacheKey := "products:all"

	cached, err := c.redisClient.Get(ctx, cacheKey).Result()

	if err == nil {
		var products []models.Product
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
		// if fail continue to pull info from db
	}

	// cache miss
	products, err := c.store.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		ctx2 := context.Background() // dont reuse context in a routine that outlives the request
		if data, err := json.Marshal(products); err == nil {
			c.redisClient.Set(ctx2, cacheKey, data, c.ttl).Err()
		}
	}()
	return products, nil
}

func (c *CachedProductsStore) Add(ctx context.Context, product models.Product) (models.Product, error) {
	result, err := c.store.Add(ctx, product)
	if err != nil {
		return models.Product{}, err
	}

	// Invalidate related cache entries
	c.redisClient.Del(ctx, "products:all").Err() // Remove the list
	// Don't check error - if Redis fails, still want to return success

	return result, nil
}

func (c *CachedProductsStore) Update(ctx context.Context, id string, product models.Product) (models.Product, error) {
	result, err := c.store.Update(ctx, id, product)
	if err != nil {
		return result, err
	}

	// Invalidate both the list and the individual product
	c.redisClient.Del(ctx,
		"products:all",
		"product:"+id,
	).Err()

	go func() {
		ctx2 := context.Background() // dont reuse context in a routine that outlives the request

		if data, err := json.Marshal(result); err == nil {
			c.redisClient.Set(ctx2, "product:"+id, data, c.ttl)
		}
	}()
	return result, nil
}

func (c *CachedProductsStore) GetByID(ctx context.Context, id string) (models.Product, error) {
	cached, err := c.redisClient.Get(ctx, "product:"+id).Result()
	if err == nil {
		var product models.Product
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			return product, nil
		}
	}

	product, err := c.store.GetByID(ctx, id)
	if err != nil {
		return models.Product{}, err
	}
	go func() {
		ctx2 := context.Background() // dont reuse context in a routine that outlives the request

		if data, err := json.Marshal(product); err == nil {
			c.redisClient.Set(ctx2, "product:"+id, data, c.ttl)
		}
	}()
	return product, nil
}

func (c *CachedProductsStore) GetByIDBatch(ctx context.Context, productIDs []string) (map[string]models.Product, error) {
	// Check individual product caches
	productsMap := make(map[string]models.Product)
	var missingIDs []string

	for _, id := range productIDs {
		cached, err := c.redisClient.Get(ctx, "product:"+id).Result()
		if err == nil {
			var product models.Product
			if json.Unmarshal([]byte(cached), &product) == nil {
				productsMap[product.ID] = product
				continue
			}
		}
		missingIDs = append(missingIDs, id)
	}

	// Fetch only missing products from database
	if len(missingIDs) > 0 {
		missingProductsMap, err := c.store.GetByIDBatch(ctx, missingIDs)
		if err != nil {
			return nil, err
		}

		for id, product := range missingProductsMap { //how to merge 2 maps
			productsMap[id] = product

			go func(product models.Product) {
				ctx2 := context.Background() // dont reuse context in a routine that outlives the request
				if data, err := json.Marshal(product); err == nil {
					c.redisClient.Set(ctx2, "product:"+id, data, c.ttl).Err() // cache only the missing ones
				}
			}(product)
		}

	}
	return productsMap, nil
} // to be fixed

func (c *CachedProductsStore) AddTx(ctx context.Context, product models.Product, tx pgx.Tx) (models.Product, error) {
	return c.store.AddTx(ctx, product, tx)
}

func (c *CachedProductsStore) Delete(ctx context.Context, id string) error {
	err := c.store.Delete(ctx, id)
	if err != nil {
		return err
	}
	c.redisClient.Del(ctx, "products:all", "product:"+id).Err()
	return nil
}

func (c *CachedProductsStore) DeleteTx(ctx context.Context, id string, tx pgx.Tx) error {
	return c.store.DeleteTx(ctx, id, tx)
}

func (c *CachedProductsStore) UpdateStock(ctx context.Context, productID string, stockChange int) error {
	err := c.store.UpdateStock(ctx, productID, stockChange)
	if err != nil {
		return err
	}

	c.redisClient.Del(ctx, "products:all", "product:"+productID).Err()

	return nil
}

func (c *CachedProductsStore) UpdateStockTx(ctx context.Context, productID string, stockChange int, tx pgx.Tx) error {
	return c.store.UpdateStockTx(ctx, productID, stockChange, tx)
}
