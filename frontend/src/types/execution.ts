export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';
export type NodeStatus = 'idle' | 'running' | 'success' | 'error';
export type WorkflowStatus = 'draft' | 'active' | 'inactive' | 'archived';

export interface ExecutionFilter {
  status?: ExecutionStatus;
  workflowId?: string;
  dateFrom?: Date;
  dateTo?: Date;
  search?: string;
}

export interface WorkflowFilter {
  status?: WorkflowStatus;
  category?: string;
  search?: string;
  sortBy?: 'name' | 'createdAt' | 'updatedAt' | 'executions';
  sortOrder?: 'asc' | 'desc';
}

export interface ExecutionMetrics {
  totalExecutions: number;
  successRate: number;
  averageDuration: number;
  errorRate: number;
  executionsByStatus: Record<ExecutionStatus, number>;
  executionsByWorkflow: Record<string, number>;
  executionsOverTime: Array<{
    date: string;
    count: number;
    successRate: number;
  }>;
}

export interface NodeMetrics {
  nodeId: string;
  nodeName: string;
  totalExecutions: number;
  successRate: number;
  averageDuration: number;
  errorCount: number;
  lastExecution?: Date;
}