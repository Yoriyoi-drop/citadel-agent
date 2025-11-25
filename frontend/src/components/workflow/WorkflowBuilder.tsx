"use client"

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
    ReactFlow,
    Background,
    Controls,
    MiniMap,
    ReactFlowProvider,
    addEdge,
    Connection,
    Edge,
    Node,
    useNodesState,
    useEdgesState,
    NodeChange,
    EdgeChange,
    applyNodeChanges,
    applyEdgeChanges,
    Panel
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { useWorkflowStore } from '@/stores/workflowStore';
import BaseNode from '../nodes/BaseNode';
import { NodePalette } from './NodePalette';
import { NodeEditor } from './NodeEditor';
import { NodeType } from '@/types/workflow';
import CustomEdge from './CustomEdge';
import ConnectionLineComponent from './ConnectionLine';

// Use a Proxy to handle any node type by rendering BaseNode
// This allows us to support hundreds of node types without explicitly registering each one
const nodeTypes = new Proxy({}, {
    get: (target, prop) => BaseNode
});

// Edge types configuration
const edgeTypes = {
    default: CustomEdge,
};

interface WorkflowBuilderProps {
    workflowId?: string;
}

function WorkflowBuilderContent({ workflowId }: WorkflowBuilderProps) {
    const {
        currentWorkflow,
        updateNode,
        addNode,
        selectNodes,
        updateWorkflow
    } = useWorkflowStore();

    // Local state for ReactFlow (synced with store)
    const [nodes, setNodes] = useNodesState<Node>([]);
    const [edges, setEdges] = useEdgesState<Edge>([]);
    const [reactFlowInstance, setReactFlowInstance] = useState<any>(null);

    // Sync from store to local state when workflow changes
    useEffect(() => {
        if (currentWorkflow) {
            setNodes(currentWorkflow.nodes.map(n => ({
                ...n,
                // Ensure type is preserved or defaults to the specific type
                type: n.type || 'default',
                data: {
                    ...n.data,
                    nodeType: n.type || 'default',
                    // Pass selected state if needed, though ReactFlow handles it
                }
            })));
            setEdges(currentWorkflow.edges);
        }
    }, [currentWorkflow, setNodes, setEdges]);

    const onNodesChange = useCallback(
        (changes: NodeChange[]) => {
            setNodes((nds) => applyNodeChanges(changes, nds));

            // Sync changes back to store (debounced in a real app, but direct here for simplicity)
            changes.forEach(change => {
                if (change.type === 'position' && change.position && currentWorkflow) {
                    updateNode(change.id, { position: change.position });
                }
                if (change.type === 'select') {
                    // Handle selection in store if needed
                    if (change.selected) {
                        selectNodes([change.id]);
                    }
                }
            });
        },
        [setNodes, updateNode, currentWorkflow, selectNodes]
    );

    const onEdgesChange = useCallback(
        (changes: EdgeChange[]) => {
            setEdges((eds) => applyEdgeChanges(changes, eds));
        },
        [setEdges]
    );

    const onConnect = useCallback(
        (params: Connection) => {
            setEdges((eds) => addEdge(params, eds));
            // Update store
            if (currentWorkflow) {
                const newEdge = {
                    id: `e${params.source}-${params.target}`,
                    source: params.source,
                    target: params.target,
                    sourceHandle: params.sourceHandle || '',
                    targetHandle: params.targetHandle || ''
                };
                // We would need an addEdge action in store, but for now we can update the workflow
                const updatedEdges = [...currentWorkflow.edges, newEdge];
                updateWorkflow(currentWorkflow.id, { edges: updatedEdges });
            }
        },
        [setEdges, currentWorkflow, updateWorkflow]
    );

    const onDragOver = useCallback((event: React.DragEvent) => {
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
    }, []);

    const onDrop = useCallback(
        (event: React.DragEvent) => {
            event.preventDefault();

            if (!reactFlowInstance) {
                console.warn('ReactFlow instance not ready');
                return;
            }

            const dataStr = event.dataTransfer.getData('application/reactflow');
            if (!dataStr) {
                console.warn('No drag data found');
                return;
            }

            try {
                const data = JSON.parse(dataStr);
                const position = reactFlowInstance.screenToFlowPosition({
                    x: event.clientX,
                    y: event.clientY,
                });

                const newNode: Node = {
                    id: `${data.nodeType}_${Date.now()}`,
                    type: data.nodeType,
                    position,
                    data: {
                        label: data.label,
                        nodeType: data.nodeType,
                        inputs: data.inputs || [],
                        outputs: data.outputs || [],
                        config: data.config || {},
                        status: 'idle' as const
                    },
                };

                console.log('Adding node:', newNode);
                addNode(newNode as any);
            } catch (error) {
                console.error('Error dropping node:', error);
            }
        },
        [reactFlowInstance, addNode]
    );

    const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
        selectNodes([node.id]);
    }, [selectNodes]);

    const onPaneClick = useCallback(() => {
        selectNodes([]);
    }, [selectNodes]);

    // Find selected node for the editor
    const selectedNodeId = currentWorkflow?.nodes.find(n =>
        nodes.find(ln => ln.id === n.id && ln.selected)
    )?.id;

    return (
        <div className="flex h-full w-full">
            {/* Palette */}
            <div className="w-64 border-r bg-muted/20 flex-shrink-0">
                <NodePalette />
            </div>

            {/* Canvas */}
            <div className="flex-1 h-full relative">
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onNodesChange={onNodesChange}
                    onEdgesChange={onEdgesChange}
                    onConnect={onConnect}
                    onInit={setReactFlowInstance}
                    onDrop={onDrop}
                    onDragOver={onDragOver}
                    onNodeClick={onNodeClick}
                    onPaneClick={onPaneClick}
                    nodeTypes={nodeTypes}
                    edgeTypes={edgeTypes}
                    connectionLineComponent={ConnectionLineComponent}
                    panOnScroll={true}
                    zoomOnScroll={true}
                    zoomOnPinch={true}
                    panOnDrag={true}
                    minZoom={0.2}
                    maxZoom={2}
                    className="bg-background"
                >
                    <Background color="#888" gap={16} size={1} />
                    <Controls />
                    <MiniMap />

                    {/* Floating Panel for Workflow Info or Controls could go here */}
                </ReactFlow>
            </div>

            {/* Editor Sidebar (Right) */}
            {selectedNodeId && (
                <div className="w-80 border-l bg-background shadow-xl z-10">
                    <NodeEditor nodeId={selectedNodeId} />
                </div>
            )}
        </div>
    );
}

export function WorkflowBuilder(props: WorkflowBuilderProps) {
    return (
        <ReactFlowProvider>
            <WorkflowBuilderContent {...props} />
        </ReactFlowProvider>
    );
}

export default WorkflowBuilder;
