// frontend/src/components/workflow-builder/WorkflowBuilder.jsx
import React, { useCallback, useMemo } from 'react';
import ReactFlow, { 
  Controls, 
  Background, 
  useNodesState, 
  useEdgesState, 
  addEdge, 
  BackgroundVariant,
  MiniMap
} from 'reactflow';
import 'reactflow/dist/style.css';

import {
  StartNode,
  EndNode,
  HTTPNode,
  DatabaseNode,
  DecisionNode,
  DelayNode,
  AINode,
  NotificationNode,
  // Security Nodes
  FirewallManagerNode,
  EncryptionNode,
  AccessControlNode,
  APIKeyManagerNode,
  JWTHandlerNode,
  OAuth2ProviderNode,
  SecurityOperationNode
} from './nodes';

const nodeTypes = {
  start: StartNode,
  end: EndNode,
  http: HTTPNode,
  database: DatabaseNode,
  decision: DecisionNode,
  delay: DelayNode,
  ai: AINode,
  notification: NotificationNode,
  // Security Nodes
  firewallManager: FirewallManagerNode,
  encryption: EncryptionNode,
  accessControl: AccessControlNode,
  apiKeyManager: APIKeyManagerNode,
  jwtHandler: JWTHandlerNode,
  oauth2Provider: OAuth2ProviderNode,
  securityOperation: SecurityOperationNode,
};

const WorkflowBuilder = ({ initialWorkflow = null }) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  // Initialize with sample workflow if provided
  React.useEffect(() => {
    if (initialWorkflow) {
      setNodes(initialWorkflow.nodes || []);
      setEdges(initialWorkflow.edges || []);
    }
  }, [initialWorkflow, setNodes, setEdges]);

  const onConnect = useCallback((params) => {
    setEdges((eds) => addEdge({ ...params, animated: true }, eds));
  }, [setEdges]);

  const onLoad = useCallback((instance) => {
    // Auto zoom to fit content
    instance.fitView({ padding: 0.2 });
  }, []);

  // Node types configuration
  const nodeMenuItems = [
    { id: 'start', label: 'Start', type: 'start', icon: '▶️' },
    { id: 'end', label: 'End', type: 'end', icon: '⏹️' },
    { id: 'http', label: 'HTTP Request', type: 'http', icon: '🌐' },
    { id: 'database', label: 'Database Query', type: 'database', icon: '🗄️' },
    { id: 'decision', label: 'Decision', type: 'decision', icon: '❓' },
    { id: 'delay', label: 'Delay', type: 'delay', icon: '⏱️' },
    { id: 'ai', label: 'AI Agent', type: 'ai', icon: '🤖' },
    { id: 'notification', label: 'Notification', type: 'notification', icon: '🔔' },
    // Security Nodes
    { id: 'firewallManager', label: 'Firewall Manager', type: 'firewallManager', icon: '🛡️' },
    { id: 'encryption', label: 'Encryption/Decryption', type: 'encryption', icon: '🔒' },
    { id: 'accessControl', label: 'Access Control', type: 'accessControl', icon: '🔐' },
    { id: 'apiKeyManager', label: 'API Key Manager', type: 'apiKeyManager', icon: '🔑' },
    { id: 'jwtHandler', label: 'JWT Handler', type: 'jwtHandler', icon: '.JWT' },
    { id: 'oauth2Provider', label: 'OAuth2 Provider', type: 'oauth2Provider', icon: 'OAuth' },
    { id: 'securityOperation', label: 'Security Operations', type: 'securityOperation', icon: '🛡️' },
  ];

  const onDragStart = (event, nodeType) => {
    event.dataTransfer.setData('application/reactflow', nodeType);
    event.dataTransfer.effectAllowed = 'move';
  };

  // Helper functions for drag and drop
  const onDragOver = (event) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  };

  const onDrop = (event, rfInstance) => {
    event.preventDefault();

    const type = event.dataTransfer.getData('application/reactflow');

    if (typeof type === 'undefined' || !type) {
      return;
    }

    const position = rfInstance.screenToFlowPosition({
      x: event.clientX,
      y: event.clientY,
    });

    const newNode = {
      id: `node-${Date.now()}`,
      type,
      position,
      data: {
        label: getNodeLabel(type),
        config: {}
      },
    };

    rfInstance.addNodes(newNode);
  };

  // Helper function to generate proper labels based on node type
  const getNodeLabel = (type) => {
    const labelMap = {
      // Standard nodes
      'start': 'Start',
      'end': 'End',
      'http': 'HTTP Request',
      'database': 'Database Query',
      'decision': 'Decision',
      'delay': 'Delay',
      'ai': 'AI Agent',
      'notification': 'Notification',
      // Security nodes
      'firewallManager': 'Firewall Manager',
      'encryption': 'Encryption/Decryption',
      'accessControl': 'Access Control',
      'apiKeyManager': 'API Key Manager',
      'jwtHandler': 'JWT Handler',
      'oauth2Provider': 'OAuth2 Provider',
      'securityOperation': 'Security Operations',
    };

    return labelMap[type] || `${type.charAt(0).toUpperCase() + type.slice(1)} Node`;
  };

  return (
    <div className="workflow-builder-container flex h-screen bg-gray-50">
      {/* Node Palette */}
      <div className="w-64 bg-white shadow-lg border-r border-gray-200 p-4 overflow-y-auto">
        <h3 className="font-semibold text-lg mb-4 text-gray-800">Nodes</h3>
        <div className="space-y-2">
          {nodeMenuItems.map((item) => (
            <div
              key={item.id}
              draggable
              onDragStart={(event) => onDragStart(event, item.type)}
              className="flex items-center p-3 bg-gray-50 rounded-lg border border-gray-200 cursor-move hover:bg-blue-50 hover:border-blue-300 transition-colors"
            >
              <span className="text-xl mr-3">{item.icon}</span>
              <span className="font-medium text-gray-700">{item.label}</span>
            </div>
          ))}
        </div>

        <div className="mt-6">
          <h4 className="font-medium text-gray-700 mb-2">Controls</h4>
          <div className="space-y-2 text-sm text-gray-600">
            <div><kbd className="px-2 py-1 bg-gray-100 rounded">Click</kbd> Select node</div>
            <div><kbd className="px-2 py-1 bg-gray-100 rounded">Drag</kbd> Move node</div>
            <div><kbd className="px-2 py-1 bg-gray-100 rounded">Drag connection</kbd> Link nodes</div>
          </div>
        </div>
      </div>

      {/* Canvas */}
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onInit={onLoad}
          nodeTypes={nodeTypes}
          onDrop={onDrop}
          onDragOver={onDragOver}
          fitView
          attributionPosition="bottom-left"
          className="bg-gray-100"
        >
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
          <Controls />
          <MiniMap />
        </ReactFlow>
      </div>

      {/* Properties Panel */}
      <div className="w-80 bg-white shadow-lg border-l border-gray-200 p-4 overflow-y-auto">
        <h3 className="font-semibold text-lg mb-4 text-gray-800">Properties</h3>
        <div className="text-gray-500 text-center py-10">
          Select a node to view properties
        </div>
      </div>
    </div>
  );
};

export default WorkflowBuilder;