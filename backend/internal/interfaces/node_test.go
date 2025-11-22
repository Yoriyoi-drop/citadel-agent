package interfaces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStruct that implements NodeInstance for testing
type TestNode struct {
	executionCount int
}

func (tn *TestNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	tn.executionCount++
	
	result := map[string]interface{}{
		"execution_count": tn.executionCount,
		"inputs":          inputs,
		"timestamp":       time.Now().Unix(),
	}
	
	return result, nil
}

func TestNodeInstanceInterface(t *testing.T) {
	t.Run("TestNode implements NodeInstance interface", func(t *testing.T) {
		var nodeInstance NodeInstance = &TestNode{}
		assert.NotNil(t, nodeInstance)
	})
	
	t.Run("TestNode Execute method works", func(t *testing.T) {
		node := &TestNode{}
		ctx := context.Background()
		inputs := map[string]interface{}{
			"test_key": "test_value",
		}
		
		result, err := node.Execute(ctx, inputs)
		assert.NoError(t, err)
		assert.Equal(t, 1, result["execution_count"])
		assert.Equal(t, inputs, result["inputs"])
		assert.NotNil(t, result["timestamp"])
		
		// Execute again to test increment
		result2, err := node.Execute(ctx, inputs)
		assert.NoError(t, err)
		assert.Equal(t, 2, result2["execution_count"])
	})
}

func TestExecutionResult(t *testing.T) {
	t.Run("ExecutionResult structure", func(t *testing.T) {
		result := ExecutionResult{
			Status:    "success",
			Data:      "test data",
			Error:     "test error",
			Timestamp: time.Now(),
		}
		
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "test data", result.Data)
		assert.Equal(t, "test error", result.Error)
		assert.NotZero(t, result.Timestamp)
	})
	
	t.Run("ExecutionResult with nil data", func(t *testing.T) {
		result := ExecutionResult{
			Status:    "error",
			Data:      nil,
			Error:     "something went wrong",
			Timestamp: time.Now(),
		}
		
		assert.Equal(t, "error", result.Status)
		assert.Nil(t, result.Data)
		assert.Equal(t, "something went wrong", result.Error)
		assert.NotZero(t, result.Timestamp)
	})
}