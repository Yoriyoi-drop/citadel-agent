import { create } from 'zustand';
import { subscribeWithSelector, persist } from 'zustand/middleware';
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
  persist(
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
        // Ensure node has a unique ID. If not provided, generate one.
        const newNode = { ...node, id: node.id || `node_${Date.now()}_${Math.random().toString(36).substring(2, 9)}` };

        // If no current workflow, create a new one
        if (!state.currentWorkflow) {
          const newWorkflow = {
            id: `workflow_${Date.now()}`,
            name: 'New Workflow',
            description: 'Untitled workflow',
            nodes: [newNode],
            edges: [],
            settings: {
              autoSave: true,
              errorHandling: 'stop' as const,
              retryCount: 3
            },
            createdAt: new Date(),
            updatedAt: new Date(),
            version: 1,
            isActive: false
          };

          return {
            currentWorkflow: newWorkflow,
            workflows: [...state.workflows, newWorkflow],
            error: null // Clear any previous error
          };
        }

        // Check if a node with the same ID already exists
        if (state.currentWorkflow.nodes.some(n => n.id === newNode.id)) {
          console.warn(`Node with ID ${newNode.id} already exists. Not adding.`);
          return { ...state, error: `Node with ID ${newNode.id} already exists.` };
        }

        const updatedWorkflow = {
          ...state.currentWorkflow,
          nodes: [...state.currentWorkflow.nodes, newNode],
          updatedAt: new Date()
        };

        return {
          currentWorkflow: updatedWorkflow,
          workflows: state.workflows.map(w =>
            w.id === updatedWorkflow.id ? updatedWorkflow : w
          ),
          error: null // Clear any previous error
        };
      }),

      updateNode: (id, updates) => set((state) => {
        if (!state.currentWorkflow) {
          console.warn('Cannot update node: No current workflow selected.');
          return { ...state, error: 'Cannot update node: No current workflow selected.' };
        }

        const nodeExists = state.currentWorkflow.nodes.some(n => n.id === id);
        if (!nodeExists) {
          console.warn(`Node with ID ${id} not found for update.`);
          return { ...state, error: `Node with ID ${id} not found.` };
        }

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
          ),
          error: null // Clear any previous error
        };
      }),

      deleteNode: (id) => set((state) => {
        if (!state.currentWorkflow) {
          console.warn('Cannot delete node: No current workflow selected.');
          return { ...state, error: 'Cannot delete node: No current workflow selected.' };
        }

        const nodeExists = state.currentWorkflow.nodes.some(n => n.id === id);
        if (!nodeExists) {
          console.warn(`Node with ID ${id} not found for deletion.`);
          return { ...state, error: `Node with ID ${id} not found.` };
        }

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
          ),
          error: null // Clear any previous error
        };
      }),

      duplicateNode: (id) => set((state) => {
        if (!state.currentWorkflow) {
          console.warn('Cannot duplicate node: No current workflow selected.');
          return { ...state, error: 'Cannot duplicate node: No current workflow selected.' };
        }

        const nodeToDuplicate = state.currentWorkflow.nodes.find(n => n.id === id);
        if (!nodeToDuplicate) {
          console.warn(`Node with ID ${id} not found for duplication.`);
          return { ...state, error: `Node with ID ${id} not found for duplication.` };
        }

        // Generate a more robust unique ID for the duplicated node
        const duplicatedNode: BaseNode = {
          ...nodeToDuplicate,
          id: `${nodeToDuplicate.id}_copy_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
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
          ),
          error: null // Clear any previous error
        };
      }),

      // Edge actions
      addEdge: (edge) => set((state) => {
        if (!state.currentWorkflow) {
          console.warn('Cannot add edge: No current workflow selected.');
          return { ...state, error: 'Cannot add edge: No current workflow selected.' };
        }

        // Check if edge already exists (simple check based on ID, or source/target)
        if (state.currentWorkflow.edges.some(e => e.id === edge.id)) {
          console.warn(`Edge with ID ${edge.id} already exists. Not adding.`);
          return { ...state, error: `Edge with ID ${edge.id} already exists.` };
        }

        const updatedWorkflow = {
          ...state.currentWorkflow,
          edges: [...state.currentWorkflow.edges, edge],
          updatedAt: new Date()
        };

        return {
          currentWorkflow: updatedWorkflow,
          workflows: state.workflows.map(w =>
            w.id === updatedWorkflow.id ? updatedWorkflow : w
          ),
          error: null // Clear any previous error
        };
      }),

      updateEdge: (id, updates) => set((state) => {
        if (!state.currentWorkflow) {
          console.warn('Cannot update edge: No current workflow selected.');
          return { ...state, error: 'Cannot update edge: No current workflow selected.' };
        }

        const edgeExists = state.currentWorkflow.edges.some(e => e.id === id);
        if (!edgeExists) {
          console.warn(`Edge with ID ${id} not found for update.`);
          return { ...state, error: `Edge with ID ${id} not found.` };
        }

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
          ),
          error: null // Clear any previous error
        };
      }),

      deleteEdge: (id) => set((state) => {
        if (!state.currentWorkflow) {
          console.warn('Cannot delete edge: No current workflow selected.');
          return { ...state, error: 'Cannot delete edge: No current workflow selected.' };
        }

        const edgeExists = state.currentWorkflow.edges.some(e => e.id === id);
        if (!edgeExists) {
          console.warn(`Edge with ID ${id} not found for deletion.`);
          return { ...state, error: `Edge with ID ${id} not found.` };
        }

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
          ),
          error: null // Clear any previous error
        };
      }),

      // Selection actions
      selectNodes: (nodeIds) => set({ selectedNodes: nodeIds }),

      selectEdges: (edgeIds) => set({ selectedEdges: edgeIds }),

      clearSelection: () => set({ selectedNodes: [], selectedEdges: [] }),

      // Utility actions
      setLoading: (loading) => set({ isLoading: loading }),

      setError: (error) => set({ error })
    })),
    {
      name: 'citadel-workflow-storage', // localStorage key
      partialize: (state) => ({
        currentWorkflow: state.currentWorkflow,
        workflows: state.workflows,
      }),
    }
  )
);

