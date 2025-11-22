package basic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicNode_Execute(t *testing.T) {
	tests := []struct {
		name     string
		config   *BasicConfig
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "constant operation",
			config: &BasicConfig{
				Operation: BasicOpConstant,
				Value:     "test_value",
			},
			input: map[string]interface{}{},
			expected: map[string]interface{}{
				"success":   true,
				"result":    "test_value",
				"operation": "constant",
			},
		},
		{
			name: "passthrough operation",
			config: &BasicConfig{
				Operation: BasicOpPassthrough,
			},
			input: map[string]interface{}{
				"input1": "value1",
				"input2": 123,
			},
			expected: map[string]interface{}{
				"success":   true,
				"operation": "passthrough",
				"input1":    "value1",
				"input2":    123.0, // float64 due to JSON unmarshaling
			},
		},
		{
			name: "counter operation",
			config: &BasicConfig{
				Operation: BasicOpCounter,
			},
			input: map[string]interface{}{},
			expected: map[string]interface{}{
				"success":   true,
				"operation": "counter",
			},
		},
		{
			name: "condition operation - true",
			config: &BasicConfig{
				Operation: BasicOpCondition,
			},
			input: map[string]interface{}{
				"left":     5,
				"right":    3,
				"operator": ">",
			},
			expected: map[string]interface{}{
				"success":   true,
				"operation": "condition",
				"result":    true,
			},
		},
		{
			name: "condition operation - false",
			config: &BasicConfig{
				Operation: BasicOpCondition,
			},
			input: map[string]interface{}{
				"left":     2,
				"right":    3,
				"operator": ">",
			},
			expected: map[string]interface{}{
				"success":   true,
				"operation": "condition",
				"result":    false,
			},
		},
		{
			name: "math operation add",
			config: &BasicConfig{
				Operation:  BasicOpMath,
				MathValues: []interface{}{5, 3},
				MathOp:     "add",
			},
			input: map[string]interface{}{},
			expected: map[string]interface{}{
				"success":   true,
				"operation": "math",
				"result":    8.0,
			},
		},
		{
			name: "math operation multiply",
			config: &BasicConfig{
				Operation:  BasicOpMath,
				MathValues: []interface{}{4, 3},
				MathOp:     "multiply",
			},
			input: map[string]interface{}{},
			expected: map[string]interface{}{
				"success":   true,
				"operation": "math",
				"result":    12.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewBasicNode(tt.config)
			require.NotNil(t, node)

			result, err := node.Execute(context.Background(), tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected["success"], result["success"])
			assert.Equal(t, tt.expected["operation"], result["operation"])

			if tt.expected["result"] != nil {
				assert.Equal(t, tt.expected["result"], result["result"])
			}
			if tt.expected["input1"] != nil {
				assert.Equal(t, tt.expected["input1"], result["input1"])
			}
			if tt.expected["input2"] != nil {
				assert.Equal(t, tt.expected["input2"], result["input2"])
			}
		})
	}
}

func TestBasicNodeFromConfig(t *testing.T) {
	t.Run("valid constant config", func(t *testing.T) {
		config := map[string]interface{}{
			"operation": "constant",
			"value":     "test_value",
		}

		node, err := BasicNodeFromConfig(config)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		result, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, "test_value", result["result"])
		assert.Equal(t, "constant", result["operation"])
	})

	t.Run("valid counter config", func(t *testing.T) {
		config := map[string]interface{}{
			"operation": "counter",
		}

		// Test counter incrementing
		node, err := BasicNodeFromConfig(config)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		result1, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, true, result1["success"])
		assert.Equal(t, 1.0, result1["counter"]) // Counter starts at 1

		result2, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, true, result2["success"])
		assert.Equal(t, 2.0, result2["counter"])
	})

	t.Run("valid condition config", func(t *testing.T) {
		config := map[string]interface{}{
			"operation": "condition",
		}

		node, err := BasicNodeFromConfig(config)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		result, err := node.Execute(context.Background(), map[string]interface{}{
			"left":     10,
			"right":    5,
			"operator": ">",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, true, result["result"])
	})

	t.Run("valid math config", func(t *testing.T) {
		config := map[string]interface{}{
			"operation":      "math",
			"math_operation": "add",
			"math_values":    []interface{}{10, 5},
		}

		node, err := BasicNodeFromConfig(config)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		result, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, 15.0, result["result"])
	})

	t.Run("default operation is passthrough", func(t *testing.T) {
		config := map[string]interface{}{
			// No operation specified
		}

		node, err := BasicNodeFromConfig(config)
		assert.NoError(t, err)
		assert.NotNil(t, node)

		result, err := node.Execute(context.Background(), map[string]interface{}{
			"input_key": "input_value",
		})
		assert.NoError(t, err)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, "input_value", result["input_key"])
	})
}

func TestMathOperation_Errors(t *testing.T) {
	t.Run("division by zero", func(t *testing.T) {
		config := &BasicConfig{
			Operation:  BasicOpMath,
			MathValues: []interface{}{10, 0},
			MathOp:     "divide",
		}

		node := NewBasicNode(config)
		result, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err) // We expect no error, just a result with error info
		assert.Equal(t, false, result["success"])
	})

	t.Run("modulo by zero", func(t *testing.T) {
		config := &BasicConfig{
			Operation:  BasicOpMath,
			MathValues: []interface{}{10, 0},
			MathOp:     "modulo",
		}

		node := NewBasicNode(config)
		result, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err) // We expect no error, just a result with error info
		assert.Equal(t, false, result["success"])
	})

	t.Run("insufficient values for math", func(t *testing.T) {
		config := &BasicConfig{
			Operation:  BasicOpMath,
			MathValues: []interface{}{10}, // Only one value
			MathOp:     "add",
		}

		node := NewBasicNode(config)
		result, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err) // We expect no error, just a result with error info
		assert.Equal(t, false, result["success"])
	})

	t.Run("unsupported operation", func(t *testing.T) {
		config := &BasicConfig{
			Operation:  BasicOpMath,
			MathValues: []interface{}{10, 5},
			MathOp:     "unsupported_op",
		}

		node := NewBasicNode(config)
		result, err := node.Execute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err) // We expect no error, just a result with error info
		assert.Equal(t, false, result["success"])
	})
}