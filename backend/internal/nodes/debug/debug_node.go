// backend/internal/nodes/debug/debug_node.go
package debug

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// DebugOperationType represents the type of debug operation
type DebugOperationType string

const (
	DebugOpLog     DebugOperationType = "log"
	DebugOpBreak   DebugOperationType = "break"
	DebugOpInspect DebugOperationType = "inspect"
	DebugOpMeasure DebugOperationType = "measure"
	DebugOpTrace   DebugOperationType = "trace"
)

// DebugLevel represents the debug level/verbosity
type DebugLevel string

const (
	DebugLevelLow    DebugLevel = "low"
	DebugLevelMedium DebugLevel = "medium"
	DebugLevelHigh   DebugLevel = "high"
	DebugLevelFull   DebugLevel = "full"
)

// DebugConfig represents the configuration for a debug node
type DebugConfig struct {
	Operation   DebugOperationType `json:"operation"`
	Level       DebugLevel         `json:"level"`
	Label       string             `json:"label"`
	Enabled     bool               `json:"enabled"`
	ShowInputs  bool               `json:"show_inputs"`
	ShowOutputs bool               `json:"show_outputs"`
	MeasureTime bool               `json:"measure_time"`
	CaptureStack bool              `json:"capture_stack"`
	MaxDepth    int                `json:"max_depth"`
	FilterKeys  []string           `json:"filter_keys"`
}

// DebugNode represents a debug node
type DebugNode struct {
	config *DebugConfig
}

// NewDebugNode creates a new debug node
func NewDebugNode(config *DebugConfig) *DebugNode {
	if config.Level == "" {
		config.Level = DebugLevelMedium
	}
	if config.MaxDepth == 0 {
		config.MaxDepth = 3
	}

	return &DebugNode{
		config: config,
	}
}

// Execute executes the debug operation
func (dn *DebugNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	if !dn.config.Enabled {
		return map[string]interface{}{
			"success": true,
			"skipped": true,
			"reason":  "debug disabled",
		}, nil
	}

	operation := dn.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = DebugOperationType(opStr)
		}
	}

	startTime := time.Now()
	result := make(map[string]interface{})

	switch operation {
	case DebugOpLog:
		result = dn.logOperation(inputs)
	case DebugOpInspect:
		result = dn.inspectOperation(inputs)
	case DebugOpMeasure:
		result = dn.measureOperation(inputs, startTime)
	case DebugOpTrace:
		result = dn.traceOperation(inputs)
	case DebugOpBreak:
		result = dn.breakOperation(inputs)
	default:
		result = dn.logOperation(inputs) // Default to log
	}

	// Add metadata
	result["operation"] = string(operation)
	result["label"] = dn.config.Label
	result["timestamp"] = time.Now().Unix()

	return result, nil
}

// logOperation logs the debug information
func (dn *DebugNode) logOperation(inputs map[string]interface{}) map[string]interface{} {
	debugInfo := make(map[string]interface{})

	// Add label if provided
	if dn.config.Label != "" {
		debugInfo["label"] = dn.config.Label
	}

	// Add inputs if requested
	if dn.config.ShowInputs {
		debugInfo["inputs"] = dn.filterOrTruncate(inputs, dn.config.MaxDepth)
	}

	// Add system info based on debug level
	switch dn.config.Level {
	case DebugLevelHigh, DebugLevelFull:
		debugInfo["system_info"] = dn.getSystemInfo()
		if dn.config.CaptureStack {
			debugInfo["stack_trace"] = dn.getStackTrace()
		}
		fallthrough
	case DebugLevelMedium:
		debugInfo["memory_stats"] = dn.getMemoryStats()
		fallthrough
	default:
		debugInfo["goroutine_count"] = runtime.NumGoroutine()
	}

	// Print debug info (in real implementation, this would go to a proper logger)
	fmt.Printf("[DEBUG] %s: %+v\n", dn.config.Label, debugInfo)

	return map[string]interface{}{
		"success":    true,
		"operation":  "log",
		"debug_info": debugInfo,
		"level":      string(dn.config.Level),
	}
}

