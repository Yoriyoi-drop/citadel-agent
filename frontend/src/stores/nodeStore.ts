import { create } from 'zustand';
import { NodeType } from '@/types/workflow';

interface NodeState {
  // Node types registry
  nodeTypes: NodeType[];
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setNodeTypes: (nodeTypes: NodeType[]) => void;
  addNodeType: (nodeType: NodeType) => void;
  updateNodeType: (id: string, updates: Partial<NodeType>) => void;
  deleteNodeType: (id: string) => void;
  getNodeTypesByCategory: (category: string) => NodeType[];
  searchNodeTypes: (query: string) => NodeType[];
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useNodeStore = create<NodeState>((set, get) => ({
  // Initial state
  nodeTypes: [
    // HTTP Nodes
    {
      id: 'http_request',
      name: 'HTTP Request',
      description: 'Make HTTP requests to any API',
      category: 'action',
      icon: 'Globe',
      inputs: [
        { id: 'url', name: 'URL', type: 'string', required: true },
        { id: 'method', name: 'Method', type: 'string', required: true },
        { id: 'headers', name: 'Headers', type: 'object' },
        { id: 'body', name: 'Body', type: 'object' }
      ],
      outputs: [
        { id: 'response', name: 'Response', type: 'object' },
        { id: 'status', name: 'Status Code', type: 'number' },
        { id: 'headers', name: 'Response Headers', type: 'object' }
      ],
      config: [
        {
          name: 'url',
          type: 'string',
          label: 'URL',
          required: true,
          description: 'The URL to make the request to'
        },
        {
          name: 'method',
          type: 'select',
          label: 'HTTP Method',
          required: true,
          default: 'GET',
          options: [
            { label: 'GET', value: 'GET' },
            { label: 'POST', value: 'POST' },
            { label: 'PUT', value: 'PUT' },
            { label: 'DELETE', value: 'DELETE' },
            { label: 'PATCH', value: 'PATCH' }
          ]
        },
        {
          name: 'headers',
          type: 'json',
          label: 'Headers',
          description: 'HTTP headers to send with the request'
        },
        {
          name: 'body',
          type: 'json',
          label: 'Body',
          description: 'Request body data'
        }
      ],
      version: '1.0.0'
    },
    
    // Database Nodes
    {
      id: 'database_query',
      name: 'Database Query',
      description: 'Execute SQL queries on database',
      category: 'database',
      icon: 'Database',
      inputs: [
        { id: 'query', name: 'SQL Query', type: 'string', required: true },
        { id: 'params', name: 'Parameters', type: 'array' }
      ],
      outputs: [
        { id: 'results', name: 'Results', type: 'array' },
        { id: 'affectedRows', name: 'Affected Rows', type: 'number' }
      ],
      config: [
        {
          name: 'connection',
          type: 'select',
          label: 'Database Connection',
          required: true,
          description: 'Select database connection'
        },
        {
          name: 'query',
          type: 'textarea',
          label: 'SQL Query',
          required: true,
          description: 'SQL query to execute'
        },
        {
          name: 'params',
          type: 'json',
          label: 'Query Parameters',
          description: 'Parameters for the SQL query'
        }
      ],
      version: '1.0.0'
    },
    
    // AI Nodes
    {
      id: 'ai_chat',
      name: 'AI Chat',
      description: 'Generate text using AI models',
      category: 'ai',
      icon: 'Brain',
      inputs: [
        { id: 'prompt', name: 'Prompt', type: 'string', required: true },
        { id: 'context', name: 'Context', type: 'string' },
        { id: 'messages', name: 'Messages', type: 'array' }
      ],
      outputs: [
        { id: 'response', name: 'Response', type: 'string' },
        { id: 'usage', name: 'Token Usage', type: 'object' }
      ],
      config: [
        {
          name: 'model',
          type: 'select',
          label: 'AI Model',
          required: true,
          default: 'gpt-3.5-turbo',
          options: [
            { label: 'GPT-3.5 Turbo', value: 'gpt-3.5-turbo' },
            { label: 'GPT-4', value: 'gpt-4' },
            { label: 'Claude-3', value: 'claude-3' },
            { label: 'Gemini Pro', value: 'gemini-pro' }
          ]
        },
        {
          name: 'temperature',
          type: 'number',
          label: 'Temperature',
          default: 0.7,
          validation: { min: 0, max: 2 }
        },
        {
          name: 'maxTokens',
          type: 'number',
          label: 'Max Tokens',
          default: 1000,
          validation: { min: 1, max: 4000 }
        },
        {
          name: 'systemPrompt',
          type: 'textarea',
          label: 'System Prompt',
          description: 'System instructions for the AI model'
        }
      ],
      version: '1.0.0'
    },
    
    // Trigger Nodes
    {
      id: 'webhook',
      name: 'Webhook',
      description: 'Trigger workflow via HTTP webhook',
      category: 'trigger',
      icon: 'Globe',
      inputs: [],
      outputs: [
        { id: 'body', name: 'Request Body', type: 'object' },
        { id: 'headers', name: 'Headers', type: 'object' },
        { id: 'query', name: 'Query Parameters', type: 'object' }
      ],
      config: [
        {
          name: 'path',
          type: 'string',
          label: 'Webhook Path',
          required: true,
          description: 'URL path for the webhook'
        },
        {
          name: 'method',
          type: 'select',
          label: 'HTTP Method',
          default: 'POST',
          options: [
            { label: 'GET', value: 'GET' },
            { label: 'POST', value: 'POST' },
            { label: 'PUT', value: 'PUT' }
          ]
        }
      ],
      version: '1.0.0'
    },
    
    // Transform Nodes
    {
      id: 'transform_data',
      name: 'Transform Data',
      description: 'Transform and manipulate data',
      category: 'transform',
      icon: 'Settings',
      inputs: [
        { id: 'data', name: 'Input Data', type: 'object', required: true }
      ],
      outputs: [
        { id: 'output', name: 'Transformed Data', type: 'object' }
      ],
      config: [
        {
          name: 'transformation',
          type: 'json',
          label: 'Transformation Rules',
          required: true,
          description: 'JSON transformation rules'
        }
      ],
      version: '1.0.0'
    },
    
    // Utility Nodes
    {
      id: 'delay',
      name: 'Delay',
      description: 'Add delay between nodes',
      category: 'utility',
      icon: 'Clock',
      inputs: [
        { id: 'input', name: 'Input', type: 'any' }
      ],
      outputs: [
        { id: 'output', name: 'Output', type: 'any' }
      ],
      config: [
        {
          name: 'duration',
          type: 'number',
          label: 'Delay (seconds)',
          required: true,
          default: 1,
          validation: { min: 0 }
        }
      ],
      version: '1.0.0'
    }
  ],
  isLoading: false,
  error: null,
  
  // Actions
  setNodeTypes: (nodeTypes) => set({ nodeTypes }),
  
  addNodeType: (nodeType) => set((state) => ({
    nodeTypes: [...state.nodeTypes, nodeType]
  })),
  
  updateNodeType: (id, updates) => set((state) => ({
    nodeTypes: state.nodeTypes.map(nt => 
      nt.id === id ? { ...nt, ...updates } : nt
    )
  })),
  
  deleteNodeType: (id) => set((state) => ({
    nodeTypes: state.nodeTypes.filter(nt => nt.id !== id)
  })),
  
  getNodeTypesByCategory: (category) => {
    const { nodeTypes } = get();
    return nodeTypes.filter(nt => nt.category === category);
  },
  
  searchNodeTypes: (query) => {
    const { nodeTypes } = get();
    const lowercaseQuery = query.toLowerCase();
    return nodeTypes.filter(nt => 
      nt.name.toLowerCase().includes(lowercaseQuery) ||
      nt.description.toLowerCase().includes(lowercaseQuery)
    );
  },
  
  setLoading: (loading) => set({ isLoading: loading }),
  
  setError: (error) => set({ error })
}));