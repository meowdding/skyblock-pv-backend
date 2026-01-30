package internal

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteContext struct {
	redis   *redis.Client
	Config  *Config
	Pool    *pgxpool.Pool
	Context *context.Context
}

func NewRouteContext() RouteContext {
	config := NewConfig()
	var client *redis.Client = nil
	if config.RedisAddress != "" {
		if config.RedisUsername == nil {
			opts, err := redis.ParseURL(config.RedisAddress)
			if err != nil {
				panic(err)
			}

			client = redis.NewClient(opts)
		} else {
			client = redis.NewClient(&redis.Options{
				Addr:     config.RedisAddress,
				Username: *config.RedisUsername,
				Password: *config.RedisPassword,
			})
		}
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, config.PostgresUri)
	if err != nil {
		panic(err)
	}

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	routeContext := RouteContext{client, &config, pool, &ctx}
	if err := setupDatabase(&routeContext); err != nil {
		panic(err)
	}

	return routeContext
}

func (ctx *RouteContext) IsHighProfileAccount(playerId string) bool {
	if ctx.Config == nil {
		return false
	}
	return slices.Contains(ctx.Config.HighProfileAccounts, playerId)
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

//go:embed migrations/*.sql
var migrationFS embed.FS

func setupDatabase(ctx *RouteContext) error {
	connection, err := sql.Open("pgx", ctx.Config.PostgresUri)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer connection.Close()

	driver, err := migratepgx.WithInstance(connection, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate database driver: %w", err)
	}

	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migrate source driver: %w", err)
	}

	instance, err := migrate.NewWithInstance("migration-fs", source, "migrate-pgx-db", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err = instance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	srcErr, dbErr := instance.Close()
	if srcErr != nil {
		return fmt.Errorf("failed to close migrate source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migrate database: %w", dbErr)
	}
	return nil
}
