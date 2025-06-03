package gogi

import (
	"context"
	"fmt"

	"sync"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Writer *redis.Client
	Reader *redis.Client
}

var (
	redisClients   = make(map[string]*RedisClient)
	redisClientsMu sync.Mutex
)

func GetRedisClient(role string) (*RedisClient, error) {
	redisClientsMu.Lock()
	defer redisClientsMu.Unlock()

	if client, exists := redisClients[role]; exists {
		return client, nil
	}

	cfg := config.GetConfig().Redis[role]
	if cfg == nil {
		return nil, fmt.Errorf("no Redis config found for role: %s", role)
	}

	client, err := newRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	redisClients[role] = client
	return client, nil
}

func newRedisClient(cfg *config.RedisRoleConfig) (*RedisClient, error) {
	writer := redis.NewClient(&redis.Options{
		Addr:     cfg.Writer.Addr,
		Username: cfg.Writer.Username,
		Password: cfg.Writer.Password,
		DB:       cfg.Writer.DB,
	})

	if err := writer.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("writer redis ping failed: %w", err)
	}

	var reader *redis.Client
	if cfg.Reader == nil {
		reader = writer
	} else {
		reader = redis.NewClient(&redis.Options{
			Addr:     cfg.Reader.Addr,
			Username: cfg.Reader.Username,
			Password: cfg.Reader.Password,
			DB:       cfg.Reader.DB,
		})

		if err := reader.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("reader redis ping failed: %w", err)
		}
	}

	GetLogger().Debug(fmt.Sprintf("[Redis] Connected: writer=%s reader=%s", cfg.Writer.Addr, cfg.Reader.Addr))
	return &RedisClient{Writer: writer, Reader: reader}, nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Reader.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Writer.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Close() error {
	if r.Reader != r.Writer {
		r.Reader.Close()
	}
	return r.Writer.Close()
}
