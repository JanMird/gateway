package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

type Cache struct {
	*redis.Client
}

func NewRedisClient(cfg Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	ctx := context.Background()
	if err := db.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cannot create client: %w", err)
	}

	return db, nil
}

func New(client *redis.Client) *Cache {
	return &Cache{Client: client}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	result, err := c.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	return result, nil
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if c.Client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	err := c.Client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if c.Client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	err := c.Client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}
