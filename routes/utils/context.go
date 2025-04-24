package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RouteContext struct {
	redis  *redis.Client
	Config *Config
}

func NewRouteContext() RouteContext {
	config := NewConfig()
	var client *redis.Client = nil
	if config.RedisAddress != "" {
		client = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddress,
			Username: config.RedisUsername,
			Password: config.RedisPassword,
		})
	}

	return RouteContext{client, &config}
}

func (ctx *RouteContext) GetFromCache(path string, key string) (string, error) {
	if ctx.redis == nil {
		return "", fmt.Errorf("not found")
	}
	result := ctx.redis.Get(context.Background(), fmt.Sprintf("%s:%s", path, key))
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

func (ctx *RouteContext) AddToCache(path string, key string, value *string, duration time.Duration) error {
	if ctx.redis == nil {
		return nil
	}
	result := ctx.redis.Set(context.Background(), fmt.Sprintf("%s:%s", path, key), value, duration)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}
