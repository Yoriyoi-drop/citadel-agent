import React, { memo, useState, useCallback, useMemo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger
} from '@/components/ui/tooltip';
import {
    Play,
    Copy,
    Trash2,
    Settings,
    Zap,
    CheckCircle2,
    XCircle,
    Loader2,
    AlertCircle,
    Sparkles,
    TrendingUp
} from 'lucide-react';
import { useWorkflowStore } from '@/stores/workflowStore';
import { NodeIconBadge } from '@/components/NodeIcon';

// Enhanced node data interface
interface EnhancedNodeData {
    label: string;
    nodeType: string;
    inputs: any[];
    outputs: any[];
    config?: any;
    status?: 'idle' | 'running' | 'success' | 'error' | 'warning';
    outputData?: any;
    error?: string;
    executionTime?: number;
    lastExecuted?: Date;
    aiSuggestions?: string[];
    performance?: {
        avgTime: number;
        successRate: number;
        totalRuns: number;
    };
}

type BaseNodeComponentProps = NodeProps<EnhancedNodeData>;

// Category colors with gradients
const getCategoryColor = (category: string) => {
    const colors = {
        trigger: { base: '#10b981', gradient: 'from-emerald-500 to-green-600' },
        http: { base: '#3b82f6', gradient: 'from-blue-500 to-indigo-600' },
        database: { base: '#8b5cf6', gradient: 'from-purple-500 to-violet-600' },
        ai: { base: '#ec4899', gradient: 'from-pink-500 to-rose-600' },
        transform: { base: '#f59e0b', gradient: 'from-amber-500 to-orange-600' },
        communication: { base: '#06b6d4', gradient: 'from-cyan-500 to-teal-600' },
        utility: { base: '#64748b', gradient: 'from-slate-500 to-gray-600' },
        flow: { base: '#0ea5e9', gradient: 'from-sky-500 to-blue-600' },
    };
    return colors[category as keyof typeof colors] || colors.utility;
};

const getCategoryFromType = (nodeType: string): string => {
    if (nodeType.includes('http') || nodeType.includes('webhook')) return 'http';
    if (nodeType.includes('db') || nodeType.includes('database') || nodeType.includes('sql')) return 'database';
    if (nodeType.includes('ai') || nodeType.includes('llm') || nodeType.includes('gpt')) return 'ai';
    if (nodeType.includes('transform') || nodeType.includes('map') || nodeType.includes('filter')) return 'transform';
    if (nodeType.includes('email') || nodeType.includes('slack') || nodeType.includes('telegram')) return 'communication';
    if (nodeType.includes('if') || nodeType.includes('switch') || nodeType.includes('delay')) return 'flow';
    if (nodeType.includes('trigger') || nodeType.includes('schedule')) return 'trigger';
    return 'utility';
};

