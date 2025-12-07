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

func (ctx *RouteContext) GetAll(path string) ([]string, error) {
	if ctx.redis == nil {
		return nil, fmt.Errorf("not found")
	}
	result := ctx.redis.Keys(context.Background(), fmt.Sprintf("%s:*", path))
	if result.Err() != nil {
		return nil, result.Err()
	}

	keys := result.Val()
	data := make([]string, len(keys))

	var err error = nil
	for i, key := range keys {
		data[i], err = ctx.GetFromCacheByKey(key)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func createKey(path string, key string) string {
	return fmt.Sprintf("%s:%s", path, key)
}

func (ctx *RouteContext) IsCached(path string, key string) bool {
	if ctx.redis == nil {
		return false
	}
	result := ctx.redis.Get(context.Background(), createKey(path, key))
	return result.Err() == nil
}

func (ctx *RouteContext) HasErrorCached(path string, key string) bool {
	if ctx.redis == nil {
		return false
	}
	result := ctx.redis.Get(context.Background(), createKey(path, createKey(key, "error")))
	return result.Err() == nil
}

func (ctx *RouteContext) GetTtlMilli(path string, key string) (time.Duration, error) {
	if ctx.redis == nil {
		return -1, nil
	}
	result := ctx.redis.PTTL(context.Background(), createKey(path, key))
	if result.Err() != nil {
		return -1, result.Err()
	}
	return result.Val() / time.Millisecond, nil
}

func (ctx *RouteContext) Delete(path string, key string) error {
	if ctx.redis == nil {
		return fmt.Errorf("not found")
	}
	result := ctx.redis.Del(context.Background(), createKey(path, key))
	return result.Err()
}

func (ctx *RouteContext) GetFromCacheByKey(key string) (string, error) {
	if ctx.redis == nil {
		return "", fmt.Errorf("not found")
	}
	result := ctx.redis.Get(context.Background(), key)
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

func (ctx *RouteContext) GetFromCache(authContext *AuthenticationContext, path string, key string) (string, error) {
	if authContext != nil && (*authContext).BypassCache {
		return "", fmt.Errorf("not found")
	}
	return ctx.GetFromCacheByKey(createKey(path, key))
}

func (ctx *RouteContext) AddToCache(path string, key string, value interface{}, duration time.Duration) error {
	if ctx.redis == nil {
		return nil
	}
	result := ctx.redis.Set(context.Background(), createKey(path, key), value, duration)
	return result.Err()
}

func (ctx *RouteContext) AddToErrorCache(path string, key string, duration time.Duration) error {
	if ctx.redis == nil {
		return nil
	}
	result := ctx.redis.Set(context.Background(), createKey(path, createKey(key, "error")), "", duration)
	return result.Err()
}
