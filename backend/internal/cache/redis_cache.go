package cache

import (
    "context"
    "errors"
    "time"

    "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type Client interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
    DeleteByPattern(ctx context.Context, pattern string) error
}

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
    return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
    value, err := c.client.Get(ctx, key).Result()
    if errors.Is(err, redis.Nil) {
        return "", ErrCacheMiss
    }

    return value, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
    iterator := c.client.Scan(ctx, 0, pattern, 0).Iterator()
    keys := make([]string, 0)

    for iterator.Next(ctx) {
        keys = append(keys, iterator.Val())
    }

    if err := iterator.Err(); err != nil {
        return err
    }

    if len(keys) == 0 {
        return nil
    }

    return c.client.Del(ctx, keys...).Err()
}
