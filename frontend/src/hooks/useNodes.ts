import { useCallback, useEffect, useState } from 'react';
import { useNodeStore } from '@/stores/nodeStore';
import { NodeType, NodeConfig } from '@/types/workflow';

export function useNodes() {
  const {
    nodeTypes,
    setNodeTypes,
    addNodeType,
    updateNodeType,
    deleteNodeType,
    getNodeTypesByCategory,
    searchNodeTypes,
    setLoading,
    setError
  } = useNodeStore();

  const [filteredNodes, setFilteredNodes] = useState<NodeType[]>(nodeTypes);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');

  // Load node types from API
  const loadNodeTypes = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/nodes');
      if (!response.ok) {
        throw new Error('Failed to load node types');
      }
      
      const data = await response.json();
      setNodeTypes(data.data);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load node types');
    } finally {
      setLoading(false);
    }
  }, [setNodeTypes, setLoading, setError]);

  // Create custom node type
  const createNodeType = useCallback(async (nodeTypeData: Partial<NodeType>) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/nodes', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(nodeTypeData),
      });
      
      if (!response.ok) {
        throw new Error('Failed to create node type');
      }
      
      const data = await response.json();
      addNodeType(data.data);
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to create node type');
      return null;
    } finally {
      setLoading(false);
    }
  }, [addNodeType, setLoading, setError]);

  // Update node type
  const updateNodeTypeConfig = useCallback(async (id: string, updates: Partial<NodeType>) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/nodes/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      });
      
      if (!response.ok) {
        throw new Error('Failed to update node type');
      }
      
      const data = await response.json();
      updateNodeType(id, data.data);
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to update node type');
      return null;
    } finally {
      setLoading(false);
    }
  }, [updateNodeType, setLoading, setError]);

  // Delete node type
  const removeNodeType = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/nodes/${id}`, {
        method: 'DELETE',
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete node type');
      }
      
      deleteNodeType(id);
      return true;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to delete node type');
      return false;
    } finally {
      setLoading(false);
    }
  }, [deleteNodeType, setLoading, setError]);

  // Validate node configuration
  const validateNodeConfig = useCallback((nodeType: NodeType, config: Record<string, any>) => {
    const errors: string[] = [];
    const warnings: string[] = [];
    
    nodeType.config.forEach((configField: NodeConfig) => {
      const value = config[configField.name];
      
      // Check required fields
      if (configField.required && (value === undefined || value === null || value === '')) {
        errors.push(`"${configField.label}" is required`);
      }
      
      // Type validation
      if (value !== undefined && value !== null) {
        switch (configField.type) {
          case 'number':
            if (isNaN(Number(value))) {
              errors.push(`"${configField.label}" must be a number`);
            } else {
              const numValue = Number(value);
              if (configField.validation?.min !== undefined && numValue < configField.validation.min) {
                errors.push(`"${configField.label}" must be at least ${configField.validation.min}`);
              }
              if (configField.validation?.max !== undefined && numValue > configField.validation.max) {
                errors.push(`"${configField.label}" must be at most ${configField.validation.max}`);
              }
            }
            break;
            
          case 'string':
            if (configField.validation?.pattern && !new RegExp(configField.validation.pattern).test(value)) {
              errors.push(`"${configField.label}" format is invalid`);
            }
            break;
            
          case 'select':
            if (configField.options && !configField.options.some(option => option.value === value)) {
              errors.push(`"${configField.label}" must be one of the provided options`);
            }
            break;
        }
      }
    });
    
    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }, []);

  // Get node type by ID
  const getNodeTypeById = useCallback((id: string) => {
    return nodeTypes.find(nt => nt.id === id);
  }, [nodeTypes]);

  // Get node types by category
  const getNodeTypesByCategoryFilter = useCallback((category: string) => {
    return getNodeTypesByCategory(category);
  }, [getNodeTypesByCategory]);

  // Search node types
  const searchNodes = useCallback((query: string) => {
    return searchNodeTypes(query);
  }, [searchNodeTypes]);

  // Filter nodes based on category and search
  useEffect(() => {
    let filtered = nodeTypes;
    
    if (selectedCategory !== 'all') {
      filtered = getNodeTypesByCategory(selectedCategory);
    }
    
    if (searchQuery) {
      filtered = filtered.filter(node => 
        node.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        node.description.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }
    
    setFilteredNodes(filtered);
  }, [nodeTypes, selectedCategory, searchQuery, getNodeTypesByCategory]);

  // Load node types on mount
  useEffect(() => {
    loadNodeTypes();
  }, [loadNodeTypes]);

  return {
    // State
    nodeTypes,
    filteredNodes,
    selectedCategory,
    searchQuery,
    
    // Actions
    loadNodeTypes,
    createNodeType,
    updateNodeTypeConfig,
    removeNodeType,
    
    // Filters
    setSelectedCategory,
    setSearchQuery,
    
    // Utilities
    validateNodeConfig,
    getNodeTypeById,
    getNodeTypesByCategory: getNodeTypesByCategoryFilter,
    searchNodes
  };
}