// citadel-agent/backend/internal/repositories/repository_factory.go
package repositories

import (
	"gorm.io/gorm"
)

// RepositoryFactory creates and manages repositories
type RepositoryFactory struct {
	db *gorm.DB
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// GetWorkflowRepository returns a new workflow repository instance
func (f *RepositoryFactory) GetWorkflowRepository() *WorkflowRepository {
	return NewWorkflowRepository(f.db)
}

// GetNodeRepository returns a new node repository instance
func (f *RepositoryFactory) GetNodeRepository() *NodeRepository {
	return NewNodeRepository(f.db)
}

// GetExecutionRepository returns a new execution repository instance
func (f *RepositoryFactory) GetExecutionRepository() *ExecutionRepository {
	return NewExecutionRepository(f.db)
}

// GetUserRepository returns a new user repository instance
func (f *RepositoryFactory) GetUserRepository() *UserRepository {
	return NewUserRepository(f.db)
}

// GetDB returns the underlying database connection
func (f *RepositoryFactory) GetDB() *gorm.DB {
	return f.db
}