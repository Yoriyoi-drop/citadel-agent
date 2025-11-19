// citadel-agent/sdk/core/node-factory.js
/**
 * NodeFactory - Factory for creating node instances
 * Handles the creation and registration of nodes
 */

const { NodeExecutor } = require('./node-executor');
const { NodeRegistry } = require('./node-registry');

class NodeFactory {
  constructor() {
    this.registry = new NodeRegistry();
    this.nodeInstances = new Map(); // Maps instanceId -> instance
  }

  /**
   * Creates a new node instance
   */
  createNode(nodeId, config = {}) {
    // Look up the node class in the registry
    const NodeClass = this.registry.getNodeType(nodeId);
    if (!NodeClass) {
      // If not found in registry, try to register it dynamically
      // This could involve loading from a file or database
      throw new Error(\`Node type '\${nodeId}' is not registered in the registry\`);
    }

    // Create instance with config
    const instance = new NodeClass(config);
    
    // Add to registry of active instances
    const instanceId = config.id || \`\${nodeId}-\${Date.now()}-\${Math.random().toString(36).substring(2, 9)}\`;
    this.nodeInstances.set(instanceId, {
      id: instanceId,
      nodeId,
      instance,
      created: new Date()
    });

    return instance;
  }

  /**
   * Creates multiple node instances at once
   */
  createNodes(nodeDefs) {
    const instances = {};
    
    for (const nodeDef of nodeDefs) {
      const { id, type, config } = nodeDef;
      instances[id] = this.createNode(type, config);
    }
    
    return instances;
  }

  /**
   * Creates a node from a workflow definition
   */
  createNodeFromDefinition(nodeDef) {
    const { id, type, settings = {} } = nodeDef;
    return this.createNode(type, { ...settings, id });
  }

  /**
   * Gets a node instance by instance ID
   */
  getNodeInstance(instanceId) {
    const nodeEntry = this.nodeInstances.get(instanceId);
    return nodeEntry ? nodeEntry.instance : null;
  }

  /**
   * Disposes of a node instance
   */
  async disposeNode(instanceId) {
    const nodeEntry = this.nodeInstances.get(instanceId);
    if (!nodeEntry) {
      return false;
    }

    // Call cleanup method if available
    if (nodeEntry.instance && typeof nodeEntry.instance.destroy === 'function') {
      try {
        await nodeEntry.instance.destroy();
      } catch (error) {
        console.error(\`Error cleaning up node instance \${instanceId}:\`, error);
      }
    }

    // Remove from registry
    this.nodeInstances.delete(instanceId);
    return true;
  }

  /**
   * Disposes all node instances
   */
  async disposeAllNodes() {
    const promises = [];
    
    for (const instanceId of this.nodeInstances.keys()) {
      promises.push(this.disposeNode(instanceId));
    }
    
    await Promise.all(promises);
  }

  /**
   * Registers a custom node type
   */
  registerNode(nodeId, NodeClass, metadata = {}) {
    return this.registry.registerNode(nodeId, NodeClass, metadata);
  }

  /**
   * Unregisters a node type
   */
  unregisterNode(nodeId) {
    return this.registry.unregisterNode(nodeId);
  }

  /**
   * Gets node metadata
   */
  getNodeMetadata(nodeId) {
    return this.registry.getMetadata(nodeId);
  }

  /**
   * Gets all available node types
   */
  getAvailableNodeTypes() {
    return this.registry.getNodeIds();
  }

  /**
   * Gets nodes by category
   */
  getNodesByCategory(category) {
    return this.registry.getByCategory(category);
  }

  /**
   * Gets nodes by tag
   */
  getNodesByTag(tag) {
    return this.registry.getByTag(tag);
  }

  /**
   * Searches for nodes
   */
  searchNodes(query) {
    return this.registry.search(query);
  }

  /**
   * Gets registry stats
   */
  getRegistryStats() {
    return this.registry.getStats();
  }
}
}

// Built-in node classes that NodeFactory should know about initially
const builtInNodes = {
  // Http Request Node
  'http_request': class HttpRequestNode extends NodeExecutor {
    async process(input) {
      // Simulating HTTP request functionality
      const config = this.config || {};
      const url = input.url || config.url || 'http://localhost';
      const method = input.method || config.method || 'GET';
      
      // In a real implementation, this would make the actual HTTP request
      return {
        statusCode: 200,
        statusText: 'OK',
        url: url,
        method: method,
        data: input.body || input.data || {},
        headers: { 'content-type': 'application/json' },
        timestamp: new Date().toISOString()
      };
    }
  },

  // Data Transformer Node
  'data_transformer': class DataTransformerNode extends NodeExecutor {
    async process(input) {
      const config = this.config || {};
      const operations = config.operations || [];
      
      let transformed = { ...input };
      
      for (const operation of operations) {
        switch (operation.type) {
          case 'json_to_object':
            if (typeof transformed.data === 'string') {
              try {
                transformed.data = JSON.parse(transformed.data);
              } catch (e) {
                // Invalid JSON, leave as string
              }
            }
            break;
          case 'object_to_json':
            if (typeof transformed.data === 'object') {
              transformed.data = JSON.stringify(transformed.data);
            }
            break;
          case 'field_mapping':
            if (operation.mapping && typeof transformed.data === 'object') {
              const mapped = {};
              for (const [source, target] of Object.entries(operation.mapping)) {
                const sourceVal = this.getNestedValue(transformed.data, source);
                this.setNestedValue(mapped, target, sourceVal);
              }
              transformed.data = mapped;
            }
            break;
          case 'filter':
            if (operation.field && operation.condition) {
              // Filter array data based on condition
              if (Array.isArray(transformed.data)) {
                transformed.data = transformed.data.filter(item => {
                  const fieldValue = this.getNestedValue(item, operation.field);
                  switch (operation.condition.operator) {
                    case 'equals':
                      return fieldValue === operation.condition.value;
                    case 'contains':
                      return String(fieldValue).includes(String(operation.condition.value));
                    default:
                      return true;
                  }
                });
              }
            }
            break;
        }
      }
      
      return {
        transformed: transformed,
        operationCount: operations.length,
        timestamp: new Date().toISOString()
      };
    }
    
    getNestedValue(obj, path) {
      const keys = path.split('.');
      let value = obj;
      for (const key of keys) {
        if (value === null || value === undefined) return undefined;
        value = value[key];
      }
      return value;
    }
    
    setNestedValue(obj, path, value) {
      const keys = path.split('.');
      let current = obj;
      for (let i = 0; i < keys.length - 1; i++) {
        if (current[keys[i]] === undefined) {
          current[keys[i]] = {};
        }
        current = current[keys[i]];
      }
      current[keys[keys.length - 1]] = value;
    }
  },

  // Condition Node
  'condition': class ConditionNode extends NodeExecutor {
    async process(input) {
      const config = this.config || {};
      
      if (!config.conditions || !Array.isArray(config.conditions) || config.conditions.length === 0) {
        // If no conditions are provided, treat as always true
        return {
          match: true,
          conditionIndex: null,
          output: input,
          timestamp: new Date().toISOString()
        };
      }
      
      // Evaluate each condition
      for (let i = 0; i < config.conditions.length; i++) {
        const condition = config.conditions[i];
        let result = false;
        
        if (condition.type === 'expression') {
          // In real implementation, this would safely evaluate the expression
          // For now, use simple comparisons
          const { left, operator, right } = condition.value;
          const leftValue = this.getValue(input, left);
          const rightValue = this.getValue(input, right);
          
          switch (operator) {
            case 'equals':
              result = leftValue == rightValue;
              break;
            case 'not_equals':
              result = leftValue != rightValue;
              break;
            case 'greater_than':
              result = leftValue > rightValue;
              break;
            case 'less_than':
              result = leftValue < rightValue;
              break;
            case 'contains':
              result = String(leftValue).includes(String(rightValue));
              break;
            case 'starts_with':
              result = String(leftValue).startsWith(String(rightValue));
              break;
            case 'ends_with':
              result = String(leftValue).endsWith(String(rightValue));
              break;
            default:
              result = leftValue == rightValue; // Default to equals
          }
        }
        
        if (result) {
          return {
            match: true,
            conditionIndex: i,
            output: input,
            condition: condition,
            timestamp: new Date().toISOString()
          };
        }
      }
      
      // No conditions matched
      return {
        match: false,
        conditionIndex: null,
        output: input,
        timestamp: new Date().toISOString()
      };
    }
    
    getValue(input, path) {
      if (typeof path === 'string' && path.startsWith('{{') && path.endsWith('}}')) {
        // Handle template expressions like {{input.value}}
        const varPath = path.substring(2, path.length - 2);
        const parts = varPath.split('.');
        let value = input;
        
        for (const part of parts) {
          if (value && value[part] !== undefined) {
            value = value[part];
          } else {
            return undefined;
          }
        }
        
        return value;
      }
      
      // If it's not a template expression, return the path as a literal value
      return path;
    }
  },

  // Loop Node
  'loop': class LoopNode extends NodeExecutor {
    async process(input) {
      const config = this.config || {};
      const items = input.items || config.items || [];
      const iterationVar = config.iterationVar || 'item';
      const indexVar = config.indexVar || 'index';
      
      if (!Array.isArray(items)) {
        throw new Error('Loop node requires an array of items to iterate over');
      }
      
      const results = [];
      
      // In a real implementation, this would execute the loop body
      // For now, we'll just return the processed items
      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        const loopContext = {
          [iterationVar]: item,
          [indexVar]: i,
          total: items.length,
          input: input
        };
        
        // This would process the loop body in a real implementation
        results.push({
          index: i,
          item: item,
          context: loopContext
        });
      }
      
      return {
        count: results.length,
        results: results,
        originalItems: items,
        timestamp: new Date().toISOString()
      };
    }
  }
};

// Register built-in nodes
const nodeFactory = new NodeFactory();

// Register all built-in nodes
for (const [nodeId, NodeClass] of Object.entries(builtInNodes)) {
  nodeFactory.registerNode(nodeId, NodeClass, {
    id: nodeId,
    name: NodeClass.name.replace('Node', ''),
    description: \`Built-in \${nodeId.replace('_', ' ')} node\`,
    category: 'core',
    type: 'builtin',
    builtin: true,
    version: '1.0.0'
  });
}

module.exports = { NodeFactory };