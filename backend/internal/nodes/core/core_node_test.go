package core

import (
	"context"
	"testing"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorNode_Execute(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		input    map[string]interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "valid input with required field",
			config: map[string]interface{}{
				"struct_tags": map[string]interface{}{
					"email": "required,email",
				},
			},
			input: map[string]interface{}{
				"email": "test@example.com",
			},
			expected: map[string]interface{}{
				"success": true,
				"result":  []string{}, // No validation errors expected
			},
			wantErr: false,
		},
		{
			name: "invalid email input",
			config: map[string]interface{}{
				"struct_tags": map[string]interface{}{
					"email": "required,email",
				},
			},
			input: map[string]interface{}{
				"email": "invalid-email",
			},
			expected: map[string]interface{}{
				"success": true,
				"result":  []string{"email: email"}, // Validation error expected
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			config: map[string]interface{}{
				"struct_tags": map[string]interface{}{
					"name": "required",
				},
			},
			input: map[string]interface{}{},
			expected: map[string]interface{}{
				"success": true,
				"result":  []string{"name: required field is missing"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewValidatorNode(tt.config)
			require.NoError(t, err)
			require.NotNil(t, node)

			result, err := node.Execute(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected["success"], result["success"])
			assert.Equal(t, tt.expected["result"], result["result"])
		})
	}
}

func TestUUIDGeneratorNode_Execute(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		input    map[string]interface{}
		wantErr  bool
	}{
		{
			name: "generate UUID v4",
			config: map[string]interface{}{
				"version": 4,
			},
			input:   map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "generate multiple UUIDs",
			config: map[string]interface{}{
				"version": 4,
				"count":   2.0,
			},
			input:   map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewUUIDGeneratorNode(tt.config)
			require.NoError(t, err)
			require.NotNil(t, node)

			result, err := node.Execute(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, result["success"].(bool))
			assert.NotNil(t, result["uuids"])
		})
	}
}

func TestConfigManagerNode_Execute(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		input    map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "execute with defaults source",
			config: map[string]interface{}{
				"source": "defaults",
				"defaults": map[string]interface{}{
					"key1": "default_value",
				},
			},
			input: map[string]interface{}{
				"key1": "input_value",
			},
			expected: "input_value", // input should override defaults
			wantErr:  false,
		},
		{
			name: "execute with defaults source and no override",
			config: map[string]interface{}{
				"source": "defaults",
				"defaults": map[string]interface{}{
					"key1": "default_value",
				},
			},
			input:    map[string]interface{}{},
			expected: "default_value",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewConfigManagerNode(tt.config)
			require.NoError(t, err)
			require.NotNil(t, node)

			result, err := node.Execute(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, result["success"].(bool))
		})
	}
}

func TestLoggerNode_Execute(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		input    map[string]interface{}
		wantErr  bool
	}{
		{
			name: "execute with message",
			config: map[string]interface{}{
				"level":   "info",
				"message": "test message",
			},
			input:   map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "execute with fields",
			config: map[string]interface{}{
				"level": "info",
				"fields": map[string]interface{}{
					"field1": "value1",
				},
			},
			input:   map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewLoggerNode(tt.config)
			require.NoError(t, err)
			require.NotNil(t, node)

			result, err := node.Execute(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, result["success"].(bool))
		})
	}
}

func TestNewConfigManagerNode(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := map[string]interface{}{
			"source": "defaults",
			"defaults": map[string]interface{}{
				"key1": "value1",
			},
		}

		node, err := NewConfigManagerNode(config)
		assert.NoError(t, err)
		assert.NotNil(t, node)
		assert.Implements(t, (*interfaces.NodeInstance)(nil), node)
	})

	t.Run("invalid config", func(t *testing.T) {
		config := map[string]interface{}{
			"source": 123, // Invalid type for source
		}

		node, err := NewConfigManagerNode(config)
		assert.NoError(t, err) // Should not error, but handle gracefully
		assert.NotNil(t, node)
	})
}