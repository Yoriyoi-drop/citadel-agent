package drivers

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrMongoNotConnected = errors.New("mongodb not connected")
	ErrInvalidOperation  = errors.New("invalid operation")
)

// MongoDBDriver handles MongoDB operations
type MongoDBDriver struct {
	client   *mongo.Client
	database string
	timeout  time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// NewMongoDBDriver creates a new MongoDB driver
func NewMongoDBDriver(config MongoDBConfig) (*MongoDBDriver, error) {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoDBDriver{
		client:   client,
		database: config.Database,
		timeout:  config.Timeout,
	}, nil
}

// Insert inserts a single document
func (d *MongoDBDriver) Insert(ctx context.Context, collection string, document interface{}) (string, error) {
	if d.client == nil {
		return "", ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return "", err
	}

	if oid, ok := result.InsertedID.(string); ok {
		return oid, nil
	}

	return "", nil
}

// InsertMany inserts multiple documents
func (d *MongoDBDriver) InsertMany(ctx context.Context, collection string, documents []interface{}) ([]string, error) {
	if d.client == nil {
		return nil, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.InsertedIDs))
	for _, id := range result.InsertedIDs {
		if oid, ok := id.(string); ok {
			ids = append(ids, oid)
		}
	}

	return ids, nil
}

// Find finds documents matching filter
func (d *MongoDBDriver) Find(ctx context.Context, collection string, filter interface{}, opts *options.FindOptions) ([]bson.M, error) {
	if d.client == nil {
		return nil, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// FindOne finds a single document
func (d *MongoDBDriver) FindOne(ctx context.Context, collection string, filter interface{}) (bson.M, error) {
	if d.client == nil {
		return nil, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	var result bson.M
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateOne updates a single document
func (d *MongoDBDriver) UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) (int64, error) {
	if d.client == nil {
		return 0, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// UpdateMany updates multiple documents
func (d *MongoDBDriver) UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (int64, error) {
	if d.client == nil {
		return 0, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// DeleteOne deletes a single document
func (d *MongoDBDriver) DeleteOne(ctx context.Context, collection string, filter interface{}) (int64, error) {
	if d.client == nil {
		return 0, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// DeleteMany deletes multiple documents
func (d *MongoDBDriver) DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error) {
	if d.client == nil {
		return 0, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// Aggregate performs aggregation pipeline
func (d *MongoDBDriver) Aggregate(ctx context.Context, collection string, pipeline interface{}) ([]bson.M, error) {
	if d.client == nil {
		return nil, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Count counts documents matching filter
func (d *MongoDBDriver) Count(ctx context.Context, collection string, filter interface{}) (int64, error) {
	if d.client == nil {
		return 0, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateIndex creates an index
func (d *MongoDBDriver) CreateIndex(ctx context.Context, collection string, keys interface{}, unique bool) (string, error) {
	if d.client == nil {
		return "", ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique),
	}

	name, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return "", err
	}

	return name, nil
}

// DropIndex drops an index
func (d *MongoDBDriver) DropIndex(ctx context.Context, collection string, name string) error {
	if d.client == nil {
		return ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	_, err := coll.Indexes().DropOne(ctx, name)
	return err
}

// Close closes the MongoDB connection
func (d *MongoDBDriver) Close(ctx context.Context) error {
	if d.client == nil {
		return nil
	}
	return d.client.Disconnect(ctx)
}

// BulkWrite performs bulk write operations
func (d *MongoDBDriver) BulkWrite(ctx context.Context, collection string, operations []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	if d.client == nil {
		return nil, ErrMongoNotConnected
	}

	coll := d.client.Database(d.database).Collection(collection)
	result, err := coll.BulkWrite(ctx, operations)
	if err != nil {
		return nil, err
	}

	return result, nil
}
