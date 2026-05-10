package domain

import "time"

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Delete(key string) error
}
