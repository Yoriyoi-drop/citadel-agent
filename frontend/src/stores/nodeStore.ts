import { create } from 'zustand';
import { NodeType } from '@/types/workflow';

interface NodeState {
  // Node types registry
  nodeTypes: NodeType[];
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchNodes: () => Promise<void>;
  setNodeTypes: (nodeTypes: NodeType[]) => void;
  addNodeType: (nodeType: NodeType) => void;
  updateNodeType: (id: string, updates: Partial<NodeType>) => void;
  deleteNodeType: (id: string) => void;
  getNodeTypesByCategory: (category: string) => NodeType[];
  searchNodeTypes: (query: string) => NodeType[];
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const useNodeStore = create<NodeState>((set, get) => ({
  // Initial state
  nodeTypes: [],
  isLoading: false,
  error: null,

  // Fetch nodes from API
  fetchNodes: async () => {
    set({ isLoading: true, error: null });

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/registry/nodes`);

      if (!response.ok) {
        throw new Error('Failed to fetch nodes');
      }

      const data = await response.json();

      if (data.success && data.data?.nodes) {
        // Transform backend node format to frontend format
        const nodes: NodeType[] = data.data.nodes.map((node: any) => ({
          id: node.id,
          name: node.name,
          description: node.description,
          category: node.category,
          icon: node.icon,
          inputs: node.inputs || [],
          outputs: node.outputs || [],
          config: node.config || [],
          version: node.version || '1.0.0',
        }));

        set({ nodeTypes: nodes, isLoading: false });
      } else {
        throw new Error('Invalid response format');
      }
    } catch (error) {
      console.error('Error fetching nodes:', error);
      set({
        error: error instanceof Error ? error.message : 'Unknown error',
        isLoading: false
      });
    }
  },

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