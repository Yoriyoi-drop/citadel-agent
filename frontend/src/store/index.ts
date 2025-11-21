// frontend/src/store/index.ts
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

interface Workflow {
  id: string;
  name: string;
  description: string;
  nodes: any[];
  edges: any[];
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

interface Execution {
  id: string;
  workflowId: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  startedAt: string;
  endedAt?: string;
  results?: any;
  error?: string;
}

interface WorkflowState {
  user: User | null;
  workflows: Workflow[];
  executions: Execution[];
  selectedWorkflow: Workflow | null;
  selectedExecution: Execution | null;
  loading: boolean;
  error: string | null;

  // User actions
  setUser: (user: User | null) => void;
  clearUser: () => void;

  // Workflow actions
  setWorkflows: (workflows: Workflow[]) => void;
  addWorkflow: (workflow: Workflow) => void;
  updateWorkflow: (id: string, workflow: Partial<Workflow>) => void;
  deleteWorkflow: (id: string) => void;
  selectWorkflow: (workflow: Workflow | null) => void;

  // Execution actions
  setExecutions: (executions: Execution[]) => void;
  addExecution: (execution: Execution) => void;
  updateExecution: (id: string, execution: Partial<Execution>) => void;
  selectExecution: (execution: Execution | null) => void;

  // Loading and error state
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

const useStore = create<WorkflowState>()(
  devtools((set, get) => ({
    user: null,
    workflows: [],
    executions: [],
    selectedWorkflow: null,
    selectedExecution: null,
    loading: false,
    error: null,

    setUser: (user) => set({ user }),
    clearUser: () => set({ user: null }),

    setWorkflows: (workflows) => set({ workflows }),
    addWorkflow: (workflow) => set((state) => ({ 
      workflows: [...state.workflows, workflow] 
    })),
    updateWorkflow: (id, updates) => set((state) => ({
      workflows: state.workflows.map(wf => 
        wf.id === id ? { ...wf, ...updates } : wf
      )
    })),
    deleteWorkflow: (id) => set((state) => ({
      workflows: state.workflows.filter(wf => wf.id !== id),
      selectedWorkflow: state.selectedWorkflow?.id === id ? null : state.selectedWorkflow
    })),
    selectWorkflow: (workflow) => set({ selectedWorkflow: workflow }),

    setExecutions: (executions) => set({ executions }),
    addExecution: (execution) => set((state) => ({ 
      executions: [...state.executions, execution] 
    })),
    updateExecution: (id, updates) => set((state) => ({
      executions: state.executions.map(ex => 
        ex.id === id ? { ...ex, ...updates } : ex
      )
    })),
    selectExecution: (execution) => set({ selectedExecution: execution }),

    setLoading: (loading) => set({ loading }),
    setError: (error) => set({ error }),
  }))
);

export default useStore;