// inspectOperation inspects the data structure
func (dn *DebugNode) inspectOperation(inputs map[string]interface{}) map[string]interface{} {
	inspection := make(map[string]interface{})

	// Inspect each input
	for key, value := range inputs {
		inspection[key] = dn.inspectValue(value, 0)
	}

	// Add metadata
	inspection["input_count"] = len(inputs)
	inspection["timestamp"] = time.Now().Unix()

	fmt.Printf("[INSPECT] %s: %+v\n", dn.config.Label, inspection)

	return map[string]interface{}{
		"success":    true,
		"operation":  "inspect",
		"inspection": inspection,
		"level":      string(dn.config.Level),
	}
}

// inspectValue recursively inspects a value to understand its structure
func (dn *DebugNode) inspectValue(value interface{}, depth int) interface{} {
	if depth >= dn.config.MaxDepth {
		return fmt.Sprintf("[max_depth_reached: %T]", value)
	}

	switch v := value.(type) {
	case nil:
		return map[string]interface{}{
			"type":  "nil",
			"value": nil,
		}
	case bool:
		return map[string]interface{}{
			"type":  "bool",
			"value": v,
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return map[string]interface{}{
			"type":  fmt.Sprintf("%T", v),
			"value": v,
		}
	case float32, float64:
		return map[string]interface{}{
			"type":  fmt.Sprintf("%T", v),
			"value": v,
		}
	case string:
		return map[string]interface{}{
			"type":  "string",
			"value": v,
			"length": len(v),
		}
	case map[string]interface{}:
		result := make(map[string]interface{})
		result["type"] = "map[string]interface{}"
		result["value"] = make(map[string]interface{})
		result["length"] = len(v)
		
		for k, val := range v {
			result["value"].(map[string]interface{})[k] = dn.inspectValue(val, depth+1)
		}
		
		return result
	case []interface{}:
		result := make(map[string]interface{})
		result["type"] = "[]interface{}"
		result["value"] = make([]interface{}, len(v))
		result["length"] = len(v)
		
		for i, val := range v {
			result["value"].([]interface{})[i] = dn.inspectValue(val, depth+1)
		}
		
		return result
	default:
		return map[string]interface{}{
			"type":  fmt.Sprintf("%T", v),
			"value": fmt.Sprintf("%+v", v),
		}
	}
}

// measureOperation measures execution time and performance
func (dn *DebugNode) measureOperation(inputs map[string]interface{}, startTime time.Time) map[string]interface{} {
	elapsed := time.Since(startTime)

	measurements := map[string]interface{}{
		"elapsed_time_ns": elapsed.Nanoseconds(),
		"elapsed_time_ms": elapsed.Milliseconds(),
		"elapsed_time_s":  elapsed.Seconds(),
		"start_time":      startTime.Unix(),
		"end_time":        time.Now().Unix(),
	}

	if dn.config.ShowInputs {
		measurements["inputs"] = dn.filterOrTruncate(inputs, dn.config.MaxDepth)
	}

	// Add memory stats if in high/full mode
	if dn.config.Level == DebugLevelHigh || dn.config.Level == DebugLevelFull {
		measurements["memory_before"] = dn.getMemoryStats()
	}

	fmt.Printf("[MEASURE] %s: %+v\n", dn.config.Label, measurements)

	return map[string]interface{}{
		"success":      true,
		"operation":    "measure",
		"measurements": measurements,
		"level":        string(dn.config.Level),
	}
}

// traceOperation traces execution path
func (dn *DebugNode) traceOperation(inputs map[string]interface{}) map[string]interface{} {
	traceInfo := map[string]interface{}{
		"location": dn.getStackTrace(),
	}

	if dn.config.ShowInputs {
		traceInfo["inputs"] = dn.filterOrTruncate(inputs, dn.config.MaxDepth)
	}

	traceInfo["timestamp"] = time.Now().Unix()
	traceInfo["goroutine_id"] = dn.getGoroutineID()

	fmt.Printf("[TRACE] %s: %+v\n", dn.config.Label, traceInfo)

	return map[string]interface{}{
		"success":   true,
		"operation": "trace",
		"trace":     traceInfo,
		"level":     string(dn.config.Level),
	}
}

// breakOperation simulates a breakpoint (for debugging)
func (dn *DebugNode) breakOperation(inputs map[string]interface{}) map[string]interface{} {
	breakInfo := map[string]interface{}{
		"message": "breakpoint hit",
	}

	if dn.config.ShowInputs {
		breakInfo["inputs"] = dn.filterOrTruncate(inputs, dn.config.MaxDepth)
	}

	breakInfo["location"] = dn.getStackTrace()
	breakInfo["timestamp"] = time.Now().Unix()

	fmt.Printf("[BREAK] %s: %+v\n", dn.config.Label, breakInfo)

	// In a real implementation, this might pause execution
	// For now, we'll just continue

	return map[string]interface{}{
		"success":   true,
		"operation": "break",
		"break":     breakInfo,
		"level":     string(dn.config.Level),
	}
}

// getSystemInfo gets system information
func (dn *DebugNode) getSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"go_version":    runtime.Version(),
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH(),
		"num_cpu":       runtime.NumCPU(),
		"compiler":      runtime.Compiler,
		"timestamp":     time.Now().Unix(),
	}
}

