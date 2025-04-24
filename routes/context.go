package routes

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const EnableRedis = false

type RouteContext struct {
	redis *redis.Client
}

func NewRouteContext(redisClient *redis.Client) RouteContext {
	return RouteContext{
		redis: redisClient,
	}
}

func (ctx *RouteContext) GetFromCache(path string, key string) (string, error) {
	if !EnableRedis {
		return "", fmt.Errorf("not found")
	}
	result := ctx.redis.Get(context.Background(), fmt.Sprintf("%s:%s", path, key))
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.String(), nil
}

func (ctx *RouteContext) AddToCache(path string, key string, value *string, duration time.Duration) error {
	if !EnableRedis {
		return nil
	}
	result := ctx.redis.Set(context.Background(), fmt.Sprintf("%s:%s", path, key), value, duration)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}
