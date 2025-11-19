package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisDB holds the Redis connection
type RedisDB struct {
	Client *redis.Client
	Ctx    context.Context
}

// NewRedisDB creates a new Redis connection
func NewRedisDB(addr, password string, db int) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisDB{
		Client: client,
		Ctx:    context.Background(),
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisDB) Get(key string) (string, error) {
	return r.Client.Get(r.Ctx, key).Result()
}

// Set sets a value in Redis
func (r *RedisDB) Set(key string, value interface{}, expiration int) error {
	return r.Client.Set(r.Ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

// SetEx sets a value with expiration in Redis
func (r *RedisDB) SetEx(key string, value interface{}, expiration int) error {
	return r.Client.SetEx(r.Ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

// Delete deletes a key from Redis
func (r *RedisDB) Delete(key string) error {
	return r.Client.Del(r.Ctx, key).Err()
}

// Close closes the Redis connection
func (r *RedisDB) Close() error {
	return r.Client.Close()
}