const BaseNodeComponent = memo(({ id, data, selected }: BaseNodeComponentProps) => {
    const { deleteNode, duplicateNode, updateNode } = useWorkflowStore();
    const [isHovered, setIsHovered] = useState(false);
    const [showAI, setShowAI] = useState(false);

    const category = useMemo(() => getCategoryFromType(data.nodeType), [data.nodeType]);
    const categoryColor = useMemo(() => getCategoryColor(category), [category]);

    // Status styling with enhanced visuals
    const getStatusStyle = () => {
        const baseClasses = 'transition-all duration-300';
        switch (data.status) {
            case 'running':
                return `${baseClasses} ring-2 ring-blue-400 ring-offset-2 shadow-lg shadow-blue-500/50 animate-pulse`;
            case 'success':
                return `${baseClasses} ring-2 ring-green-400 ring-offset-2 shadow-lg shadow-green-500/50`;
            case 'error':
                return `${baseClasses} ring-2 ring-red-400 ring-offset-2 shadow-lg shadow-red-500/50`;
            case 'warning':
                return `${baseClasses} ring-2 ring-yellow-400 ring-offset-2 shadow-lg shadow-yellow-500/50`;
            default:
                return `${baseClasses} hover:shadow-xl hover:scale-[1.02] ${selected ? 'ring-2 ring-primary ring-offset-2 shadow-xl' : 'shadow-md'}`;
        }
    };

    const getStatusIcon = () => {
        switch (data.status) {
            case 'running':
                return <Loader2 className="w-3.5 h-3.5 text-blue-600 animate-spin" />;
            case 'success':
                return <CheckCircle2 className="w-3.5 h-3.5 text-green-600" />;
            case 'error':
                return <XCircle className="w-3.5 h-3.5 text-red-600" />;
            case 'warning':
                return <AlertCircle className="w-3.5 h-3.5 text-yellow-600" />;
            default:
                return null;
        }
    };

    const handleRun = useCallback((e: React.MouseEvent) => {
        e.stopPropagation();
        const startTime = Date.now();

        updateNode(id, { data: { ...data, status: 'running' } });

        // Simulate execution with realistic timing
        setTimeout(() => {
            const executionTime = Date.now() - startTime;
            const success = Math.random() > 0.15; // 85% success rate

            updateNode(id, {
                data: {
                    ...data,
                    status: success ? 'success' : 'error',
                    outputData: success ? {
                        message: 'Executed successfully',
                        timestamp: new Date().toISOString(),
                        data: { result: 'Sample output data' }
                    } : undefined,
                    error: success ? undefined : 'Execution failed: Network timeout',
                    executionTime,
                    lastExecuted: new Date(),
                    performance: {
                        avgTime: executionTime,
                        successRate: success ? 100 : 0,
                        totalRuns: 1
                    }
                }
            });
        }, 1500 + Math.random() * 1000);
    }, [id, data, updateNode]);

    const handleDelete = useCallback((e: React.MouseEvent) => {
        e.stopPropagation();
        deleteNode(id);
    }, [id, deleteNode]);

    const handleDuplicate = useCallback((e: React.MouseEvent) => {
        e.stopPropagation();
        duplicateNode(id);
    }, [id, duplicateNode]);

    return (
        <div
            className="relative group"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
        >
            {/* AI Suggestions Badge */}
            {data.aiSuggestions && data.aiSuggestions.length > 0 && (
                <div className="absolute -top-2 -right-2 z-50">
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <button
                                    className="relative flex items-center justify-center w-6 h-6 rounded-full bg-gradient-to-r from-purple-500 to-pink-500 shadow-lg hover:scale-110 transition-transform"
                                    onClick={() => setShowAI(!showAI)}
                                >
                                    <Sparkles className="w-3 h-3 text-white animate-pulse" />
                                    <span className="absolute -top-1 -right-1 flex h-3 w-3">
                                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-pink-400 opacity-75"></span>
                                        <span className="relative inline-flex rounded-full h-3 w-3 bg-pink-500"></span>
                                    </span>
                                </button>
                            </TooltipTrigger>
                            <TooltipContent>
                                <p className="font-semibold">AI Suggestions Available</p>
                                <p className="text-xs">{data.aiSuggestions.length} optimization tips</p>
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
            )}

            {/* Performance Badge */}
            {data.performance && data.performance.totalRuns > 0 && (
                <div className="absolute -top-2 -left-2 z-50">
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-gradient-to-r from-green-500 to-emerald-600 text-white text-[10px] font-bold shadow-lg">
                                    <TrendingUp className="w-2.5 h-2.5" />
                                    {data.performance.successRate.toFixed(0)}%
                                </div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <p className="font-semibold">Performance Stats</p>
                                <p className="text-xs">Success Rate: {data.performance.successRate.toFixed(1)}%</p>
                                <p className="text-xs">Avg Time: {data.performance.avgTime}ms</p>
                                <p className="text-xs">Total Runs: {data.performance.totalRuns}</p>
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
            )}

            {/* Enhanced Floating Action Toolbar */}
            <div className={`absolute -top-12 left-1/2 -translate-x-1/2 flex items-center gap-1 p-1.5 rounded-xl bg-gradient-to-r from-gray-900 to-gray-800 border border-gray-700 shadow-2xl transition-all duration-300 ${selected || isHovered ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2 pointer-events-none'
                } z-50`}>
                <TooltipProvider>
                    {/* Execute Button */}
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 hover:bg-blue-600 hover:text-white transition-all duration-200"
                                onClick={handleRun}
                                disabled={data.status === 'running'}
                            >
                                {data.status === 'running' ? (
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                ) : (
                                    <Play className="w-4 h-4" />
                                )}
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent><p>Execute Node</p></TooltipContent>
                    </Tooltip>

                    {/* Settings Button */}
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 hover:bg-purple-600 hover:text-white transition-all duration-200"
                            >
                                <Settings className="w-4 h-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent><p>Configure</p></TooltipContent>
                    </Tooltip>

                    {/* Duplicate Button */}
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 hover:bg-green-600 hover:text-white transition-all duration-200"
                                onClick={handleDuplicate}
                            >
                                <Copy className="w-4 h-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent><p>Duplicate</p></TooltipContent>
                    </Tooltip>

                    {/* Delete Button */}
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 hover:bg-red-600 hover:text-white transition-all duration-200"
                                onClick={handleDelete}
                            >
                                <Trash2 className="w-4 h-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent><p>Delete</p></TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            </div>

            {/* Enhanced Main Node Card with Gradient */}
            <Card className={`w-[220px] h-[72px] flex items-center p-0 overflow-hidden bg-gradient-to-br ${categoryColor.gradient} border-0 ${getStatusStyle()}`}>
                {/* Animated Background Pattern */}
                <div className="absolute inset-0 opacity-10">
                    <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGRlZnM+PHBhdHRlcm4gaWQ9ImdyaWQiIHdpZHRoPSI2MCIgaGVpZ2h0PSI2MCIgcGF0dGVyblVuaXRzPSJ1c2VyU3BhY2VPblVzZSI+PHBhdGggZD0iTSAxMCAwIEwgMCAwIDAgMTAiIGZpbGw9Im5vbmUiIHN0cm9rZT0id2hpdGUiIHN0cm9rZS13aWR0aD0iMSIvPjwvcGF0dGVybj48L2RlZnM+PHJlY3Qgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSIgZmlsbD0idXJsKCNncmlkKSIvPjwvc3ZnPg==')]"></div>
                </div>

                {/* Content */}
                <div className="relative flex items-center gap-3 px-4 w-full">
                    {/* Icon with glow effect */}
                    <div className="relative shrink-0">
                        <div className="absolute inset-0 bg-white/20 rounded-lg blur-sm"></div>
                        <div className="relative bg-white/90 backdrop-blur-sm p-2 rounded-lg shadow-lg">
                            <NodeIconBadge
                                type={data.nodeType}
                                category={category}
                                size={20}
                                className="text-gray-800"
                            />
                        </div>
                    </div>

                    {/* Text Content */}
                    <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                            <h3 className="text-sm font-bold text-white truncate drop-shadow-lg">
                                {data.label}
                            </h3>
                            {getStatusIcon()}
                        </div>

                        {/* Execution Time */}
                        {data.executionTime && (
                            <div className="flex items-center gap-1 text-[10px] text-white/80 font-medium">
                                <Zap className="w-2.5 h-2.5" />
                                <span>{data.executionTime}ms</span>
                            </div>
                        )}
                    </div>
                </div>

                {/* Pulse animation for running state */}
                {data.status === 'running' && (
                    <div className="absolute inset-0 bg-blue-400/20 animate-pulse"></div>
                )}
            </Card>

            {/* Input Handles with enhanced styling */}
            {data.inputs.map((input, index) => (
                <Handle
                    key={input.id}
                    type="target"
                    position={Position.Left}
                    id={input.id}
                    className={`!w-3 !h-3 !bg-white !border-[3px] transition-all duration-200 z-10 !left-[-6px] hover:!w-4 hover:!h-4 hover:!border-[4px]`}
                    style={{
                        top: '50%',
                        transform: 'translateY(-50%)',
                        borderColor: categoryColor.base,
                        boxShadow: `0 0 10px ${categoryColor.base}40`
                    }}
                />
            ))}

            {/* Output Handles with enhanced styling */}
            {data.outputs.map((output, index) => (
                <Handle
                    key={output.id}
                    type="source"
                    position={Position.Right}
                    id={output.id}
                    className={`!w-3 !h-3 !bg-white !border-[3px] transition-all duration-200 z-10 !right-[-6px] hover:!w-4 hover:!h-4 hover:!border-[4px]`}
                    style={{
                        top: '50%',
                        transform: 'translateY(-50%)',
                        borderColor: categoryColor.base,
                        boxShadow: `0 0 10px ${categoryColor.base}40`
                    }}
                />
            ))}
        </div>
    );
});

BaseNodeComponent.displayName = 'BaseNode';

export default BaseNodeComponent;
