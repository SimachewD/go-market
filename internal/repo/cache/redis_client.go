package cache

import (
    "github.com/redis/go-redis/v9"
    "context"
    "time"
)

type RedisClient struct {
    Client *redis.Client
    Ctx    context.Context
}

func NewRedisClient(addr, password string, db int) *RedisClient {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    return &RedisClient{
        Client: rdb,
        Ctx:    context.Background(),
    }
}

func (r *RedisClient) Close() error {
    return r.Client.Close()
}

// Example helper
func (r *RedisClient) Set(key string, value any, ttl time.Duration) error {
    return r.Client.Set(r.Ctx, key, value, ttl).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
    return r.Client.Get(r.Ctx, key).Result()
}
