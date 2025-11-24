import { useCallback, useEffect } from 'react';
import { useWorkflowStore } from '@/stores/workflowStore';
import { useNodeStore } from '@/stores/nodeStore';
import { BaseNode, Connection, Workflow } from '@/types/workflow';

export function useWorkflow(workflowId?: string) {
  const {
    currentWorkflow,
    workflows,
    setCurrentWorkflow,
    addWorkflow,
    updateWorkflow,
    deleteWorkflow,
    addNode,
    updateNode,
    deleteNode,
    duplicateNode,
    addEdge,
    updateEdge,
    deleteEdge,
    selectedNodes,
    selectedEdges,
    selectNodes,
    selectEdges,
    clearSelection,
    setLoading,
    setError
  } = useWorkflowStore();

  const { nodeTypes } = useNodeStore();

  // Load workflow by ID
  const loadWorkflow = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/workflows/${id}`);
      if (!response.ok) {
        throw new Error('Failed to load workflow');
      }
      
      const data = await response.json();
      setCurrentWorkflow(data.data);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load workflow');
    } finally {
      setLoading(false);
    }
  }, [setCurrentWorkflow, setLoading, setError]);

  // Save current workflow
  const saveWorkflow = useCallback(async () => {
    if (!currentWorkflow) return false;
    
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/workflows/${currentWorkflow.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(currentWorkflow),
      });
      
      if (!response.ok) {
        throw new Error('Failed to save workflow');
      }
      
      const data = await response.json();
      updateWorkflow(currentWorkflow.id, data.data);
      return true;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to save workflow');
      return false;
    } finally {
      setLoading(false);
    }
  }, [currentWorkflow, updateWorkflow, setLoading, setError]);

  // Create new workflow
  const createWorkflow = useCallback(async (workflowData: Partial<Workflow>) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/workflows', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(workflowData),
      });
      
      if (!response.ok) {
        throw new Error('Failed to create workflow');
      }
      
      const data = await response.json();
      addWorkflow(data.data);
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to create workflow');
      return null;
    } finally {
      setLoading(false);
    }
  }, [addWorkflow, setLoading, setError]);

  // Execute workflow
  const executeWorkflow = useCallback(async () => {
    if (!currentWorkflow) return null;
    
    try {
      const response = await fetch('/api/workflows/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          workflowId: currentWorkflow.id,
          nodes: currentWorkflow.nodes,
          edges: currentWorkflow.edges,
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to execute workflow');
      }
      
      const data = await response.json();
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to execute workflow');
      return null;
    }
  }, [currentWorkflow, setError]);

  // Validate workflow
  const validateWorkflow = useCallback(() => {
    if (!currentWorkflow) return { isValid: false, errors: [] };
    
    const errors: string[] = [];
    
    // Check if workflow has at least one node
    if (currentWorkflow.nodes.length === 0) {
      errors.push('Workflow must have at least one node');
    }
    
    // Check for orphaned nodes (nodes without connections)
    const connectedNodes = new Set();
    currentWorkflow.edges.forEach(edge => {
      connectedNodes.add(edge.source);
      connectedNodes.add(edge.target);
    });
    
    const orphanedNodes = currentWorkflow.nodes.filter(
      node => !connectedNodes.has(node.id) && currentWorkflow.nodes.length > 1
    );
    
    if (orphanedNodes.length > 0) {
      errors.push(`${orphanedNodes.length} node(s) are not connected`);
    }
    
    // Check for required node configurations
    currentWorkflow.nodes.forEach(node => {
      const nodeType = nodeTypes.find(nt => nt.id === node.type);
      if (!nodeType) return;
      
      nodeType.config.forEach(config => {
        if (config.required && !node.data.config[config.name]) {
          errors.push(`Node "${node.data.label}" missing required configuration: ${config.label}`);
        }
      });
    });
    
    return {
      isValid: errors.length === 0,
      errors
    };
  }, [currentWorkflow, nodeTypes]);

  // Auto-save effect
  useEffect(() => {
    if (!currentWorkflow) return;
    
    const autoSaveInterval = setInterval(() => {
      if (currentWorkflow.settings.autoSave) {
        saveWorkflow();
      }
    }, 30000); // Auto-save every 30 seconds
    
    return () => clearInterval(autoSaveInterval);
  }, [currentWorkflow, saveWorkflow]);

  // Load workflow on mount if ID is provided
  useEffect(() => {
    if (workflowId) {
      loadWorkflow(workflowId);
    }
  }, [workflowId, loadWorkflow]);

  return {
    // Current state
    workflow: currentWorkflow,
    workflows,
    selectedNodes,
    selectedEdges,
    
    // Actions
    loadWorkflow,
    saveWorkflow,
    createWorkflow,
    executeWorkflow,
    validateWorkflow,
    
    // Node actions
    addNode,
    updateNode,
    deleteNode,
    duplicateNode,
    
    // Edge actions
    addEdge,
    updateEdge,
    deleteEdge,
    
    // Selection actions
    selectNodes,
    selectEdges,
    clearSelection,
    
    // Utility
    setCurrentWorkflow
  };
}