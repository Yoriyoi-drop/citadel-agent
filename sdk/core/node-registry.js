// citadel-agent/sdk/core/node-registry.js
/**
 * Node Registry - Registry for managing available node types
 */

class NodeRegistry {
  constructor() {
    this.nodeTypes = new Map(); // nodeId -> nodeClass
    this.nodeMetadata = new Map(); // nodeId -> metadata
    this.categories = {}; // category -> [nodeIds]
    this.tags = {}; // tag -> [nodeIds]
    this.builtInNodes = new Set();
    
    // Register built-in node types
    this.registerBuiltInNodes();
  }

  /**
   * Registers a built-in node type
   */
  registerBuiltInNodes() {
    // Register core built-in nodes here
    // This will be expanded as we create the actual node implementations
  }

  /**
   * Registers a new node type
   */
  registerNode(nodeId, NodeClass, metadata = {}) {
    // Validate inputs
    if (typeof nodeId !== 'string' || !nodeId) {
      throw new Error('Node ID must be a non-empty string');
    }
    
    if (typeof NodeClass !== 'function') {
      throw new Error('NodeClass must be a constructor function');
    }
    
    // Validate required metadata
    if (!metadata.name) {
      metadata.name = NodeClass.name.replace('Node', '');
    }
    
    if (!metadata.description) {
      metadata.description = \`Node for \${metadata.name}\`;
    }
    
    // Store node class
    this.nodeTypes.set(nodeId, NodeClass);
    
    // Store metadata
    const nodeMetadata = {
      id: nodeId,
      name: metadata.name,
      description: metadata.description,
      category: metadata.category || 'custom',
      type: metadata.type || 'custom',
      version: metadata.version || '1.0.0',
      author: metadata.author || 'Anonymous',
      tags: Array.isArray(metadata.tags) ? metadata.tags : [],
      schema: metadata.schema || {},
      builtin: !!metadata.builtin,
      createdAt: new Date(),
      updatedAt: new Date()
    };
    
    this.nodeMetadata.set(nodeId, nodeMetadata);
    
    // Add to category index
    const category = nodeMetadata.category;
    if (!this.categories[category]) {
      this.categories[category] = [];
    }
    if (!this.categories[category].includes(nodeId)) {
      this.categories[category].push(nodeId);
    }
    
    // Add to tag indexes
    if (nodeMetadata.tags) {
      for (const tag of nodeMetadata.tags) {
        if (!this.tags[tag]) {
          this.tags[tag] = [];
        }
        if (!this.tags[tag].includes(nodeId)) {
          this.tags[tag].push(nodeId);
        }
      }
    }
    
    // Mark as built-in if specified
    if (nodeMetadata.builtin) {
      this.builtInNodes.add(nodeId);
    }
    
    return true;
  }

  /**
   * Unregisters a node type
   */
  unregisterNode(nodeId) {
    if (!this.nodeTypes.has(nodeId)) {
      return false;
    }
    
    const metadata = this.nodeMetadata.get(nodeId);
    
    // Remove from node type map
    this.nodeTypes.delete(nodeId);
    
    // Remove from metadata map
    this.nodeMetadata.delete(nodeId);
    
    // Remove from category index
    if (metadata && metadata.category) {
      const categoryIndex = this.categories[metadata.category]?.indexOf(nodeId);
      if (categoryIndex !== -1) {
        this.categories[metadata.category].splice(categoryIndex, 1);
      }
    }
    
    // Remove from tag indexes
    if (metadata && metadata.tags) {
      for (const tag of metadata.tags) {
        const tagIndex = this.tags[tag]?.indexOf(nodeId);
        if (tagIndex !== -1) {
          this.tags[tag].splice(tagIndex, 1);
        }
      }
    }
    
    // Remove from built-in set
    this.builtInNodes.delete(nodeId);
    
    return true;
  }

  /**
   * Gets a node class by ID
   */
  getNodeType(nodeId) {
    return this.nodeTypes.get(nodeId) || null;
  }

  /**
   * Gets node metadata
   */
  getMetadata(nodeId) {
    return this.nodeMetadata.get(nodeId) || null;
  }

  /**
   * Gets all node IDs
   */
  getNodeIds() {
    return Array.from(this.nodeTypes.keys());
  }

  /**
   * Gets all metadata
   */
  getAllMetadata() {
    return Array.from(this.nodeMetadata.values());
  }

  /**
   * Gets nodes by category
   */
  getByCategory(category) {
    const nodeIds = this.categories[category] || [];
    return nodeIds.map(id => this.nodeMetadata.get(id)).filter(Boolean);
  }

  /**
   * Gets nodes by tag
   */
  getByTag(tag) {
    const nodeIds = this.tags[tag] || [];
    return nodeIds.map(id => this.nodeMetadata.get(id)).filter(Boolean);
  }

  /**
   * Searches for nodes by name, description, or tags
   */
  search(query) {
    if (!query) return this.getAllMetadata();
    
    const term = query.toLowerCase();
    return this.getAllMetadata().filter(meta =>
      meta.name.toLowerCase().includes(term) ||
      meta.description.toLowerCase().includes(term) ||
      meta.tags.some(tag => tag.toLowerCase().includes(term))
    );
  }

  /**
   * Gets all categories
   */
  getCategories() {
    return Object.keys(this.categories);
  }

  /**
   * Gets all tags
   */
  getTags() {
    return Object.keys(this.tags);
  }

  /**
   * Checks if a node is built-in
   */
  isBuiltIn(nodeId) {
    return this.builtInNodes.has(nodeId);
  }

  /**
   * Gets registry statistics
   */
  getStats() {
    const allNodes = this.getAllMetadata();
    
    return {
      totalNodes: allNodes.length,
      builtinNodes: allNodes.filter(n => n.builtin).length,
      customNodes: allNodes.filter(n => !n.builtin).length,
      categories: Object.keys(this.categories).length,
      tags: Object.keys(this.tags).length,
      perCategory: Object.entries(this.categories).map(([cat, nodes]) => ({ 
        category: cat, 
        count: nodes.length 
      }))
    };
  }
}

module.exports = { NodeRegistry };