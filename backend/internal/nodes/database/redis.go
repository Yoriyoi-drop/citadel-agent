package database

import (
	"fmt"
	"time"

	"citadel-agent/backend/internal/nodes/base"
	"github.com/redis/go-redis/v9"
)

// RedisNode implements Redis operations
type RedisNode struct {
	*base.BaseNode
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Password  string `json:"password"`
	Database  int    `json:"database"`
	Operation string `json:"operation"` // get, set, delete, incr
	Key       string `json:"key"`
	Value     string `json:"value"`
	TTL       int    `json:"ttl"` // seconds
}

// NewRedisGetNode creates a Redis GET node
func NewRedisGetNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "redis_get",
		Name:        "Redis Get",
		Category:    "database",
		Description: "Get value from Redis",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "database",
		Color:       "#10b981",
		Inputs: []base.NodeInput{
			{
				ID:          "trigger",
				Name:        "Trigger",
				Type:        "any",
				Required:    false,
				Description: "Trigger the operation",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "value",
				Name:        "Value",
				Type:        "string",
				Description: "Retrieved value",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "host",
				Label:       "Host",
				Description: "Redis host",
				Type:        "string",
				Required:    true,
				Default:     "localhost",
			},
			{
				Name:        "port",
				Label:       "Port",
				Description: "Redis port",
				Type:        "number",
				Required:    true,
				Default:     6379,
			},
			{
				Name:        "password",
				Label:       "Password",
				Description: "Redis password",
				Type:        "password",
				Required:    false,
			},
			{
				Name:        "database",
				Label:       "Database",
				Description: "Redis database number",
				Type:        "number",
				Required:    false,
				Default:     0,
			},
			{
				Name:        "key",
				Label:       "Key",
				Description: "Redis key",
				Type:        "string",
				Required:    true,
			},
		},
		Tags: []string{"redis", "cache", "database"},
	}

	return &RedisNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// NewRedisSetNode creates a Redis SET node
func NewRedisSetNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "redis_set",
		Name:        "Redis Set",
		Category:    "database",
		Description: "Set value in Redis",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "database",
		Color:       "#10b981",
		Inputs: []base.NodeInput{
			{
				ID:          "trigger",
				Name:        "Trigger",
				Type:        "any",
				Required:    false,
				Description: "Trigger the operation",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "success",
				Name:        "Success",
				Type:        "boolean",
				Description: "Operation success",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "host",
				Label:       "Host",
				Description: "Redis host",
				Type:        "string",
				Required:    true,
				Default:     "localhost",
			},
			{
				Name:        "port",
				Label:       "Port",
				Description: "Redis port",
				Type:        "number",
				Required:    true,
				Default:     6379,
			},
			{
				Name:        "password",
				Label:       "Password",
				Description: "Redis password",
				Type:        "password",
				Required:    false,
			},
			{
				Name:        "database",
				Label:       "Database",
				Description: "Redis database number",
				Type:        "number",
				Required:    false,
				Default:     0,
			},
			{
				Name:        "key",
				Label:       "Key",
				Description: "Redis key",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "value",
				Label:       "Value",
				Description: "Value to set",
				Type:        "textarea",
				Required:    true,
			},
			{
				Name:        "ttl",
				Label:       "TTL (seconds)",
				Description: "Time to live in seconds (0 = no expiration)",
				Type:        "number",
				Required:    false,
				Default:     0,
			},
		},
		Tags: []string{"redis", "cache", "database"},
	}

	return &RedisNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute performs Redis operation
func (n *RedisNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config RedisConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.Database,
	})
	defer rdb.Close()

	// Ping to test connection
	if err := rdb.Ping(ctx.Context).Err(); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	var result map[string]interface{}

	// Determine operation based on node ID
	nodeID := n.GetMetadata().ID

	switch nodeID {
	case "redis_get":
		value, err := rdb.Get(ctx.Context, config.Key).Result()
		if err == redis.Nil {
			result = map[string]interface{}{
				"value":  nil,
				"exists": false,
			}
		} else if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		} else {
			result = map[string]interface{}{
				"value":  value,
				"exists": true,
			}
		}

	case "redis_set":
		ttl := time.Duration(config.TTL) * time.Second
		err := rdb.Set(ctx.Context, config.Key, config.Value, ttl).Err()
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}
		result = map[string]interface{}{
			"success": true,
			"key":     config.Key,
		}

	default:
		err := fmt.Errorf("unknown operation: %s", nodeID)
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	ctx.Logger.Info("Redis operation completed", map[string]interface{}{
		"operation": nodeID,
		"key":       config.Key,
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
