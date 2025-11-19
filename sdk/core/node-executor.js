// citadel-agent/sdk/core/node-executor.js
/**
 * NodeExecutor - Base class for all workflow nodes
 * Provides the core execution interface and common functionality
 */

class NodeExecutor {
  constructor(config = {}) {
    this.id = config.id || this.constructor.name.toLowerCase().replace(/node$/, '');
    this.config = { ...config };
    this.initialized = false;
    this.createdAt = new Date();
    this.updatedAt = new Date();
    
    // Execution statistics
    this.stats = {
      executions: 0,
      errors: 0,
      totalExecutionTime: 0,
      avgExecutionTime: 0
    };
  }

  /**
   * Initializes the node before first execution
   * Override this method in subclasses for custom initialization logic
   */
  async initialize() {
    if (this.initialized) return;
    
    // Perform any initialization logic
    this.initialized = true;
    this.updatedAt = new Date();
  }

  /**
   * Validates the node configuration
   * Override this method in subclasses to provide specific validation
   * @returns {Array<string>} Array of validation errors, empty if none
   */
  validateConfig() {
    return [];
  }

  /**
   * Main execution method - this should be called externally
   * Subclasses should implement process() method instead
   */
  async execute(input) {
    const startTime = Date.now();
    this.stats.executions++;
    
    try {
      await this.initialize();
      
      // Validate config
      const validationErrors = this.validateConfig();
      if (validationErrors.length > 0) {
        throw new Error(\`Configuration validation failed: \${validationErrors.join('; ')}\`);
      }
      
      // Execute the actual processing logic
      const result = await this.process(input);
      
      // Update statistics
      const executionTime = Date.now() - startTime;
      this.updateStats(executionTime);
      
      return {
        status: 'success',
        data: result,
        metadata: {
          executionTime,
          node: this.constructor.name,
          nodeId: this.id,
          timestamp: new Date().toISOString(),
          stats: { ...this.stats }
        }
      };
    } catch (error) {
      this.stats.errors++;
      const executionTime = Date.now() - startTime;
      this.updateStats(executionTime);
      
      return {
        status: 'error',
        error: error.message,
        stack: error.stack,
        metadata: {
          executionTime,
          node: this.constructor.name,
          nodeId: this.id,
          timestamp: new Date().toISOString(),
          stats: { ...this.stats }
        }
      };
    }
  }

  /**
   * Core processing logic - must be implemented by subclasses
   */
  async process(input) {
    throw new Error(\`process() method must be implemented by subclass \${this.constructor.name}\`);
  }

  /**
   * Updates execution statistics
   */
  updateStats(executionTime) {
    this.stats.totalExecutionTime += executionTime;
    this.stats.avgExecutionTime = this.stats.totalExecutionTime / this.stats.executions;
  }

  /**
   * Gets execution statistics
   */
  getStats() {
    return { ...this.stats };
  }

  /**
   * Cleans up resources when node is destroyed
   * Override this method in subclasses for cleanup logic
   */
  async destroy() {
    this.updatedAt = new Date();
  }

  /**
   * Gets node information
   */
  getInfo() {
    return {
      id: this.id,
      name: this.constructor.name,
      initialized: this.initialized,
      createdAt: this.createdAt,
      updatedAt: this.updatedAt,
      stats: this.getStats()
    };
  }
}

module.exports = { NodeExecutor };