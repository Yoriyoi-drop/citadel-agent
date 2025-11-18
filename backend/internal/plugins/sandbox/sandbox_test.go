// backend/internal/plugins/sandbox/sandbox_test.go
package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSSandbox_Execute(t *testing.T) {
	config := &JSSandboxConfig{
		Timeout:         1 * time.Second,
		MaxMemoryMB:     50,
		MaxOutputLength: 1000,
		NetworkAccess:   false,
		FileAccess:      false,
		BlockedFunctions: []string{"eval", "Function", "require", "import"},
	}

	sandbox, err := NewJSSandbox(config)
	require.NoError(t, err)
	defer sandbox.Close()

	t.Run("should execute safe JavaScript code", func(t *testing.T) {
		code := `
			var result = input.x + input.y;
			result;
		`

		result, err := sandbox.Execute(context.Background(), code, map[string]interface{}{
			"x": 5,
			"y": 3,
		})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, float64(8), result.Data)
	})

	t.Run("should block dangerous JavaScript code", func(t *testing.T) {
		code := `
			eval("console.log('malicious')");
		`

		result, err := sandbox.Execute(context.Background(), code, map[string]interface{}{})

		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Code validation failed")
	})

	t.Run("should timeout on infinite loop", func(t *testing.T) {
		code := `
			while(true) {
				// infinite loop
			}
		`

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		result, err := sandbox.Execute(ctx, code, map[string]interface{}{})

		require.NoError(t, err)
		assert.False(t, result.Success)
	})

	t.Run("should handle input correctly", func(t *testing.T) {
		code := `
			var output = {
				name: input.name,
				doubleAge: input.age * 2,
				isAdult: input.age >= 18
			};
			output;
		`

		input := map[string]interface{}{
			"name": "John",
			"age":  25,
		}

		result, err := sandbox.Execute(context.Background(), code, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		
		resultMap, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, float64(50), resultMap["doubleAge"])
		assert.Equal(t, true, resultMap["isAdult"])
	})
}

func TestJSSandbox_ValidateCode(t *testing.T) {
	config := &JSSandboxConfig{
		BlockedFunctions: []string{"eval", "Function", "require", "import", "process", "global"},
	}

	sandbox, err := NewJSSandbox(config)
	require.NoError(t, err)
	defer sandbox.Close()

	testCases := []struct {
		name     string
		code     string
		shouldPass bool
	}{
		{
			name:     "safe code should pass",
			code:     "var x = 5; var y = x + 10;",
			shouldPass: true,
		},
		{
			name:     "eval should be blocked",
			code:     "eval('console.log(\"test\")');",
			shouldPass: false,
		},
		{
			name:     "Function constructor should be blocked",
			code:     "var f = new Function('return 42');",
			shouldPass: false,
		},
		{
			name:     "process access should be blocked",
			code:     "process.exit();",
			shouldPass: false,
		},
		{
			name:     "global access should be blocked",
			code:     "global.test = 'value';",
			shouldPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := sandbox.validateCode(tc.code)
			if tc.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPythonSandbox_Execute(t *testing.T) {
	config := &PythonSandboxConfig{
		Timeout:          2 * time.Second,
		MaxMemoryMB:      100,
		MaxOutputLength:  1000,
		NetworkAccess:    false,
		FileAccess:       false,
		AllowedLibraries: []string{"json", "math", "datetime"},
	}

	sandbox, err := NewPythonSandbox(config)
	require.NoError(t, err)
	defer sandbox.Close()

	t.Run("should execute safe Python code", func(t *testing.T) {
		code := `
import json
result = input_value * 2
print(json.dumps({"result": result, "input": input_value}))
`

		result, err := sandbox.Execute(context.Background(), code, map[string]interface{}{
			"input_value": 21,
		})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Data, "result")
	})

	t.Run("should block dangerous Python code", func(t *testing.T) {
		code := `
import os
os.system("echo dangerous")
`

		result, err := sandbox.Execute(context.Background(), code, map[string]interface{}{})

		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "Code validation failed")
	})
}

func TestPythonSandbox_ValidateCode(t *testing.T) {
	config := &PythonSandboxConfig{
		AllowedLibraries: []string{"json", "math", "datetime"},
	}

	sandbox, err := NewPythonSandbox(config)
	require.NoError(t, err)
	defer sandbox.Close()

	testCases := []struct {
		name     string
		code     string
		shouldPass bool
	}{
		{
			name:     "safe code should pass",
			code:     "result = input_value * 2",
			shouldPass: true,
		},
		{
			name:     "import os should be blocked",
			code:     "import os",
			shouldPass: false,
		},
		{
			name:     "import subprocess should be blocked",
			code:     "import subprocess",
			shouldPass: false,
		},
		{
			name:     "eval should be blocked",
			code:     "eval('print(\"malicious\")')",
			shouldPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := sandbox.validateCode(tc.code)
			if tc.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSecurePluginRuntime_Execute(t *testing.T) {
	runtime, err := NewSecurePluginRuntime()
	require.NoError(t, err)
	defer runtime.Close()

	t.Run("should execute JavaScript plugin safely", func(t *testing.T) {
		jsCode := `
			var doubled = input.number * 2;
			var result = {
				original: input.number,
				doubled: doubled,
				isEven: doubled % 2 === 0
			};
			result;
		`

		result, err := runtime.ExecutePlugin(context.Background(), JSType, jsCode, nil, "", map[string]interface{}{
			"number": 7,
		})

		require.NoError(t, err)
		assert.True(t, result.Success)

		resultMap, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(7), resultMap["original"])
		assert.Equal(t, float64(14), resultMap["doubled"])
		assert.Equal(t, true, resultMap["isEven"])
	})

	t.Run("should validate input data", func(t *testing.T) {
		// Create a runtime with specific limits
		jsConfig := &JSSandboxConfig{
			Timeout:         1 * time.Second,
			MaxMemoryMB:     50,
			MaxOutputLength: 1000,
			NetworkAccess:   false,
			FileAccess:      false,
			BlockedFunctions: []string{"eval", "Function"},
		}

		pyConfig := &PythonSandboxConfig{
			Timeout:          1 * time.Second,
			MaxMemoryMB:      50,
			MaxOutputLength:  1000,
			NetworkAccess:    false,
			FileAccess:       false,
			AllowedLibraries: []string{"json", "math"},
		}

		wasmConfig := &WASMSandboxConfig{
			Timeout:        1 * time.Second,
			MaxMemoryPages: 100,
			MaxExecutions:  100,
		}

		manager, err := NewPluginManager(jsConfig, pyConfig, wasmConfig)
		require.NoError(t, err)
		defer manager.Close()

		// Test with valid input
		code := "result = input.value * input.multiplier;"
		result, err := manager.ExecutePlugin(context.Background(), JSType, code, nil, "", map[string]interface{}{
			"value": 5,
			"multiplier": 3,
		})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, float64(15), result.Data)
	})
}