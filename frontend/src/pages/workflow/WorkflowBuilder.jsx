// frontend/src/pages/workflow/WorkflowBuilder.jsx
import React, { useState, useCallback } from 'react';
import ReactFlow, {
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
} from 'reactflow';
import 'reactflow/dist/style.css';

// Node types
const nodeTypes = {
  http_request: {
    label: 'HTTP Request',
    icon: '🌐',
    inputs: ['url', 'method', 'headers', 'body'],
    outputs: ['response', 'status_code']
  },
  condition: {
    label: 'Condition',
    icon: '❓',
    inputs: ['condition', 'value1', 'value2'],
    outputs: ['true', 'false']
  },
  delay: {
    label: 'Delay',
    icon: '⏱️',
    inputs: ['duration'],
    outputs: ['completed']
  },
  database_query: {
    label: 'Database Query',
    icon: '💾',
    inputs: ['query', 'params'],
    outputs: ['result', 'error']
  },
  ai_agent: {
    label: 'AI Agent',
    icon: '🤖',
    inputs: ['prompt', 'context'],
    outputs: ['result', 'thoughts']
  }
};

const nodeTypeList = Object.keys(nodeTypes).map(key => ({
  id: key,
  ...nodeTypes[key]
}));

// Custom node component
const CustomNode = ({ data, id }) => {
  return (
    <div className="bg-white border-2 border-gray-200 rounded-lg shadow-md min-w-[200px]">
      <div className="bg-gray-50 px-3 py-2 rounded-t-md border-b border-gray-200">
        <div className="flex items-center">
          <span className="mr-2 text-lg">{data.icon}</span>
          <h3 className="font-medium text-gray-900">{data.label}</h3>
        </div>
      </div>
      <div className="p-3">
        <div className="mb-2">
          <label className="block text-xs font-medium text-gray-500 mb-1">Name</label>
          <input
            type="text"
            value={data.name || id}
            onChange={(e) => data.onNameChange?.(id, e.target.value)}
            className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
          />
        </div>
        
        {data.config && Object.entries(data.config).map(([key, value]) => (
          <div key={key} className="mb-2">
            <label className="block text-xs font-medium text-gray-500 mb-1">
              {key.charAt(0).toUpperCase() + key.slice(1)}
            </label>
            <input
              type="text"
              value={value}
              onChange={(e) => data.onConfigChange?.(id, key, e.target.value)}
              className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        ))}
      </div>
    </div>
  );
};

const nodeTypesMap = {
  default: CustomNode,
};

const initialNodes = [
  {
    id: '1',
    type: 'default',
    position: { x: 0, y: 0 },
    data: {
      label: 'HTTP Request',
      icon: '🌐',
      name: 'Start',
      config: { url: 'https://api.example.com/data', method: 'GET' },
      onNameChange: () => {},
      onConfigChange: () => {}
    },
  },
];

const initialEdges = [];

const WorkflowBuilder = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [nodeType, setNodeType] = useState('http_request');
  const [workflowName, setWorkflowName] = useState('New Workflow');
  const [workflowDescription, setWorkflowDescription] = useState('');

  const onConnect = useCallback(
    (params) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const addNode = (type) => {
    const newNode = {
      id: `${Date.now()}`,
      type: 'default',
      position: { 
        x: Math.random() * 400 + 200, 
        y: Math.random() * 300 + 50 
      },
      data: {
        label: nodeTypes[type].label,
        icon: nodeTypes[type].icon,
        name: nodeTypes[type].label,
        config: {},
        onNameChange: (id, name) => {
          setNodes(nds => 
            nds.map(node => 
              node.id === id ? { ...node, data: { ...node.data, name } } : node
            )
          );
        },
        onConfigChange: (nodeId, key, value) => {
          setNodes(nds => 
            nds.map(node => 
              node.id === nodeId 
                ? { ...node, data: { ...node.data, config: { ...node.data.config, [key]: value } } } 
                : node
            )
          );
        }
      },
    };

    setNodes((nds) => [...nds, newNode]);
  };

  const saveWorkflow = () => {
    // Simulate saving workflow
    alert(`Workflow "${workflowName}" saved successfully!`);
    console.log('Saving workflow:', { workflowName, workflowDescription, nodes, edges });
  };

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <div className="bg-white shadow-sm border-b border-gray-200 p-4">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <input
              type="text"
              value={workflowName}
              onChange={(e) => setWorkflowName(e.target.value)}
              className="text-xl font-bold text-gray-900 bg-transparent border-none focus:outline-none focus:ring-0 p-0"
              placeholder="Workflow name"
            />
            <textarea
              value={workflowDescription}
              onChange={(e) => setWorkflowDescription(e.target.value)}
              className="mt-1 text-sm text-gray-600 bg-transparent border-none focus:outline-none focus:ring-0 p-0 w-full resize-none"
              placeholder="Workflow description"
              rows="2"
            />
          </div>
          <div className="flex items-center space-x-3">
            <button
              onClick={saveWorkflow}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              Save Workflow
            </button>
            <button className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2">
              Run Workflow
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex">
        {/* Sidebar */}
        <div className="w-64 bg-gray-50 border-r border-gray-200 p-4 overflow-y-auto">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Node Library</h3>
          <div className="space-y-2">
            {nodeTypeList.map((node) => (
              <button
                key={node.id}
                onClick={() => addNode(node.id)}
                className="w-full text-left p-3 bg-white rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors duration-200"
              >
                <div className="flex items-center">
                  <span className="text-lg mr-2">{node.icon}</span>
                  <span className="font-medium text-gray-900">{node.label}</span>
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  {node.inputs.length} inputs, {node.outputs.length} outputs
                </p>
              </button>
            ))}
          </div>

          {/* Workflow Stats */}
          <div className="mt-8">
            <h4 className="text-md font-medium text-gray-900 mb-2">Workflow Stats</h4>
            <div className="bg-white rounded-lg p-4 border border-gray-200">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Nodes:</span>
                <span className="font-medium">{nodes.length}</span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-600">Connections:</span>
                <span className="font-medium">{edges.length}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Flow Canvas */}
        <div className="flex-1">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            nodeTypes={nodeTypesMap}
            fitView
          >
            <Controls />
            <Background variant="dots" gap={12} size={1} />
          </ReactFlow>
        </div>
      </div>
    </div>
  );
};

export default WorkflowBuilder;