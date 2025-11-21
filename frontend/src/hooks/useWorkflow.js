// frontend/src/hooks/useWorkflow.js
import { useState, useEffect, useCallback } from 'react';
import workflowService from '../services/workflowService';

export const useWorkflow = () => {
  const [workflows, setWorkflows] = useState([]);
  const [currentWorkflow, setCurrentWorkflow] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [executions, setExecutions] = useState([]);

  // Fetch all workflows
  const fetchWorkflows = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await workflowService.getWorkflows();
      setWorkflows(data.workflows || data.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch specific workflow
  const fetchWorkflow = useCallback(async (id) => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await workflowService.getWorkflow(id);
      setCurrentWorkflow(data.workflow || data.data || data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  // Create new workflow
  const createWorkflow = useCallback(async (workflowData) => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await workflowService.createWorkflow(workflowData);
      setWorkflows(prev => [...prev, data.workflow || data]);
      return data;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Update workflow
  const updateWorkflow = useCallback(async (id, workflowData) => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await workflowService.updateWorkflow(id, workflowData);
      setWorkflows(prev => prev.map(wf => wf.id === id ? (data.workflow || data) : wf));
      setCurrentWorkflow(data.workflow || data);
      return data;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Delete workflow
  const deleteWorkflow = useCallback(async (id) => {
    setLoading(true);
    setError(null);
    
    try {
      await workflowService.deleteWorkflow(id);
      setWorkflows(prev => prev.filter(wf => wf.id !== id));
      if (currentWorkflow?.id === id) {
        setCurrentWorkflow(null);
      }
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [currentWorkflow]);

  // Execute workflow
  const executeWorkflow = useCallback(async (id, executionData = {}) => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await workflowService.executeWorkflow(id, executionData);
      return data;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch workflow executions
  const fetchExecutions = useCallback(async (workflowId) => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await workflowService.getWorkflowExecutions(workflowId);
      setExecutions(data.executions || data.data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  // Validate workflow
  const validateWorkflow = useCallback(async (workflowData) => {
    try {
      return await workflowService.validateWorkflow(workflowData);
    } catch (err) {
      setError(err.message);
      throw err;
    }
  }, []);

  // Test node configuration
  const testNode = useCallback(async (nodeData) => {
    try {
      return await workflowService.testNode(nodeData);
    } catch (err) {
      setError(err.message);
      throw err;
    }
  }, []);

  // Clear error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Reset current workflow
  const resetCurrentWorkflow = useCallback(() => {
    setCurrentWorkflow(null);
  }, []);

  useEffect(() => {
    fetchWorkflows();
  }, [fetchWorkflows]);

  return {
    workflows,
    currentWorkflow,
    executions,
    loading,
    error,
    fetchWorkflows,
    fetchWorkflow,
    createWorkflow,
    updateWorkflow,
    deleteWorkflow,
    executeWorkflow,
    fetchExecutions,
    validateWorkflow,
    testNode,
    clearError,
    resetCurrentWorkflow,
  };
};