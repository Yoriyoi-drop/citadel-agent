import React, { useCallback, useRef, useState } from 'react';
import ReactFlow, {
  ReactFlowProvider,
  addEdge,
  useEdgesState,
  useNodesState,
  Connection,
  Edge,
  Node,
  Controls,
  ControlButton,
  Panel,
  NodeTypes,
  MiniMap,
  Background,
  BackgroundVariant,
  useReactFlow,
  ReactFlowInstance
} from 'reactflow';
import 'reactflow/dist/style.css';

import NodeItem from './NodeItem';
import EdgeItem from './EdgeItem';
import MiniMapComponent from './MiniMap';

// Define custom node types
const nodeTypes: NodeTypes = {
  default: NodeItem,
  trigger: NodeItem,
  action: NodeItem,
  condition: NodeItem,
  loop: NodeItem,
};

// Define initial elements for the canvas
const initialNodes: Node[] = [
  {
    id: '1',
    type: 'trigger',
    position: { x: 0, y: 0 },
    data: { 
      label: 'Trigger', 
      description: 'Starts the workflow', 
      type: 'trigger',
      parameters: {}
    },
  },
  {
    id: '2',
    type: 'action',
    position: { x: 200, y: 0 },
    data: { 
      label: 'HTTP Request', 
      description: 'Make an HTTP request', 
      type: 'action',
      parameters: {
        method: 'GET',
        url: 'https://api.example.com/data'
      }
    },
  },
  {
    id: '3',
    type: 'action',
    position: { x: 400, y: 0 },
    data: { 
      label: 'Data Process', 
      description: 'Process the received data', 
      type: 'action',
      parameters: {
        operation: 'transform',
        transformation: 'uppercase'
      }
    },
  },
];

const initialEdges: Edge[] = [
  { id: 'e1-2', source: '1', target: '2', animated: true },
  { id: 'e2-3', source: '2', target: '3', animated: true },
];

interface WorkflowCanvasProps {
  onSave?: (nodes: Node[], edges: Edge[]) => void;
  onRun?: () => void;
}

const Canvas: React.FC<WorkflowCanvasProps> = ({ onSave, onRun }) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  
  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, animated: true }, eds)),
    [setEdges]
  );

  const onLoad = useCallback((rf: ReactFlowInstance) => {
    setReactFlowInstance(rf);
  }, []);

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  const addNode = (type: string) => {
    const newNode: Node = {
      id: `node-${Date.now()}`,
      type: type,
      position: { x: 0, y: 0 },
      data: { 
        label: `${type.charAt(0).toUpperCase() + type.slice(1)} Node`, 
        description: `A ${type} node`, 
        type: type,
        parameters: {}
      },
    };

    setNodes((nds) => nds.concat(newNode));
  };

  const deleteSelectedNode = () => {
    if (selectedNode) {
      setNodes((nds) => nds.filter((node) => node.id !== selectedNode.id));
      setEdges((eds) => eds.filter(
        (edge) => edge.source !== selectedNode.id && edge.target !== selectedNode.id
      ));
      setSelectedNode(null);
    }
  };

  const handleSave = () => {
    if (onSave) {
      onSave(nodes, edges);
    }
  };

  const handleRun = () => {
    if (onRun) {
      onRun();
    }
  };

  return (
    <div className="flex flex-col h-full w-full">
      <Panel position="top-center" className="bg-white p-2 rounded shadow-md">
        <div className="flex space-x-2">
          <button 
            onClick={handleSave}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Save Workflow
          </button>
          <button 
            onClick={handleRun}
            className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
          >
            Run Workflow
          </button>
        </div>
      </Panel>

      <div className="flex-1" ref={reactFlowWrapper}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onInit={onLoad}
          onNodeClick={onNodeClick}
          onPaneClick={onPaneClick}
          nodeTypes={nodeTypes}
          connectionMode="loose"
          onConnectStart={() => console.log('Connection started')}
          onConnectEnd={() => console.log('Connection ended')}
          fitView
          fitViewOptions={{ padding: 0.5 }}
        >
          <Controls>
            <ControlButton onClick={() => addNode('trigger')} title="Add Trigger">
              +T
            </ControlButton>
            <ControlButton onClick={() => addNode('action')} title="Add Action">
              +A
            </ControlButton>
            <ControlButton onClick={() => addNode('condition')} title="Add Condition">
              +C
            </ControlButton>
            <ControlButton onClick={() => addNode('loop')} title="Add Loop">
              +L
            </ControlButton>
          </Controls>
          
          <MiniMapComponent />
          
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
        </ReactFlow>
      </div>

      {selectedNode && (
        <Panel position="right" className="w-80 bg-white p-4 shadow-lg border rounded">
          <h3 className="font-bold text-lg mb-2">Node Properties</h3>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Label</label>
            <input
              type="text"
              value={selectedNode.data.label}
              onChange={(e) => {
                const updatedNode = {
                  ...selectedNode,
                  data: { ...selectedNode.data, label: e.target.value }
                };
                setNodes((nds) => nds.map(n => n.id === selectedNode.id ? updatedNode : n));
              }}
              className="w-full p-2 border rounded"
            />
          </div>
          
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Description</label>
            <textarea
              value={selectedNode.data.description}
              onChange={(e) => {
                const updatedNode = {
                  ...selectedNode,
                  data: { ...selectedNode.data, description: e.target.value }
                };
                setNodes((nds) => nds.map(n => n.id === selectedNode.id ? updatedNode : n));
              }}
              className="w-full p-2 border rounded h-20"
            />
          </div>
          
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Parameters</label>
            <textarea
              value={JSON.stringify(selectedNode.data.parameters, null, 2)}
              onChange={(e) => {
                try {
                  const params = JSON.parse(e.target.value);
                  const updatedNode = {
                    ...selectedNode,
                    data: { ...selectedNode.data, parameters: params }
                  };
                  setNodes((nds) => nds.map(n => n.id === selectedNode.id ? updatedNode : n));
                } catch (error) {
                  // Ignore invalid JSON
                }
              }}
              className="w-full p-2 border rounded h-32 font-mono text-xs"
            />
          </div>
          
          <button
            onClick={deleteSelectedNode}
            className="w-full py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Delete Node
          </button>
        </Panel>
      )}
    </div>
  );
};

const WorkflowCanvas: React.FC<WorkflowCanvasProps> = (props) => {
  return (
    <ReactFlowProvider>
      <Canvas {...props} />
    </ReactFlowProvider>
  );
};

export default WorkflowCanvas;