// getMemoryStats gets memory statistics
func (dn *DebugNode) getMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"alloc":      m.Alloc,
		"total_alloc": m.TotalAlloc,
		"sys":        m.Sys,
		"num_gc":     m.NumGC,
		"pause_total_ns": m.PauseTotalNs,
	}
}

// getStackTrace gets the current stack trace
func (dn *DebugNode) getStackTrace() []string {
	var stack []string
	for i := 0; i < 10; i++ { // Limit to 10 stack frames
		pc, file, line, ok := runtime.Caller(i + 2) // Skip this function and its callers
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stack
}

// getGoroutineID gets the current goroutine ID
func (dn *DebugNode) getGoroutineID() int {
	// This is a simplified implementation
	// In a real system, you might want a more sophisticated approach
	return 0
}

// filterOrTruncate filters sensitive keys or truncates deep structures
func (dn *DebugNode) filterOrTruncate(data interface{}, maxDepth int) interface{} {
	// This is a simplified implementation
	// A full implementation would include proper filtering and truncation logic
	return data
}

// DebugNodeFromConfig creates a new debug node from a configuration map
func DebugNodeFromConfig(config map[string]interface{}) (engine.NodeInstance, error) {
	var operation DebugOperationType
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = DebugOperationType(opStr)
		}
	}

	var level DebugLevel
	if lvl, exists := config["level"]; exists {
		if lvlStr, ok := lvl.(string); ok {
			level = DebugLevel(lvlStr)
		}
	}

	var label string
	if lbl, exists := config["label"]; exists {
		if lblStr, ok := lbl.(string); ok {
			label = lblStr
		}
	}

	var enabled bool
	if en, exists := config["enabled"]; exists {
		if enBool, ok := en.(bool); ok {
			enabled = enBool
		}
	} else {
		enabled = true // Default to enabled
	}

	var showInputs bool
	if inputs, exists := config["show_inputs"]; exists {
		if inputsBool, ok := inputs.(bool); ok {
			showInputs = inputsBool
		}
	}

	var showOutputs bool
	if outputs, exists := config["show_outputs"]; exists {
		if outputsBool, ok := outputs.(bool); ok {
			showOutputs = outputsBool
		}
	}

	var measureTime bool
	if time, exists := config["measure_time"]; exists {
		if timeBool, ok := time.(bool); ok {
			measureTime = timeBool
		}
	}

	var captureStack bool
	if stack, exists := config["capture_stack"]; exists {
		if stackBool, ok := stack.(bool); ok {
			captureStack = stackBool
		}
	}

	var maxDepth float64
	if depth, exists := config["max_depth"]; exists {
		if depthFloat, ok := depth.(float64); ok {
			maxDepth = depthFloat
		}
	}

	var filterKeys []string
	if filters, exists := config["filter_keys"]; exists {
		if filtersSlice, ok := filters.([]interface{}); ok {
			filterKeys = make([]string, len(filtersSlice))
			for i, key := range filtersSlice {
				filterKeys[i] = fmt.Sprintf("%v", key)
			}
		}
	}

	nodeConfig := &DebugConfig{
		Operation:    operation,
		Level:        level,
		Label:        label,
		Enabled:      enabled,
		ShowInputs:   showInputs,
		ShowOutputs:  showOutputs,
		MeasureTime:  measureTime,
		CaptureStack: captureStack,
		MaxDepth:     int(maxDepth),
		FilterKeys:   filterKeys,
	}

	return NewDebugNode(nodeConfig), nil
}

// RegisterDebugNode registers the debug node type with the engine
func RegisterDebugNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("debug", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return DebugNodeFromConfig(config)
	})
}