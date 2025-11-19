// citadel-agent/sdk/core/node-manifest.js
/**
 * NodeManifest - Manages node metadata and discovery
 */

const fs = require('fs').promises;
const path = require('path');

class NodeManifest {
  constructor(basePath = './nodes') {
    this.basePath = basePath;
    this.manifest = {
      version: '1.0',
      nodes: {},
      categories: {},
      tags: {},
      dependencies: {},
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
  }

  /**
   * Loads manifest from file if it exists
   */
  async load() {
    const manifestPath = path.join(this.basePath, 'manifest.json');
    
    try {
      const manifestContent = await fs.readFile(manifestPath, 'utf8');
      this.manifest = JSON.parse(manifestContent);
      return this.manifest;
    } catch (error) {
      // If file doesn't exist, return default manifest
      if (error.code === 'ENOENT') {
        return this.manifest;
      }
      throw error;
    }
  }

  /**
   * Discovers nodes in the specified directory
   */
  async discoverNodes(scanPath = null) {
    const scanDir = scanPath || this.basePath;
    
    // Reset manifest
    this.manifest.nodes = {};
    this.manifest.categories = {};
    this.manifest.tags = {};

    if (!(await fs.access(scanDir).then(() => true).catch(() => false))) {
      console.warn(\`Scan directory does not exist: \${scanDir}\`);
      return this.manifest;
    }

    const items = await fs.readdir(scanDir);
    
    for (const item of items) {
      const itemPath = path.join(scanDir, item);
      const stat = await fs.stat(itemPath);
      
      if (stat.isDirectory()) {
        const nodePath = itemPath;
        const nodeDefFile = path.join(nodePath, 'node.json');
        
        if (await fs.access(nodeDefFile).then(() => true).catch(() => false)) {
          try {
            const nodeDefinition = JSON.parse(await fs.readFile(nodeDefFile, 'utf8'));
            
            // Validate and normalize node definition
            const normalizedDef = this.normalizeNodeDefinition(nodeDefinition, nodePath);
            
            // Add to manifest
            this.manifest.nodes[normalizedDef.id] = normalizedDef;
            
            // Index by category
            const category = normalizedDef.category || 'custom';
            if (!this.manifest.categories[category]) {
              this.manifest.categories[category] = [];
            }
            this.manifest.categories[category].push(normalizedDef.id);
            
            // Index by tags
            if (normalizedDef.tags && Array.isArray(normalizedDef.tags)) {
              for (const tag of normalizedDef.tags) {
                if (!this.manifest.tags[tag]) {
                  this.manifest.tags[tag] = [];
                }
                this.manifest.tags[tag].push(normalizedDef.id);
              }
            }
            
            console.log(\`Discovered node: \${normalizedDef.name} (\${normalizedDef.id})\`);
          } catch (error) {
            console.error(\`Failed to load node definition from \${nodeDefFile}:\`, error.message);
          }
        }
      }
    }
    
    this.manifest.updatedAt = new Date().toISOString();
    
    return this.manifest;
  }

  /**
   * Normalizes a node definition
   */
  normalizeNodeDefinition(def, nodePath) {
    return {
      id: def.id || path.basename(nodePath),
      name: def.name || def.id || path.basename(nodePath),
      description: def.description || '',
      version: def.version || '1.0.0',
      author: def.author || 'Unknown',
      category: def.category || 'custom',
      grade: def.grade || 'basic',
      type: def.type || 'custom',
      icon: def.icon || 'default',
      tags: Array.isArray(def.tags) ? def.tags : [],
      dependencies: Array.isArray(def.dependencies) ? def.dependencies : [],
      schema: def.schema || {},
      documentation: def.documentation || {},
      location: nodePath,
      enabled: def.enabled !== false, // Default to true if not specified
      builtin: def.builtin || false,
      createdAt: def.createdAt || new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
  }

  /**
   * Saves the manifest to file
   */
  async save() {
    const manifestPath = path.join(this.basePath, 'manifest.json');
    
    const manifestToSave = {
      ...this.manifest,
      updatedAt: new Date().toISOString()
    };
    
    await fs.writeFile(manifestPath, JSON.stringify(manifestToSave, null, 2));
    
    return manifestPath;
  }

  /**
   * Gets all nodes
   */
  getAllNodes() {
    return Object.values(this.manifest.nodes);
  }

  /**
   * Gets a node by ID
   */
  getNodeById(nodeId) {
    return this.manifest.nodes[nodeId] || null;
  }

  /**
   * Gets nodes by category
   */
  getNodesByCategory(category) {
    const nodeIds = this.manifest.categories[category] || [];
    return nodeIds.map(id => this.manifest.nodes[id]).filter(Boolean);
  }

  /**
   * Gets nodes by tag
   */
  getNodesByTag(tag) {
    const nodeIds = this.manifest.tags[tag] || [];
    return nodeIds.map(id => this.manifest.nodes[id]).filter(Boolean);
  }

  /**
   * Searches nodes by name, description, or tags
   */
  searchNodes(query) {
    const searchTerm = query.toLowerCase();
    
    return Object.values(this.manifest.nodes).filter(node => 
      node.name.toLowerCase().includes(searchTerm) ||
      node.description.toLowerCase().includes(searchTerm) ||
      (node.tags && node.tags.some(tag => tag.toLowerCase().includes(searchTerm)))
    );
  }

  /**
   * Gets all available categories
   */
  getCategories() {
    return Object.keys(this.manifest.categories);
  }

  /**
   * Gets all available tags
   */
  getTags() {
    return Object.keys(this.manifest.tags);
  }

  /**
   * Checks if a node exists
   */
  nodeExists(nodeId) {
    return nodeId in this.manifest.nodes;
  }

  /**
   * Updates a node's metadata
   */
  updateNode(nodeId, updates) {
    if (!this.nodeExists(nodeId)) {
      return false;
    }

    const node = this.manifest.nodes[nodeId];
    const oldCategory = node.category;
    const oldTags = [...(node.tags || [])];

    // Update node
    this.manifest.nodes[nodeId] = {
      ...node,
      ...updates,
      updatedAt: new Date().toISOString()
    };

    // If category changed, update category index
    if (oldCategory !== this.manifest.nodes[nodeId].category) {
      // Remove from old category
      const oldCategoryIndex = this.manifest.categories[oldCategory]?.indexOf(nodeId);
      if (oldCategoryIndex !== -1) {
        this.manifest.categories[oldCategory].splice(oldCategoryIndex, 1);
      }
      
      // Add to new category
      const newCategory = this.manifest.nodes[nodeId].category;
      if (!this.manifest.categories[newCategory]) {
        this.manifest.categories[newCategory] = [];
      }
      if (!this.manifest.categories[newCategory].includes(nodeId)) {
        this.manifest.categories[newCategory].push(nodeId);
      }
    }

    // If tags changed, update tag indices
    const newTags = this.manifest.nodes[nodeId].tags || [];
    const removedTags = oldTags.filter(tag => !newTags.includes(tag));
    const addedTags = newTags.filter(tag => !oldTags.includes(tag));

    // Remove from old tags
    for (const tag of removedTags) {
      const tagIndex = this.manifest.tags[tag]?.indexOf(nodeId);
      if (tagIndex !== -1) {
        this.manifest.tags[tag].splice(tagIndex, 1);
      }
    }

    // Add to new tags
    for (const tag of addedTags) {
      if (!this.manifest.tags[tag]) {
        this.manifest.tags[tag] = [];
      }
      if (!this.manifest.tags[tag].includes(nodeId)) {
        this.manifest.tags[tag].push(nodeId);
      }
    }

    this.manifest.updatedAt = new Date().toISOString();
    return true;
  }

  /**
   * Removes a node from the manifest
   */
  removeNode(nodeId) {
    if (!this.nodeExists(nodeId)) {
      return false;
    }

    const node = this.manifest.nodes[nodeId];

    // Remove from category index
    const categoryIndex = this.manifest.categories[node.category]?.indexOf(nodeId);
    if (categoryIndex !== -1) {
      this.manifest.categories[node.category].splice(categoryIndex, 1);
    }

    // Remove from tag indices
    if (node.tags) {
      for (const tag of node.tags) {
        const tagIndex = this.manifest.tags[tag]?.indexOf(nodeId);
        if (tagIndex !== -1) {
          this.manifest.tags[tag].splice(tagIndex, 1);
        }
      }
    }

    // Remove from nodes
    delete this.manifest.nodes[nodeId];

    this.manifest.updatedAt = new Date().toISOString();
    return true;
  }

  /**
   * Gets statistics about the node ecosystem
   */
  getStats() {
    const nodes = this.getAllNodes();
    
    return {
      totalNodes: nodes.length,
      categories: Object.keys(this.manifest.categories).length,
      tags: Object.keys(this.manifest.tags).length,
      builtinNodes: nodes.filter(n => n.builtin).length,
      customNodes: nodes.filter(n => !n.builtin).length,
      enabledNodes: nodes.filter(n => n.enabled).length,
      disabledNodes: nodes.filter(n => !n.enabled).length,
      gradeDistribution: {
        basic: nodes.filter(n => n.grade === 'basic').length,
        intermediate: nodes.filter(n => n.grade === 'intermediate').length,
        advanced: nodes.filter(n => n.grade === 'advanced').length,
        elite: nodes.filter(n => n.grade === 'elite').length
      },
      categoriesBreakdown: Object.entries(this.manifest.categories).map(([category, nodeIds]) => ({
        category,
        count: nodeIds.length
      })),
      updatedAt: this.manifest.updatedAt
    };
  }

  /**
   * Validates the manifest integrity
   */
  validateIntegrity() {
    const errors = [];
    
    // Check that all nodes in categories actually exist
    for (const [category, nodeIds] of Object.entries(this.manifest.categories)) {
      for (const nodeId of nodeIds) {
        if (!this.nodeExists(nodeId)) {
          errors.push(\`Category '\${category}' references non-existent node: \${nodeId}\`);
        }
      }
    }

    // Check that all nodes in tags actually exist
    for (const [tag, nodeIds] of Object.entries(this.manifest.tags)) {
      for (const nodeId of nodeIds) {
        if (!this.nodeExists(nodeId)) {
          errors.push(\`Tag '\${tag}' references non-existent node: \${nodeId}\`);
        }
      }
    }

    // Check that all defined nodes have valid paths if they exist
    for (const [nodeId, nodeDef] of Object.entries(this.manifest.nodes)) {
      if (nodeDef.location) {
        if (!fs.access(nodeDef.location).then(() => true).catch(() => false)) {
          errors.push(\`Node '\${nodeId}' points to non-existent location: \${nodeDef.location}\`);
        }
      }
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }
}

module.exports = { NodeManifest };