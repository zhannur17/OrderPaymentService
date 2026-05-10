package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/domain"
)

const orderCacheTTL = 5 * time.Minute

type CachedOrderUseCase struct {
	uc    *OrderUseCase
	cache domain.Cache
}

func NewCachedOrderUseCase(uc *OrderUseCase, cache domain.Cache) *CachedOrderUseCase {
	return &CachedOrderUseCase{uc: uc, cache: cache}
}

func cacheKey(id string) string {
	return fmt.Sprintf("order:%s", id)
}

func (c *CachedOrderUseCase) CreateOrder(input CreateOrderInput) (*CreateOrderOutput, error) {
	out, err := c.uc.CreateOrder(input)
	if err != nil {
		return nil, err
	}
	c.cache.Delete(cacheKey(out.Order.ID))
	return out, nil
}

func (c *CachedOrderUseCase) GetOrder(id string) (*domain.Order, error) {
	key := cacheKey(id)

	cached, err := c.cache.Get(key)
	if err == nil {
		var order domain.Order
		if err := json.Unmarshal([]byte(cached), &order); err == nil {
			log.Printf("[Cache] HIT for order %s", id)
			return &order, nil
		}
	}

	log.Printf("[Cache] MISS for order %s", id)
	order, err := c.uc.GetOrder(id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(order)
	if err == nil {
		c.cache.Set(key, string(data), orderCacheTTL)
	}

	return order, nil
}

func (c *CachedOrderUseCase) CancelOrder(id string) (*domain.Order, error) {
	order, err := c.uc.CancelOrder(id)
	if err != nil {
		return nil, err
	}

	c.cache.Delete(cacheKey(id))
	log.Printf("[Cache] Invalidated cache for order %s", id)
	return order, nil
}

func (c *CachedOrderUseCase) GetOrdersByAmountRange(minAmount, maxAmount int64) ([]*domain.Order, error) {
	return c.uc.GetOrdersByAmountRange(minAmount, maxAmount)
}
