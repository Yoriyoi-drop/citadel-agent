package database

import (
	"time"

	"citadel-agent/backend/internal/nodes/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBNode implements MongoDB operations
type MongoDBNode struct {
	*base.BaseNode
	client *mongo.Client
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	ConnectionString string                 `json:"connection_string"`
	Database         string                 `json:"database"`
	Collection       string                 `json:"collection"`
	Operation        string                 `json:"operation"` // find, insert, update, delete
	Filter           map[string]interface{} `json:"filter"`
	Document         map[string]interface{} `json:"document"`
	Limit            int64                  `json:"limit"`
}

// NewMongoDBNode creates a new MongoDB node
func NewMongoDBNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "mongodb_find",
		Name:        "MongoDB Find",
		Category:    "database",
		Description: "Find documents in MongoDB",
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
				Description: "Trigger the query",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "documents",
				Name:        "Documents",
				Type:        "array",
				Description: "Found documents",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "connection_string",
				Label:       "Connection String",
				Description: "MongoDB connection string",
				Type:        "password",
				Required:    true,
			},
			{
				Name:        "database",
				Label:       "Database",
				Description: "Database name",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "collection",
				Label:       "Collection",
				Description: "Collection name",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "filter",
				Label:       "Filter",
				Description: "Query filter (JSON)",
				Type:        "textarea",
				Required:    false,
			},
			{
				Name:        "limit",
				Label:       "Limit",
				Description: "Maximum number of documents",
				Type:        "number",
				Required:    false,
				Default:     100,
			},
		},
		Tags: []string{"mongodb", "database", "nosql"},
	}

	return &MongoDBNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute performs MongoDB find operation
func (n *MongoDBNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config MongoDBConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx.Context, options.Client().ApplyURI(config.ConnectionString))
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}
	defer client.Disconnect(ctx.Context)

	// Get collection
	collection := client.Database(config.Database).Collection(config.Collection)

	// Build filter
	filter := bson.M{}
	if config.Filter != nil {
		filter = config.Filter
	}

	// Set options
	findOptions := options.Find()
	if config.Limit > 0 {
		findOptions.SetLimit(config.Limit)
	}

	// Execute find
	cursor, err := collection.Find(ctx.Context, filter, findOptions)
	if err != nil {
		ctx.Logger.Error("MongoDB find failed", err, map[string]interface{}{
			"database":   config.Database,
			"collection": config.Collection,
		})
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}
	defer cursor.Close(ctx.Context)

	// Decode results
	var documents []map[string]interface{}
	if err := cursor.All(ctx.Context, &documents); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	result := map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
	}

	ctx.Logger.Info("MongoDB find completed", map[string]interface{}{
		"count": len(documents),
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
