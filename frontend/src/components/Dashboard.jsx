// frontend/src/components/Dashboard.jsx
import React, { useState, useEffect } from 'react';
import { useWorkflow } from '../hooks/useWorkflow';
import WorkflowBuilder from './workflow-builder/WorkflowBuilder';
import PropertiesPanel from './workflow-builder/PropertiesPanel';
import { Plus, Play, Save, Trash2, Eye, Clock, AlertCircle } from 'lucide-react';

const Dashboard = () => {
  const { 
    workflows, 
    currentWorkflow, 
    loading, 
    error, 
    createWorkflow, 
    updateWorkflow, 
    deleteWorkflow,
    executeWorkflow,
    clearError
  } = useWorkflow();

  const [showBuilder, setShowBuilder] = useState(false);
  const [selectedWorkflow, setSelectedWorkflow] = useState(null);
  const [newWorkflowName, setNewWorkflowName] = useState('');

  const handleCreateWorkflow = async () => {
    if (!newWorkflowName.trim()) return;
    
    try {
      const workflowData = {
        name: newWorkflowName.trim(),
        description: `Workflow: ${newWorkflowName.trim()}`,
        nodes: [],
        edges: [],
        status: 'draft'
      };
      
      await createWorkflow(workflowData);
      setNewWorkflowName('');
    } catch (err) {
      console.error('Failed to create workflow:', err);
    }
  };

  const handleExecuteWorkflow = async (workflowId) => {
    if (!workflowId) return;
    
    try {
      await executeWorkflow(workflowId);
      // Show success notification
      alert(`Workflow ${workflowId} executed successfully!`);
    } catch (err) {
      console.error('Failed to execute workflow:', err);
      alert(`Failed to execute workflow: ${err.message}`);
    }
  };

  const handleDeleteWorkflow = async (workflowId) => {
    if (!workflowId || !confirm('Are you sure you want to delete this workflow?')) return;
    
    try {
      await deleteWorkflow(workflowId);
    } catch (err) {
      console.error('Failed to delete workflow:', err);
    }
  };

  const handleEditWorkflow = (workflow) => {
    setSelectedWorkflow(workflow);
    setShowBuilder(true);
  };

  const handleSaveWorkflow = async () => {
    if (!selectedWorkflow) return;
    
    try {
      await updateWorkflow(selectedWorkflow.id, selectedWorkflow);
      alert('Workflow saved successfully!');
    } catch (err) {
      console.error('Failed to save workflow:', err);
      alert(`Failed to save workflow: ${err.message}`);
    }
  };

  if (loading && workflows.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading workflows...</p>
        </div>
      </div>
    );
  }

  if (showBuilder) {
    return (
      <div className="h-screen flex flex-col">
        {/* Top bar */}
        <div className="bg-white shadow-sm border-b border-gray-200 px-6 py-4 flex justify-between items-center">
          <div>
            <h1 className="text-xl font-semibold text-gray-900">
              {selectedWorkflow?.name || 'New Workflow'}
            </h1>
            <p className="text-sm text-gray-500">
              Visual workflow editor
            </p>
          </div>
          <div className="flex space-x-2">
            <button
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              onClick={handleSaveWorkflow}
            >
              <Save className="w-4 h-4 mr-2" />
              Save
            </button>
            <button
              className="flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500"
              onClick={() => selectedWorkflow && handleExecuteWorkflow(selectedWorkflow.id)}
            >
              <Play className="w-4 h-4 mr-2" />
              Run
            </button>
            <button
              className="flex items-center px-4 py-2 bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500"
              onClick={() => setShowBuilder(false)}
            >
              Back
            </button>
          </div>
        </div>
        
        {/* Workflow Builder */}
        <div className="flex-1">
          <WorkflowBuilder initialWorkflow={selectedWorkflow} />
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Workflow Dashboard</h1>
              <p className="mt-2 text-gray-600">Manage and execute your automated workflows</p>
            </div>
            <div className="flex space-x-3">
              <div className="relative">
                <input
                  type="text"
                  placeholder="Search workflows..."
                  className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
                <svg className="absolute left-3 top-2.5 h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
              <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500">
                <Plus className="w-4 h-4 mr-2" />
                New Workflow
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Error Notification */}
      {error && (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-4">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-center">
            <AlertCircle className="h-5 w-5 text-red-400 mr-3" />
            <div>
              <p className="text-sm font-medium text-red-800">Error</p>
              <p className="text-sm text-red-700">{error}</p>
            </div>
            <button 
              onClick={clearError} 
              className="ml-auto text-red-500 hover:text-red-700"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="rounded-full bg-blue-100 p-3">
                <svg className="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
              </div>
              <div className="ml-4">
                <h3 className="text-2xl font-semibold text-gray-900">{workflows.length}</h3>
                <p className="text-gray-600">Total Workflows</p>
              </div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="rounded-full bg-green-100 p-3">
                <Play className="h-6 w-6 text-green-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-2xl font-semibold text-gray-900">124</h3>
                <p className="text-gray-600">Executions Today</p>
              </div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="rounded-full bg-yellow-100 p-3">
                <Clock className="h-6 w-6 text-yellow-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-2xl font-semibold text-gray-900">89%</h3>
                <p className="text-gray-600">Success Rate</p>
              </div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="rounded-full bg-purple-100 p-3">
                <Eye className="h-6 w-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-2xl font-semibold text-gray-900">23</h3>
                <p className="text-gray-600">Active Workflows</p>
              </div>
            </div>
          </div>
        </div>

        {/* New Workflow Form */}
        <div className="bg-white rounded-lg shadow p-6 mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Create New Workflow</h2>
          <div className="flex space-x-3">
            <input
              type="text"
              placeholder="Enter workflow name..."
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              value={newWorkflowName}
              onChange={(e) => setNewWorkflowName(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleCreateWorkflow()}
            />
            <button
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              onClick={handleCreateWorkflow}
            >
              Create
            </button>
          </div>
        </div>

        {/* Workflows List */}
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Your Workflows</h2>
          </div>
          
          {workflows.length === 0 ? (
            <div className="p-12 text-center">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">No workflows</h3>
              <p className="mt-1 text-sm text-gray-500">Get started by creating a new workflow.</p>
              <div className="mt-6">
                <button
                  className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  onClick={() => {}}
                >
                  <Plus className="w-4 h-4 mr-2" />
                  New Workflow
                </button>
              </div>
            </div>
          ) : (
            <ul className="divide-y divide-gray-200">
              {workflows.map((workflow) => (
                <li key={workflow.id} className="hover:bg-gray-50 transition-colors">
                  <div className="px-6 py-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        <div className="flex-shrink-0 h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center">
                          <svg className="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                          </svg>
                        </div>
                        <div className="ml-4">
                          <h3 className="text-sm font-medium text-gray-900">{workflow.name}</h3>
                          <p className="text-sm text-gray-500">{workflow.description}</p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-4">
                        <div className="text-sm text-gray-500">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            workflow.status === 'active' 
                              ? 'bg-green-100 text-green-800' 
                              : workflow.status === 'draft' 
                                ? 'bg-yellow-100 text-yellow-800' 
                                : 'bg-red-100 text-red-800'
                          }`}>
                            {workflow.status}
                          </span>
                        </div>
                        <div className="text-sm text-gray-500">
                          {workflow.nodes?.length || 0} nodes
                        </div>
                        <div className="text-sm text-gray-500">
                          {new Date(workflow.created_at || workflow.createdAt).toLocaleDateString()}
                        </div>
                        <div className="flex space-x-2">
                          <button
                            className="text-blue-600 hover:text-blue-900"
                            onClick={() => handleEditWorkflow(workflow)}
                          >
                            <Eye className="h-5 w-5" />
                          </button>
                          <button
                            className="text-green-600 hover:text-green-900"
                            onClick={() => handleExecuteWorkflow(workflow.id)}
                          >
                            <Play className="h-5 w-5" />
                          </button>
                          <button
                            className="text-red-600 hover:text-red-900"
                            onClick={() => handleDeleteWorkflow(workflow.id)}
                          >
                            <Trash2 className="h-5 w-5" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;