"use client"

import { useCallback, useMemo } from 'react';
import {
  ReactFlow,
  Node,
  Edge,
  addEdge,
  useNodesState,
  useEdgesState,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  Connection,
  Panel,
  ViewportPortal,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useWorkflowStore } from '@/stores/workflowStore';
import { BaseNode as BaseNodeType } from '@/types/workflow';
import BaseNode from '../nodes/BaseNode';
import ConnectionLine from './ConnectionLine';
import { NodePalette } from './NodePalette';
import { NodeEditor } from './NodeEditor';

// Node types mapping
const nodeTypes = {
  base: BaseNode,
};

interface WorkflowBuilderProps {
  workflowId?: string;
}

export function WorkflowBuilder({ workflowId }: WorkflowBuilderProps) {
  const { 
    currentWorkflow, 
    addNode, 
    updateNode, 
    deleteNode, 
    addEdge, 
    updateEdge, 
    deleteEdge,
    selectedNodes,
    selectedEdges,
    selectNodes,
    selectEdges
  } = useWorkflowStore();

  // Convert workflow nodes to React Flow nodes
  const initialNodes = useMemo(() => {
    if (!currentWorkflow) return [];
    
    return currentWorkflow.nodes.map((node): Node => ({
      id: node.id,
      type: 'base',
      position: node.position,
      data: {
        ...node.data,
        nodeType: node.type,
        onUpdate: (updates: Partial<BaseNodeType>) => {
          updateNode(node.id, updates);
        },
        onDelete: () => {
          deleteNode(node.id);
        },
        isSelected: selectedNodes.includes(node.id)
      }
    }));
  }, [currentWorkflow, selectedNodes, updateNode, deleteNode]);

  // Convert workflow edges to React Flow edges
  const initialEdges = useMemo(() => {
    if (!currentWorkflow) return [];
    
    return currentWorkflow.edges.map((edge): Edge => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle,
      targetHandle: edge.targetHandle,
      type: edge.type || 'default',
      animated: true,
      selected: selectedEdges.includes(edge.id)
    }));
  }, [currentWorkflow, selectedEdges]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Handle new connections
  const onConnect = useCallback(
    (params: Connection) => {
      const newEdge = {
        id: `edge_${Date.now()}`,
        source: params.source!,
        target: params.target!,
        sourceHandle: params.sourceHandle,
        targetHandle: params.targetHandle,
        type: 'default'
      };
      
      addEdge(newEdge);
    },
    [addEdge]
  );

  // Handle node selection
  const onSelectionChange = useCallback(
    ({ nodes: selectedNodesList, edges: selectedEdgesList }: { nodes: Node[], edges: Edge[] }) => {
      const selectedNodeIds = selectedNodesList.map(n => n.id);
      const selectedEdgeIds = selectedEdgesList.map(e => e.id);
      
      selectNodes(selectedNodeIds);
      selectEdges(selectedEdgeIds);
    },
    [selectNodes, selectEdges]
  );

  // Handle node drop from palette
  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const reactFlowBounds = event.currentTarget.getBoundingClientRect();
      const nodeData = JSON.parse(event.dataTransfer.getData('application/reactflow'));

      if (!nodeData.nodeType) return;

      const position = {
        x: event.clientX - reactFlowBounds.left - 75,
        y: event.clientY - reactFlowBounds.top - 25,
      };

      const newNode: BaseNodeType = {
        id: `${nodeData.nodeType}_${Date.now()}`,
        type: nodeData.nodeType,
        position,
        data: {
          label: nodeData.label,
          description: nodeData.description,
          inputs: nodeData.inputs || [],
          outputs: nodeData.outputs || [],
          config: nodeData.config || {},
          status: 'idle'
        }
      };

      addNode(newNode);
    },
    [addNode]
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  return (
    <div className="flex h-full">
      {/* Node Palette */}
      <div className="w-80 border-r bg-card">
        <NodePalette />
      </div>

      {/* Workflow Canvas */}
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onSelectionChange={onSelectionChange}
          onDrop={onDrop}
          onDragOver={onDragOver}
          nodeTypes={nodeTypes}
          connectionLineComponent={ConnectionLine}
          fitView
          attributionPosition="bottom-left"
        >
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
          <Controls />
          <MiniMap 
            nodeColor={(node) => {
              const status = node.data?.status;
              switch (status) {
                case 'running': return '#3b82f6';
                case 'success': return '#10b981';
                case 'error': return '#ef4444';
                default: return '#6b7280';
              }
            }}
            className="bg-background border"
          />
          
          {/* Top Toolbar */}
          <Panel position="top-left" className="bg-background border rounded-lg p-2 shadow-lg">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Workflow Builder</span>
              {currentWorkflow && (
                <>
                  <span className="text-muted-foreground">•</span>
                  <span className="text-sm text-muted-foreground">
                    {currentWorkflow.nodes.length} nodes, {currentWorkflow.edges.length} connections
                  </span>
                </>
              )}
            </div>
          </Panel>

          {/* Zoom Controls */}
          <Panel position="bottom-right" className="bg-background border rounded-lg p-2 shadow-lg">
            <div className="flex items-center space-x-2 text-sm">
              <span className="text-muted-foreground">Zoom:</span>
              <span className="font-medium">100%</span>
            </div>
          </Panel>
        </ReactFlow>
      </div>

      {/* Node Editor Panel */}
      {selectedNodes.length === 1 && (
        <div className="w-96 border-l bg-card">
          <NodeEditor nodeId={selectedNodes[0]} />
        </div>
      )}
    </div>
  );
}