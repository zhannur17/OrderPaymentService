package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(key string) (string, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisCache) Set(key string, value string, ttl time.Duration) error {
	return r.client.Set(context.Background(), key, value, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
