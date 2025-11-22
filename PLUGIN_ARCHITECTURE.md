# Plugin System for Citadel Agent

## Overview
Citadel Agent now supports a plugin system that allows nodes to be loaded as separate processes. This provides better isolation, security, and modularity for the workflow engine.

## Architecture
- `NodePlugin`: Interface defining how nodes interact with the plugin system
- `NodeManager`: Manages the lifecycle of plugin processes
- `NodeInstanceAdapter`: Adapts legacy NodeInstance implementations to the plugin interface
- `PluginAwareNodeRegistry`: Registry that can handle both local and plugin nodes
- `PluginAwareEngine`: Workflow engine that supports both local and plugin nodes

## Creating a New Plugin

### 1. Define your plugin struct
```go
type MyPlugin struct{}

func (m *MyPlugin) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Your plugin logic here
    return map[string]interface{}{"result": "success"}, nil
}

func (m *MyPlugin) GetConfigSchema() map[string]interface{} {
    // Return JSON schema for configuration
    return map[string]interface{}{...}
}

func (m *MyPlugin) GetMetadata() plugins.NodeMetadata {
    return plugins.NodeMetadata{
        ID:          "my_plugin",
        Name:        "My Plugin",
        Description: "A sample plugin",
        Version:     "1.0.0",
        Author:      "Your Name",
        Category:    "utility",
    }
}
```

### 2. Implement the main function
```go
func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugins.Handshake,
        Plugins: map[string]plugin.Plugin{
            "node": &plugins.NodePluginImpl{Impl: &MyPlugin{}},
        },
    })
}
```

### 3. Build the plugin
```bash
go build -o my_plugin ./plugins/my_plugin.go
```

### 4. Register the plugin with the engine
```go
// In your application code
pluginManager := plugins.NewNodeManager()
err := pluginManager.RegisterPluginAtPath("my_plugin", "./my_plugin")
if err != nil {
    log.Fatal(err)
}

pluginAwareEngine := plugins.NewPluginAwareEngine(baseEngine, pluginManager)
err = pluginAwareEngine.RegisterPluginNodeType("my_plugin")
```

## Migration Path

The system maintains backward compatibility, so you can:
1. Keep existing local nodes alongside new plugin nodes
2. Gradually migrate nodes to plugins as needed
3. Use the same workflow definition format with both local and plugin nodes
4. Register plugins with the same IDs as local node types for seamless transition

## Benefits

- **Isolation**: Each plugin runs in its own process
- **Security**: Sandboxed execution environment
- **Modularity**: Easy to add/upgrade nodes without rebuilding the entire application
- **Language Agnostic**: Plugins can be written in other languages (with appropriate adapters)
- **Hot Reload**: Plugins can be updated without restarting the main application