// Workflow types
export interface Workflow {
  id: string;
  name: string;
  description: string;
  nodes: Node[];
  edges: Edge[];
  createdAt: number;
  updatedAt: number;
}

export interface Node {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: NodeData;
}

export interface Edge {
  id: string;
  source: string;
  target: string;
  animated?: boolean;
}

export interface NodeData {
  id?: string;
  label: string;
  description: string;
  type: string;
  parameters: Record<string, any>;
}

// Execution types
export interface Execution {
  id: string;
  workflowId: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  startedAt: string;
  endedAt?: string;
  results: Record<string, any>;
  error?: string;
  variables: Record<string, any>;
}

export interface ExecutionResult {
  nodeId: string;
  status: 'success' | 'error' | 'running';
  data: any;
  error?: string;
  timestamp: string;
}

// API response types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

// Plugin types
export interface Plugin {
  id: string;
  name: string;
  description: string;
  type: 'javascript' | 'python' | 'builtin';
  schema: any;
}

// User types
export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}