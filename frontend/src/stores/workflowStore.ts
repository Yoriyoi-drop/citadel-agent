import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import { Workflow, BaseNode, Connection } from '@/types/workflow';

interface WorkflowState {
  // Current workflow
  currentWorkflow: Workflow | null;
  selectedNodes: string[];
  selectedEdges: string[];
  
  // Workflow list
  workflows: Workflow[];
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setCurrentWorkflow: (workflow: Workflow | null) => void;
  setWorkflows: (workflows: Workflow[]) => void;
  addWorkflow: (workflow: Workflow) => void;
  updateWorkflow: (id: string, updates: Partial<Workflow>) => void;
  deleteWorkflow: (id: string) => void;
  
  // Node actions
  addNode: (node: BaseNode) => void;
  updateNode: (id: string, updates: Partial<BaseNode>) => void;
  deleteNode: (id: string) => void;
  duplicateNode: (id: string) => void;
  
  // Edge actions
  addEdge: (edge: Connection) => void;
  updateEdge: (id: string, updates: Partial<Connection>) => void;
  deleteEdge: (id: string) => void;
  
  // Selection actions
  selectNodes: (nodeIds: string[]) => void;
  selectEdges: (edgeIds: string[]) => void;
  clearSelection: () => void;
  
  // Utility actions
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useWorkflowStore = create<WorkflowState>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    currentWorkflow: null,
    selectedNodes: [],
    selectedEdges: [],
    workflows: [],
    isLoading: false,
    error: null,
    
    // Workflow actions
    setCurrentWorkflow: (workflow) => set({ currentWorkflow: workflow }),
    
    setWorkflows: (workflows) => set({ workflows }),
    
    addWorkflow: (workflow) => set((state) => ({
      workflows: [...state.workflows, workflow]
    })),
    
    updateWorkflow: (id, updates) => set((state) => ({
      workflows: state.workflows.map(w => 
        w.id === id ? { ...w, ...updates, updatedAt: new Date() } : w
      ),
      currentWorkflow: state.currentWorkflow?.id === id 
        ? { ...state.currentWorkflow, ...updates, updatedAt: new Date() }
        : state.currentWorkflow
    })),
    
    deleteWorkflow: (id) => set((state) => ({
      workflows: state.workflows.filter(w => w.id !== id),
      currentWorkflow: state.currentWorkflow?.id === id ? null : state.currentWorkflow
    })),
    
    // Node actions
    addNode: (node) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        nodes: [...state.currentWorkflow.nodes, node],
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    updateNode: (id, updates) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        nodes: state.currentWorkflow.nodes.map(n => 
          n.id === id ? { ...n, ...updates } : n
        ),
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    deleteNode: (id) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        nodes: state.currentWorkflow.nodes.filter(n => n.id !== id),
        edges: state.currentWorkflow.edges.filter(e => 
          e.source !== id && e.target !== id
        ),
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        selectedNodes: state.selectedNodes.filter(nId => nId !== id),
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    duplicateNode: (id) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const nodeToDuplicate = state.currentWorkflow.nodes.find(n => n.id === id);
      if (!nodeToDuplicate) return state;
      
      const duplicatedNode: BaseNode = {
        ...nodeToDuplicate,
        id: `${nodeToDuplicate.id}_copy_${Date.now()}`,
        position: {
          x: nodeToDuplicate.position.x + 50,
          y: nodeToDuplicate.position.y + 50
        },
        data: {
          ...nodeToDuplicate.data,
          label: `${nodeToDuplicate.data.label} (Copy)`
        }
      };
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        nodes: [...state.currentWorkflow.nodes, duplicatedNode],
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    // Edge actions
    addEdge: (edge) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        edges: [...state.currentWorkflow.edges, edge],
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    updateEdge: (id, updates) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        edges: state.currentWorkflow.edges.map(e => 
          e.id === id ? { ...e, ...updates } : e
        ),
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    deleteEdge: (id) => set((state) => {
      if (!state.currentWorkflow) return state;
      
      const updatedWorkflow = {
        ...state.currentWorkflow,
        edges: state.currentWorkflow.edges.filter(e => e.id !== id),
        updatedAt: new Date()
      };
      
      return {
        currentWorkflow: updatedWorkflow,
        selectedEdges: state.selectedEdges.filter(eId => eId !== id),
        workflows: state.workflows.map(w => 
          w.id === updatedWorkflow.id ? updatedWorkflow : w
        )
      };
    }),
    
    // Selection actions
    selectNodes: (nodeIds) => set({ selectedNodes: nodeIds }),
    
    selectEdges: (edgeIds) => set({ selectedEdges: edgeIds }),
    
    clearSelection: () => set({ selectedNodes: [], selectedEdges: [] }),
    
    // Utility actions
    setLoading: (loading) => set({ isLoading: loading }),
    
    setError: (error) => set({ error })
  }))
);