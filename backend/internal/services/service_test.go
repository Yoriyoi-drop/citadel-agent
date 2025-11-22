package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/citadel-agent/backend/internal/models"
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

func TestWorkflowService(t *testing.T) {
	db := setupTestDB()
	workflowService := NewWorkflowService(db)

	// Test workflow creation
	workflow := &models.Workflow{
		ID:   "test-workflow-id",
		Name: "Test Workflow",
	}
	
	err := workflowService.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Test retrieving workflow
	retrievedWorkflow, err := workflowService.GetWorkflow("test-workflow-id")
	assert.NoError(t, err)
	assert.Equal(t, "Test Workflow", retrievedWorkflow.Name)

	// Test updating workflow
	workflow.Name = "Updated Test Workflow"
	err = workflowService.UpdateWorkflow(workflow)
	assert.NoError(t, err)

	// Verify update
	updatedWorkflow, err := workflowService.GetWorkflow("test-workflow-id")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Workflow", updatedWorkflow.Name)
}

func TestNodeService(t *testing.T) {
	db := setupTestDB()
	nodeService := NewNodeService(db)

	// Test node creation
	node := &models.Node{
		ID:         "test-node-id",
		WorkflowID: "test-workflow-id",
		Type:       "http_request",
		Name:       "Test Node",
	}
	
	err := nodeService.CreateNode(node)
	assert.NoError(t, err)

	// Test retrieving node
	retrievedNode, err := nodeService.GetNode("test-node-id")
	assert.NoError(t, err)
	assert.Equal(t, "Test Node", retrievedNode.Name)

	// Test updating node
	node.Name = "Updated Test Node"
	err = nodeService.UpdateNode(node)
	assert.NoError(t, err)

	// Verify update
	updatedNode, err := nodeService.GetNode("test-node-id")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Node", updatedNode.Name)
}

func TestExecutionService(t *testing.T) {
	db := setupTestDB()
	executionService := NewExecutionService(db)

	// Test execution creation
	execution := &models.Execution{
		ID:         "test-execution-id",
		WorkflowID: "test-workflow-id",
		Status:     "running",
	}
	
	err := executionService.CreateExecution(execution)
	assert.NoError(t, err)

	// Test retrieving execution
	retrievedExecution, err := executionService.GetExecution("test-execution-id")
	assert.NoError(t, err)
	assert.Equal(t, "running", retrievedExecution.Status)

	// Test completing execution
	err = executionService.CompleteExecution("test-execution-id", map[string]interface{}{"result": "success"})
	assert.NoError(t, err)

	// Verify completion
	completedExecution, err := executionService.GetExecution("test-execution-id")
	assert.NoError(t, err)
	assert.Equal(t, "completed", completedExecution.Status)
}

func TestUserService(t *testing.T) {
	db := setupTestDB()
	userService := NewUserService(db)

	// Test user creation
	user := &models.User{
		ID:       "test-user-id",
		Email:    os.Getenv("TEST_USER_EMAIL"),
		Username: os.Getenv("TEST_USER_USERNAME"),
		Password: os.Getenv("TEST_USER_PASSWORD"),
	}

	// Set defaults for test if environment variables are not set
	if user.Email == "" {
		user.Email = "test@example.com"
	}
	if user.Username == "" {
		user.Username = "testuser"
	}
	if user.Password == "" {
		user.Password = "password123"
	}
	
	err := userService.CreateUser(user)
	assert.NoError(t, err)

	// Test retrieving user
	retrievedUser, err := userService.GetUser("test-user-id")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", retrievedUser.Email)

	// Test updating user
	user.FirstName = "Test"
	user.LastName = "User"
	err = userService.UpdateUser(user)
	assert.NoError(t, err)

	// Verify update
	updatedUser, err := userService.GetUser("test-user-id")
	assert.NoError(t, err)
	assert.Equal(t, "Test", updatedUser.FirstName)
	assert.Equal(t, "User", updatedUser.LastName)
}

func TestServiceValidation(t *testing.T) {
	// Test validation functions
	err := ValidateRequired("", "name")
	assert.Error(t, err)

	err = ValidateRequired("test", "name")
	assert.NoError(t, err)

	errs := ValidateWorkflowInput("", "description")
	assert.True(t, errs.HasErrors())

	errs = ValidateWorkflowInput("Test Workflow", "")
	assert.False(t, errs.HasErrors())

	emailErr := ValidateEmail("invalid-email")
	assert.Error(t, emailErr)

	emailErr = ValidateEmail("valid@example.com")
	assert.NoError(t, emailErr)
}

func TestMain(m *testing.M) {
	// Set up test environment
	exitCode := m.Run()

	// Clean up
	os.Exit(exitCode)
}