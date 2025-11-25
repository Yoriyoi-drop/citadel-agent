package drivers

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"citadel-agent/backend/internal/nodes/database"
)

// RedisDriver implements the Driver interface for Redis
type RedisDriver struct {
	client *redis.Client
}

// NewRedisDriver creates a new Redis driver
func NewRedisDriver() *RedisDriver {
	return &RedisDriver{}
}

// Connect establishes a connection to Redis
func (d *RedisDriver) Connect(ctx context.Context, config database.ConnectionConfig) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	
	d.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       0, // Default DB
	})

	return d.Ping(ctx)
}

// Disconnect closes the connection
func (d *RedisDriver) Disconnect(ctx context.Context) error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// Ping checks the connection
func (d *RedisDriver) Ping(ctx context.Context) error {
	if d.client == nil {
		return fmt.Errorf("connection not established")
	}
	return d.client.Ping(ctx).Err()
}

// Execute executes a command
func (d *RedisDriver) Execute(ctx context.Context, command string, args ...interface{}) (interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("connection not established")
	}

	// Simple execution wrapper
	// In reality, we would parse the command and call the appropriate method
	// or use Do()
	
	cmdArgs := append([]interface{}{command}, args...)
	val, err := d.client.Do(ctx, cmdArgs...).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}
