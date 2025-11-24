export interface BaseNode {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: {
    label: string;
    description?: string;
    inputs: NodePort[];
    outputs: NodePort[];
    config: Record<string, any>;
    status?: 'idle' | 'running' | 'success' | 'error';
  };
}

export interface NodePort {
  id: string;
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array' | 'file';
  required?: boolean;
  description?: string;
}

export interface Workflow {
  id: string;
  name: string;
  description?: string;
  nodes: BaseNode[];
  edges: Connection[];
  settings: {
    autoSave: boolean;
    errorHandling: 'stop' | 'continue' | 'retry';
    retryCount: number;
  };
  createdAt: Date;
  updatedAt: Date;
  version: number;
  isActive: boolean;
}

export interface Connection {
  id: string;
  source: string;
  target: string;
  sourceHandle: string;
  targetHandle: string;
  type?: string;
}

export interface Execution {
  id: string;
  workflowId: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  startedAt: Date;
  completedAt?: Date;
  duration?: number;
  nodeExecutions: NodeExecution[];
  logs: ExecutionLog[];
  inputData?: Record<string, any>;
  outputData?: Record<string, any>;
  error?: string;
}

export interface NodeExecution {
  id: string;
  nodeId: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  startedAt: Date;
  completedAt?: Date;
  duration?: number;
  inputData?: Record<string, any>;
  outputData?: Record<string, any>;
  error?: string;
  retryCount: number;
}

export interface ExecutionLog {
  id: string;
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  timestamp: Date;
  nodeId?: string;
  executionId: string;
  metadata?: Record<string, any>;
}

export interface NodeType {
  id: string;
  name: string;
  description: string;
  category: 'trigger' | 'action' | 'transform' | 'utility' | 'ai' | 'database' | 'communication';
  icon: string;
  inputs: NodePort[];
  outputs: NodePort[];
  config: NodeConfig[];
  version: string;
  documentation?: string;
}

export interface NodeConfig {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'select' | 'multiselect' | 'textarea' | 'password' | 'file' | 'json';
  label: string;
  description?: string;
  required?: boolean;
  default?: any;
  options?: { label: string; value: any }[];
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
  };
}

export interface WorkflowTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  tags: string[];
  nodes: BaseNode[];
  edges: Connection[];
  preview?: string;
  downloads: number;
  rating: number;
  author: string;
  createdAt: Date;
}