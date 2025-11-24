import { useCallback, useEffect, useState } from 'react';
import { Execution, ExecutionLog, ExecutionFilter } from '@/types/execution';

export function useExecution(workflowId?: string) {
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [currentExecution, setCurrentExecution] = useState<Execution | null>(null);
  const [logs, setLogs] = useState<ExecutionLog[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load executions for a workflow
  const loadExecutions = useCallback(async (filter?: ExecutionFilter) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams();
      if (workflowId) params.append('workflowId', workflowId);
      if (filter?.status) params.append('status', filter.status);
      if (filter?.dateFrom) params.append('dateFrom', filter.dateFrom.toISOString());
      if (filter?.dateTo) params.append('dateTo', filter.dateTo.toISOString());
      if (filter?.search) params.append('search', filter.search);
      
      const response = await fetch(`/api/executions?${params.toString()}`);
      if (!response.ok) {
        throw new Error('Failed to load executions');
      }
      
      const data = await response.json();
      setExecutions(data.data);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load executions');
    } finally {
      setIsLoading(false);
    }
  }, [workflowId]);

  // Load specific execution
  const loadExecution = useCallback(async (executionId: string) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/executions/${executionId}`);
      if (!response.ok) {
        throw new Error('Failed to load execution');
      }
      
      const data = await response.json();
      setCurrentExecution(data.data);
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load execution');
      return null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Execute workflow
  const executeWorkflow = useCallback(async (workflowId: string, inputData?: Record<string, any>) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/executions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          workflowId,
          inputData
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to execute workflow');
      }
      
      const data = await response.json();
      const newExecution = data.data;
      
      setExecutions(prev => [newExecution, ...prev]);
      setCurrentExecution(newExecution);
      
      return newExecution;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to execute workflow');
      return null;
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Stop execution
  const stopExecution = useCallback(async (executionId: string) => {
    try {
      const response = await fetch(`/api/executions/${executionId}/stop`, {
        method: 'POST',
      });
      
      if (!response.ok) {
        throw new Error('Failed to stop execution');
      }
      
      const data = await response.json();
      
      setExecutions(prev => prev.map(exec => 
        exec.id === executionId ? data.data : exec
      ));
      
      if (currentExecution?.id === executionId) {
        setCurrentExecution(data.data);
      }
      
      return true;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to stop execution');
      return false;
    }
  }, [currentExecution]);

  // Retry execution
  const retryExecution = useCallback(async (executionId: string) => {
    try {
      const response = await fetch(`/api/executions/${executionId}/retry`, {
        method: 'POST',
      });
      
      if (!response.ok) {
        throw new Error('Failed to retry execution');
      }
      
      const data = await response.json();
      const newExecution = data.data;
      
      setExecutions(prev => [newExecution, ...prev]);
      
      return newExecution;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to retry execution');
      return null;
    }
  }, []);

  // Load execution logs
  const loadExecutionLogs = useCallback(async (executionId: string) => {
    try {
      const response = await fetch(`/api/executions/${executionId}/logs`);
      if (!response.ok) {
        throw new Error('Failed to load execution logs');
      }
      
      const data = await response.json();
      setLogs(data.data);
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load execution logs');
      return null;
    }
  }, []);

  // Get execution statistics
  const getExecutionStats = useCallback(async (workflowId?: string) => {
    try {
      const params = workflowId ? `?workflowId=${workflowId}` : '';
      const response = await fetch(`/api/executions/stats${params}`);
      
      if (!response.ok) {
        throw new Error('Failed to load execution statistics');
      }
      
      const data = await response.json();
      return data.data;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load execution statistics');
      return null;
    }
  }, []);

  // WebSocket for real-time updates
  const setupWebSocket = useCallback((executionId: string) => {
    const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WS_URL}/executions/${executionId}`);
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      switch (data.type) {
        case 'execution_update':
          setCurrentExecution(prev => prev ? { ...prev, ...data.payload } : null);
          break;
          
        case 'log':
          setLogs(prev => [data.payload, ...prev]);
          break;
          
        case 'execution_complete':
          setCurrentExecution(prev => prev ? { ...prev, ...data.payload } : null);
          ws.close();
          break;
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setError('WebSocket connection failed');
    };
    
    return ws;
  }, []);

  // Load executions on mount
  useEffect(() => {
    loadExecutions();
  }, [loadExecutions]);

  // Set up WebSocket when current execution changes
  useEffect(() => {
    if (currentExecution && currentExecution.status === 'running') {
      const ws = setupWebSocket(currentExecution.id);
      
      return () => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.close();
        }
      };
    }
  }, [currentExecution, setupWebSocket]);

  return {
    // State
    executions,
    currentExecution,
    logs,
    isLoading,
    error,
    
    // Actions
    loadExecutions,
    loadExecution,
    executeWorkflow,
    stopExecution,
    retryExecution,
    loadExecutionLogs,
    getExecutionStats,
    
    // Utilities
    setCurrentExecution,
    clearError: () => setError(null)
  };
}