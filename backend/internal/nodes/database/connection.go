package database

import (
	"context"
	"errors"
	"sync"
)

// DriverType represents the type of database driver
type DriverType string

const (
	DriverPostgres DriverType = "postgres"
	DriverRedis    DriverType = "redis"
	DriverMySQL    DriverType = "mysql"
	DriverMongoDB  DriverType = "mongodb"
)

// ConnectionConfig holds configuration for a database connection
type ConnectionConfig struct {
	Type     DriverType
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Options  map[string]interface{}
}

// Driver interface that all database drivers must implement
type Driver interface {
	Connect(ctx context.Context, config ConnectionConfig) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	Execute(ctx context.Context, query string, args ...interface{}) (interface{}, error)
}

// ConnectionManager manages database connections
type ConnectionManager struct {
	drivers map[DriverType]Driver
	mu      sync.RWMutex
}

var (
	instance *ConnectionManager
	once     sync.Once
)

// GetConnectionManager returns the singleton instance of ConnectionManager
func GetConnectionManager() *ConnectionManager {
	once.Do(func() {
		instance = &ConnectionManager{
			drivers: make(map[DriverType]Driver),
		}
	})
	return instance
}

// RegisterDriver registers a new database driver
func (m *ConnectionManager) RegisterDriver(driverType DriverType, driver Driver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[driverType] = driver
}

// GetDriver returns a registered driver
func (m *ConnectionManager) GetDriver(driverType DriverType) (Driver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	driver, ok := m.drivers[driverType]
	if !ok {
		return nil, errors.New("driver not found")
	}
	return driver, nil
}
