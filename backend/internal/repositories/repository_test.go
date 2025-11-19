package repositories

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"citadel-agent/backend/internal/models"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Workflow{}, &models.Node{}, &models.Execution{})

	return db
}

func TestNodeRepository(t *testing.T) {
	db := setupTestDB()
	nodeRepo := NewNodeRepository(db)

	// Test Create
	node := &models.Node{
		ID:   "test-node-id",
		Type: "http_request",
		Name: "Test Node",
	}
	
	err := nodeRepo.Create(node)
	assert.NoError(t, err)

	// Test GetByID
	retrievedNode, err := nodeRepo.GetByID("test-node-id")
	assert.NoError(t, err)
	assert.Equal(t, "Test Node", retrievedNode.Name)

	// Test GetByWorkflowID
	nodes, err := nodeRepo.GetByWorkflowID("test-workflow-id")
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
}

func TestWorkflowRepository(t *testing.T) {
	db := setupTestDB()
	workflowRepo := NewWorkflowRepository(db)

	// Test Create
	workflow := &models.Workflow{
		ID:   "test-workflow-id",
		Name: "Test Workflow",
	}
	
	err := workflowRepo.Create(workflow)
	assert.NoError(t, err)

	// Test GetByID
	retrievedWorkflow, err := workflowRepo.GetByID("test-workflow-id")
	assert.NoError(t, err)
	assert.Equal(t, "Test Workflow", retrievedWorkflow.Name)

	// Test GetAll
	workflows, err := workflowRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, workflows, 1)
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)

	// Test Create
	user := &models.User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Username: "testuser",
	}
	
	err := userRepo.Create(user)
	assert.NoError(t, err)

	// Test GetByID
	retrievedUser, err := userRepo.GetByID("test-user-id")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", retrievedUser.Email)

	// Test GetByEmail
	retrievedUserByEmail, err := userRepo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", retrievedUserByEmail.Username)
}

func TestExecutionRepository(t *testing.T) {
	db := setupTestDB()
	executionRepo := NewExecutionRepository(db)

	// Test Create
	execution := &models.Execution{
		ID:         "test-execution-id",
		WorkflowID: "test-workflow-id",
		Status:     "running",
	}
	
	err := executionRepo.Create(execution)
	assert.NoError(t, err)

	// Test GetByID
	retrievedExecution, err := executionRepo.GetByID("test-execution-id")
	assert.NoError(t, err)
	assert.Equal(t, "running", retrievedExecution.Status)

	// Test GetByWorkflowID
	executions, err := executionRepo.GetByWorkflowID("test-workflow-id")
	assert.NoError(t, err)
	assert.Len(t, executions, 1)
}

func TestRepositoryFactory(t *testing.T) {
	db := setupTestDB()
	factory := NewRepositoryFactory(db)

	// Test that we can get all repository types
	nodeRepo := factory.GetNodeRepository()
	assert.NotNil(t, nodeRepo)

	workflowRepo := factory.GetWorkflowRepository()
	assert.NotNil(t, workflowRepo)

	userRepo := factory.GetUserRepository()
	assert.NotNil(t, userRepo)

	executionRepo := factory.GetExecutionRepository()
	assert.NotNil(t, executionRepo)
}

func TestRepositoryBase(t *testing.T) {
	db := setupTestDB()
	baseRepo := NewBaseRepository(db)

	// Test that we can get the DB
	assert.Equal(t, db, baseRepo.GetDB())

	// Test that transaction method exists (just test that it doesn't panic)
	err := baseRepo.WithTransaction(func(tx *gorm.DB) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestMain(m *testing.M) {
	// Set up test environment
	exitCode := m.Run()

	// Clean up
	os.Exit(exitCode)
}