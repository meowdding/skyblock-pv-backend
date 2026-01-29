package internal

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
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
	pool, err := pgxpool.New(ctx, config.dbUri())
	if err != nil {
		panic(err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}

	routeContext := RouteContext{client, &config, pool, &ctx}

	setupDatabase(&routeContext)

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

func (conf Config) dbUri() string {
	if conf.DevMode {
		return *conf.PostgresUri
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, db)
}

//go:embed migrations/*.sql
var migrationFS embed.FS

func setupDatabase(ctx *RouteContext) {
	// Create a dedicated connection for migrations because migrate wont take a pgx conn (needs database/sql conn)
	migrateConn, err := sql.Open("pgx", ctx.Config.dbUri())
	if err != nil {
		panic(fmt.Sprintf("failed to acquire connection for migrations: %w", err))
	}
	//goland:noinspection GoUnhandledErrorResult
	defer migrateConn.Close()
	migrateDriver, err := migratepgx.WithInstance(migrateConn, &migratepgx.Config{})

	if err != nil {
		panic(fmt.Sprintf("failed to create migrate driver: %s", err))
	}
	migrateSource, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		panic(fmt.Sprintf("failed to create migrate source: %s", err))
	}
	m, err := migrate.NewWithInstance("migration-fs", migrateSource, "migration-db", migrateDriver)
	if err != nil {
		panic(fmt.Sprintf("failed to create migrate instance: %s", err))
	}

	// Apply all migrations up to the latest
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Sprintf("failed to apply migrations: %s", err))
	}
}
