// frontend/src/components/Sidebar/Sidebar.tsx
import { useState } from 'react';
import {
  HomeIcon,
  CogIcon,
  RectangleStackIcon,
  UserGroupIcon,
  DocumentTextIcon,
  CloudIcon,
  PlusIcon,
  TrashIcon,
  Square2StackIcon,
  ArrowsPointingOutIcon
} from '@heroicons/react/24/solid';

interface SidebarProps {
  selectedWorkflow: string | null;
  onWorkflowSelect: (id: string | null) => void;
}

interface Workflow {
  id: string;
  name: string;
  description: string;
  nodes: number;
  lastRun: string;
  status: 'active' | 'inactive' | 'error';
}

const Sidebar: React.FC<SidebarProps> = ({ selectedWorkflow, onWorkflowSelect }) => {
  const [workflows, setWorkflows] = useState<Workflow[]>([
    {
      id: '1',
      name: 'Email Notification Workflow',
      description: 'Automated email notifications',
      nodes: 5,
      lastRun: '2023-06-15T10:30:00Z',
      status: 'active'
    },
    {
      id: '2',
      name: 'Data Processing Pipeline',
      description: 'Process and transform data',
      nodes: 8,
      lastRun: '2023-06-15T09:15:00Z',
      status: 'active'
    },
    {
      id: '3',
      name: 'API Integration',
      description: 'Sync data with external APIs',
      nodes: 12,
      lastRun: '2023-06-14T16:45:00Z',
      status: 'error'
    }
  ]);

  const [showNewWorkflow, setShowNewWorkflow] = useState(false);
  const [newWorkflowName, setNewWorkflowName] = useState('');
  const [activeTab, setActiveTab] = useState('workflows');

  const createWorkflow = () => {
    if (newWorkflowName.trim()) {
      const newWorkflow: Workflow = {
        id: `wf - ${Date.now()} `,
        name: newWorkflowName,
        description: 'New workflow',
        nodes: 1,
        lastRun: new Date().toISOString(),
        status: 'active'
      };

      setWorkflows([newWorkflow, ...workflows]);
      setNewWorkflowName('');
      setShowNewWorkflow(false);
      onWorkflowSelect(newWorkflow.id);
    }
  };

  const duplicateWorkflow = (id: string) => {
    const workflowToDuplicate = workflows.find(w => w.id === id);
    if (workflowToDuplicate) {
      const duplicatedWorkflow: Workflow = {
        ...workflowToDuplicate,
        id: `wf - ${Date.now()} `,
        name: `${workflowToDuplicate.name} (Copy)`
      };
      setWorkflows([duplicatedWorkflow, ...workflows]);
    }
  };

  const deleteWorkflow = (id: string) => {
    setWorkflows(workflows.filter(w => w.id !== id));
    if (selectedWorkflow === id) {
      onWorkflowSelect(null);
    }
  };

  const getStatusColor = (status: Workflow['status']) => {
    switch (status) {
      case 'active': return 'bg-green-500';
      case 'inactive': return 'bg-gray-500';
      case 'error': return 'bg-red-500';
      default: return 'bg-gray-500';
    }
  };

  return (
    <div className="w-64 bg-white shadow-md flex flex-col">
      {/* Logo */}
      <div className="p-4 border-b">
        <div className="flex items-center">
          <ArrowsPointingOutIcon className="h-8 w-8 text-blue-600 mr-2" />
          <h1 className="text-xl font-bold text-gray-900">Citadel Agent</h1>
        </div>
      </div>

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto">
        <nav className="p-2">
          <div className="mb-6">
            <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-3 mb-2">Navigation</h2>
            <ul className="space-y-1">
              <li>
                <button
                  onClick={() => setActiveTab('workflows')}
                  className={`w - full flex items - center px - 3 py - 2 text - sm font - medium rounded - md ${activeTab === 'workflows'
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-700 hover:bg-gray-100'
                    } `}
                >
                  <RectangleStackIcon className="mr-3 h-5 w-5" />
                  Workflows
                </button>
              </li>
              <li>
                <button
                  onClick={() => setActiveTab('dashboard')}
                  className={`w - full flex items - center px - 3 py - 2 text - sm font - medium rounded - md ${activeTab === 'dashboard'
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-700 hover:bg-gray-100'
                    } `}
                >
                  <HomeIcon className="mr-3 h-5 w-5" />
                  Dashboard
                </button>
              </li>
              <li>
                <button
                  onClick={() => setActiveTab('nodes')}
                  className={`w - full flex items - center px - 3 py - 2 text - sm font - medium rounded - md ${activeTab === 'nodes'
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-700 hover:bg-gray-100'
                    } `}
                >
                  <CogIcon className="mr-3 h-5 w-5" />
                  Nodes
                </button>
              </li>
            </ul>
          </div>

          {activeTab === 'workflows' && (
            <div>
              <div className="flex justify-between items-center mb-3 px-3">
                <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Workflows</h2>
                <button
                  onClick={() => setShowNewWorkflow(!showNewWorkflow)}
                  className="text-blue-600 hover:text-blue-800"
                >
                  <PlusIcon className="h-4 w-4" />
                </button>
              </div>

              {showNewWorkflow && (
                <div className="mb-4 px-3">
                  <div className="flex space-x-2">
                    <input
                      type="text"
                      value={newWorkflowName}
                      onChange={(e) => setNewWorkflowName(e.target.value)}
                      placeholder="Workflow name"
                      className="flex-1 px-3 py-1 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      onKeyPress={(e) => e.key === 'Enter' && createWorkflow()}
                    />
                    <button
                      onClick={createWorkflow}
                      className="px-2 py-1 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700"
                    >
                      Create
                    </button>
                  </div>
                </div>
              )}

              <ul className="space-y-1">
                {workflows.map((workflow) => (
                  <li key={workflow.id}>
                    <div className={`flex items - center justify - between px - 3 py - 2 text - sm rounded - md cursor - pointer ${selectedWorkflow === workflow.id
                        ? 'bg-blue-100 text-blue-700'
                        : 'text-gray-700 hover:bg-gray-100'
                      } `}
                      onClick={() => onWorkflowSelect(workflow.id)}
                    >
                      <div className="flex items-center truncate">
                        <div className={`w - 2 h - 2 rounded - full mr - 2 ${getStatusColor(workflow.status)} `}></div>
                        <span className="truncate">{workflow.name}</span>
                      </div>
                      <div className="flex space-x-1">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            duplicateWorkflow(workflow.id);
                          }}
                          className="text-gray-500 hover:text-gray-700"
                        >
                          <Square2StackIcon className="h-4 w-4" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteWorkflow(workflow.id);
                          }}
                          className="text-gray-500 hover:text-red-500"
                        >
                          <TrashIcon className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {activeTab === 'nodes' && (
            <div>
              <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-3 mb-2">Node Categories</h2>
              <ul className="space-y-1">
                <li>
                  <a href="#" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-700 hover:bg-gray-100">
                    <CloudIcon className="mr-3 h-5 w-5" />
                    Integrations
                  </a>
                </li>
                <li>
                  <a href="#" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-700 hover:bg-gray-100">
                    <CogIcon className="mr-3 h-5 w-5" />
                    Utilities
                  </a>
                </li>
                <li>
                  <a href="#" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-700 hover:bg-gray-100">
                    <DocumentTextIcon className="mr-3 h-5 w-5" />
                    Data Processing
                  </a>
                </li>
                <li>
                  <a href="#" className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-gray-700 hover:bg-gray-100">
                    <UserGroupIcon className="mr-3 h-5 w-5" />
                    AI Agents
                  </a>
                </li>
              </ul>
            </div>
          )}
        </nav>
      </div>

      {/* User profile */}
      <div className="p-4 border-t border-gray-200">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className="h-8 w-8 rounded-full bg-blue-500 flex items-center justify-center text-white font-bold">
              U
            </div>
          </div>
          <div className="ml-3">
            <p className="text-sm font-medium text-gray-700">User Name</p>
            <p className="text-xs font-medium text-gray-500">user@example.com</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;