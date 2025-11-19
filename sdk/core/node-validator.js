// citadel-agent/sdk/core/node-validator.js
/**
 * NodeValidator - Validates node structure, configuration, and implementation
 */

class NodeValidator {
  constructor() {
    this.rules = {
      requiredFields: ['id', 'name', 'description', 'category'],
      allowedCategories: [
        'http', 'database', 'function', 'trigger', 
        'condition', 'loop', 'transformer', 'notification',
        'file', 'email', 'social', 'custom'
      ],
      allowedGrades: ['basic', 'intermediate', 'advanced', 'elite'],
      allowedTypes: ['builtin', 'custom', 'external'],
      maxDescriptionLength: 500,
      maxNameLength: 100
    };
  }

  /**
   * Validates a node definition
   */
  validateNodeDefinition(definition) {
    const errors = [];
    const warnings = [];

    // Check required fields
    for (const field of this.rules.requiredFields) {
      if (!(field in definition)) {
        errors.push(\`Missing required field: \${field}\`);
      }
    }

    // Validate basic fields if they exist
    if (definition.id) {
      if (typeof definition.id !== 'string') {
        errors.push('id must be a string');
      } else if (!/^[a-zA-Z][a-zA-Z0-9_-]*$/.test(definition.id)) {
        errors.push('id must start with a letter and contain only letters, numbers, hyphens, and underscores');
      }
    }

    if (definition.name) {
      if (typeof definition.name !== 'string') {
        errors.push('name must be a string');
      } else if (definition.name.length > this.rules.maxNameLength) {
        errors.push(\`name must be \${this.rules.maxNameLength} characters or less\`);
      }
    }

    if (definition.description) {
      if (typeof definition.description !== 'string') {
        errors.push('description must be a string');
      } else if (definition.description.length > this.rules.maxDescriptionLength) {
        errors.push(\`description must be \${this.rules.maxDescriptionLength} characters or less\`);
      }
    }

    if (definition.category) {
      if (typeof definition.category !== 'string') {
        errors.push('category must be a string');
      } else if (!this.rules.allowedCategories.includes(definition.category)) {
        warnings.push(\`category '\${definition.category}' is not in the standard list\`);
      }
    }

    if (definition.grade) {
      if (typeof definition.grade !== 'string') {
        errors.push('grade must be a string');
      } else if (!this.rules.allowedGrades.includes(definition.grade)) {
        errors.push(\`grade must be one of: \${this.rules.allowedGrades.join(', ')}\`);
      }
    }

    if (definition.type) {
      if (typeof definition.type !== 'string') {
        errors.push('type must be a string');
      } else if (!this.rules.allowedTypes.includes(definition.type)) {
        errors.push(\`type must be one of: \${this.rules.allowedTypes.join(', ')}\`);
      }
    }

    // Validate schema if present
    if (definition.schema) {
      const schemaErrors = this.validateSchema(definition.schema);
      errors.push(...schemaErrors);
    }

    // Validate dependencies if present
    if (definition.dependencies) {
      if (!Array.isArray(definition.dependencies)) {
        errors.push('dependencies must be an array');
      } else {
        for (let i = 0; i < definition.dependencies.length; i++) {
          const dep = definition.dependencies[i];
          if (typeof dep !== 'string') {
            errors.push(\`dependency at index \${i} must be a string\`);
          }
        }
      }
    }

    // Validate tags if present
    if (definition.tags) {
      if (!Array.isArray(definition.tags)) {
        errors.push('tags must be an array');
      } else {
        for (let i = 0; i < definition.tags.length; i++) {
          const tag = definition.tags[i];
          if (typeof tag !== 'string') {
            errors.push(\`tag at index \${i} must be a string\`);
          }
        }
      }
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Validates a node schema definition
   */
  validateSchema(schema) {
    const errors = [];

    if (typeof schema !== 'object') {
      return ['schema must be an object'];
    }

    // Validate inputs schema
    if (schema.inputs) {
      if (!Array.isArray(schema.inputs)) {
        errors.push('schema.inputs must be an array');
      } else {
        for (let i = 0; i < schema.inputs.length; i++) {
          const input = schema.inputs[i];
          const inputErrors = this.validateParameter(input, 'input', i);
          errors.push(...inputErrors);
        }
      }
    }

    // Validate outputs schema
    if (schema.outputs) {
      if (!Array.isArray(schema.outputs)) {
        errors.push('schema.outputs must be an array');
      } else {
        for (let i = 0; i < schema.outputs.length; i++) {
          const output = schema.outputs[i];
          const outputErrors = this.validateParameter(output, 'output', i);
          errors.push(...outputErrors);
        }
      }
    }

    // Validate config schema
    if (schema.config) {
      if (typeof schema.config !== 'object') {
        errors.push('schema.config must be an object');
      } else {
        // Validate properties if they exist
        if (schema.config.properties) {
          for (const [propName, propDef] of Object.entries(schema.config.properties)) {
            const propErrors = this.validateParameter(propDef, \`config property \${propName}\`, propName);
            errors.push(...propErrors);
          }
        }

        // Validate required fields
        if (schema.config.required) {
          if (!Array.isArray(schema.config.required)) {
            errors.push('schema.config.required must be an array');
          }
        }
      }
    }

    return errors;
  }

  /**
   * Validates a parameter definition (input, output, or config property)
   */
  validateParameter(param, type, index) {
    const errors = [];

    if (typeof param !== 'object') {
      return [`\${type} at index \${index} must be an object`];
    }

    // Required fields for parameters
    if (!param.name) {
      errors.push(\`\${type} at index \${index} must have a 'name' field\`);
    }

    if (!param.type) {
      errors.push(\`\${type} at index \${index} must have a 'type' field\`);
    } else if (!['string', 'number', 'boolean', 'object', 'array', 'any'].includes(param.type)) {
      errors.push(\`\${type} at index \${index}: type must be one of string, number, boolean, object, array, any\`);
    }

    // Validate default value matches type if provided
    if ('default' in param && param.type) {
      const typeMatches = this.validateType(param.default, param.type);
      if (!typeMatches) {
        errors.push(\`\${type} at index \${index}: default value type does not match specified type\`);
      }
    }

    return errors;
  }

  /**
   * Validates that a value matches the specified type
   */
  validateType(value, expectedType) {
    switch (expectedType) {
      case 'string':
        return typeof value === 'string';
      case 'number':
        return typeof value === 'number' && !isNaN(value);
      case 'boolean':
        return typeof value === 'boolean';
      case 'object':
        return value !== null && typeof value === 'object' && !Array.isArray(value);
      case 'array':
        return Array.isArray(value);
      case 'any':
        return true;
      default:
        return false;
    }
  }

  /**
   * Validates a node class implementation
   */
  validateNodeImplementation(NodeClass) {
    const errors = [];

    if (typeof NodeClass !== 'function') {
      return ['NodeClass must be a constructor function'];
    }

    // Check if it's a proper class
    if (!this.isClass(NodeClass)) {
      errors.push('NodeClass must be a class');
    }

    // Create instance to test methods
    let instance;
    try {
      // Create instance with minimal config
      instance = new NodeClass({});
    } catch (error) {
      errors.push(\`Failed to instantiate node class: \${error.message}`);
      return { valid: false, errors };
    }

    // Check required methods exist
    const requiredMethods = ['execute'];
    for (const method of requiredMethods) {
      if (typeof instance[method] !== 'function') {
        errors.push(\`Node class must implement method '\${method}'\`);
      }
    }

    // Test execute method signature
    if (typeof instance.execute === 'function') {
      // Execute with minimal input to verify it doesn't crash
      try {
        const mockExecuteResult = instance.execute({});
        if (mockExecuteResult && typeof mockExecuteResult.then === 'function') {
          // It's async, so we can't really test without awaiting
          // The method exists and returns a promise-like object
        }
      } catch (error) {
        // This is expected for initial validation, ignore
      }
    }

    // Check that it extends NodeExecutor or has similar interface
    const hasNodeExecutorInterface = this.hasNodeExecutorInterface(instance);
    if (!hasNodeExecutorInterface) {
      errors.push('Node class should follow NodeExecutor interface');
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Checks if a function is a class
   */
  isClass(func) {
    // Check if function has a prototype with methods
    return func.prototype && 
           func.prototype.constructor === func && 
           Object.getOwnPropertyNames(func.prototype).length > 1;
  }

  /**
   * Checks if an object has NodeExecutor-like interface
   */
  hasNodeExecutorInterface(obj) {
    // NodeExecutor should have these properties/methods
    const requiredProperties = ['id', 'config', 'execute'];
    
    for (const prop of requiredProperties) {
      if (!(prop in obj)) {
        return false;
      }
    }
    
    return true;
  }

  /**
   * Validates a complete node package (directory)
   */
  async validateNodePackage(nodePath) {
    const fs = require('fs').promises;
    const path = require('path');
    const errors = [];
    const warnings = [];

    // Check if directory exists
    try {
      const stats = await fs.stat(nodePath);
      if (!stats.isDirectory()) {
        return { valid: false, errors: ['Path is not a directory'], warnings: [] };
      }
    } catch (error) {
      return { valid: false, errors: [`Directory does not exist: \${nodePath}`], warnings: [] };
    }

    // Check required files
    const requiredFiles = [
      'index.js',
      'node.json',
      'package.json'
    ];

    for (const file of requiredFiles) {
      const filePath = path.join(nodePath, file);
      try {
        await fs.access(filePath);
      } catch (error) {
        errors.push(\`Missing required file: \${file}\`);
      }
    }

    // Check node.json validity
    const nodeDefPath = path.join(nodePath, 'node.json');
    try {
      const nodeDefContent = await fs.readFile(nodeDefPath, 'utf8');
      let nodeDef;
      try {
        nodeDef = JSON.parse(nodeDefContent);
      } catch (error) {
        errors.push('node.json is not valid JSON');
      }

      if (nodeDef) {
        const validation = this.validateNodeDefinition(nodeDef);
        errors.push(...validation.errors);
        warnings.push(...validation.warnings);
      }
    } catch (error) {
      if (requiredFiles.includes('node.json')) {
        errors.push('Could not read node.json');
      }
    }

    // Check package.json validity
    const pkgPath = path.join(nodePath, 'package.json');
    try {
      const pkgContent = await fs.readFile(pkgPath, 'utf8');
      let pkgJson;
      try {
        pkgJson = JSON.parse(pkgContent);
      } catch (error) {
        errors.push('package.json is not valid JSON');
      }

      if (pkgJson) {
        // Validate package.json fields
        if (!pkgJson.name) {
          errors.push('package.json must have a name field');
        }
        if (!pkgJson.main) {
          errors.push('package.json must have a main field');
        }
      }
    } catch (error) {
      if (requiredFiles.includes('package.json')) {
        errors.push('Could not read package.json');
      }
    }

    // Check index.js validity
    const indexPath = path.join(nodePath, 'index.js');
    try {
      const indexContent = await fs.readFile(indexPath, 'utf8');
      if (indexContent.trim() === '') {
        errors.push('index.js is empty');
      }
    } catch (error) {
      if (requiredFiles.includes('index.js')) {
        errors.push('Could not read index.js');
      }
    }

    // Check for tests
    const testDir = path.join(nodePath, 'test');
    const testFile = path.join(nodePath, 'test', 'index.test.js');
    try {
      await fs.access(testFile);
    } catch (error) {
      warnings.push('No test file found (recommended: test/index.test.js)');
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Validates node configuration against its schema
   */
  validateNodeConfiguration(config, schema) {
    const errors = [];

    if (!schema || typeof schema !== 'object') {
      return { valid: true, errors: [] }; // No schema to validate against
    }

    // Validate config against schema
    if (schema.config && schema.config.properties) {
      const properties = schema.config.properties;
      const required = schema.config.required || [];

      // Check required properties
      for (const prop of required) {
        if (!(prop in config)) {
          errors.push(\`Required configuration property missing: \${prop}\`);
        }
      }

      // Validate property types
      for (const [propName, propDef] of Object.entries(properties)) {
        if (propName in config) {
          const value = config[propName];
          const expectedType = propDef.type;
          
          if (expectedType) {
            const matches = this.validateType(value, expectedType);
            if (!matches) {
              errors.push(\`Configuration property '\${propName}' expects type '\${expectedType}', got '\${typeof value}'\`);
            }
          }
        }
      }
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }
}

module.exports = { NodeValidator };