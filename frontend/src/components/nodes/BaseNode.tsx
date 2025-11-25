"use client"

import { memo, useMemo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Play,
  Copy,
  Trash2,
  AlertCircle,
  CheckCircle2,
  Loader2,
  MoreHorizontal
} from 'lucide-react';
import { BaseNode as BaseNodeType } from '@/types/workflow';
import { NodeIcon } from '@/components/NodeIcon';
import { getCategoryColor } from '@/config/nodeIcons';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useWorkflowStore } from '@/stores/workflowStore';

interface BaseNodeComponentProps extends NodeProps {
  data: {
    label: string;
    description?: string;
    nodeType: string;
    inputs: any[];
    outputs: any[];
    config: Record<string, any>;
    status?: 'idle' | 'running' | 'success' | 'error';
    // Callbacks are no longer needed in data as we use the store
  };
}

// Helper to guess category from node type
const getCategoryFromType = (type: string): string => {
  if (type.startsWith('http') || type.startsWith('graphql') || type.startsWith('api') || type.startsWith('webhook')) return 'http';
  if (type.startsWith('postgres') || type.startsWith('mysql') || type.startsWith('mongo') || type.startsWith('redis') || type.includes('db')) return 'database';
  if (type.startsWith('llm') || type.startsWith('openai') || type.startsWith('claude') || type.startsWith('gemini') || type.startsWith('gpt')) return 'ai_llm';
  if (type.includes('image') || type.includes('vision') || type.includes('face')) return 'ai_vision';
  if (type.includes('speech') || type.includes('audio') || type.includes('voice')) return 'ai_speech';
  if (type.includes('text') || type.includes('sentiment') || type.includes('translation')) return 'ai_nlp';
  if (type.startsWith('json') || type.startsWith('xml') || type.startsWith('csv') || type.startsWith('data')) return 'transform';
  if (type.startsWith('if') || type.startsWith('switch') || type.startsWith('validate') || type.startsWith('compare')) return 'validation';
  if (type.startsWith('loop') || type.startsWith('delay') || type.startsWith('wait')) return 'flow';
  if (type.includes('file')) return 'file';
  if (type.includes('s3') || type.includes('gcs') || type.includes('blob')) return 'cloud';
  if (type.includes('email') || type.includes('slack') || type.includes('discord') || type.includes('telegram')) return 'communication';
  if (type.includes('salesforce') || type.includes('hubspot')) return 'crm';
  if (type.includes('stripe') || type.includes('paypal')) return 'payment';
  if (type.includes('cron') || type.includes('schedule')) return 'schedule';
  if (type.includes('encrypt') || type.includes('jwt')) return 'security';
  if (type.includes('log') || type.includes('alert')) return 'monitoring';
  return 'utility';
};

const BaseNodeComponent = memo(({ id, data, selected }: BaseNodeComponentProps) => {
  const { deleteNode, duplicateNode, updateNode } = useWorkflowStore();

  const category = useMemo(() => getCategoryFromType(data.nodeType), [data.nodeType]);
  const categoryColor = useMemo(() => getCategoryColor(category), [category]);

  const getStatusColor = () => {
    switch (data.status) {
      case 'running': return 'ring-2 ring-blue-500 shadow-[0_0_15px_rgba(59,130,246,0.4)]';
      case 'success': return 'ring-2 ring-green-500 shadow-[0_0_15px_rgba(34,197,94,0.4)]';
      case 'error': return 'ring-2 ring-red-500 shadow-[0_0_15px_rgba(239,68,68,0.4)]';
      default: return selected ? `ring-2 ring-primary shadow-lg` : 'hover:shadow-md';
    }
  };

  const getStatusIcon = () => {
    switch (data.status) {
      case 'success': return <CheckCircle2 className="w-3.5 h-3.5 text-green-500" />;
      case 'error': return <AlertCircle className="w-3.5 h-3.5 text-red-500" />;
      case 'running': return <Loader2 className="w-3.5 h-3.5 text-blue-500 animate-spin" />;
      default: return null;
    }
  };

  const handleRun = (e: React.MouseEvent) => {
    e.stopPropagation();
    // Set status to running
    updateNode(id, { data: { ...data, status: 'running' } });

    // Simulate execution (in a real app, this would trigger an API call)
    setTimeout(() => {
      const success = Math.random() > 0.2;
      updateNode(id, {
        data: {
          ...data,
          status: success ? 'success' : 'error',
          outputData: success ? { message: 'Executed successfully' } : undefined,
          error: success ? undefined : 'Execution failed'
        }
      });
    }, 1500);
  };

  return (
    <div className="relative group">
      {/* Floating Action Toolbar (n8n style) */}
      <div className={`
        absolute -top-10 left-1/2 -translate-x-1/2
        flex items-center gap-1 p-1 rounded-lg
        bg-background border border-border shadow-lg
        opacity-0 group-hover:opacity-100 transition-all duration-200
        ${selected ? 'opacity-100 -top-12' : 'translate-y-2 group-hover:translate-y-0'}
        z-50
      `}>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 hover:bg-muted hover:text-primary"
                onClick={handleRun}
              >
                <Play className="w-3.5 h-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent><p>Execute Node</p></TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 hover:bg-muted hover:text-primary"
                onClick={(e) => { e.stopPropagation(); duplicateNode(id); }}
              >
                <Copy className="w-3.5 h-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent><p>Duplicate</p></TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 hover:bg-red-100 hover:text-red-600"
                onClick={(e) => { e.stopPropagation(); deleteNode(id); }}
              >
                <Trash2 className="w-3.5 h-3.5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent><p>Delete</p></TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      {/* Main Node Card */}
      <Card
        className={`
          w-[200px] h-[64px]
          flex items-center p-0 overflow-hidden
          bg-card border-0
          transition-all duration-200
          shadow-sm
          ${getStatusColor()}
        `}
      >
        {/* Left Color Strip / Icon Container */}
        <div
          className="h-full w-[50px] flex items-center justify-center shrink-0 transition-colors"
          style={{ backgroundColor: `${categoryColor}20` }}
        >
          <div
            className="w-8 h-8 rounded-lg flex items-center justify-center"
            style={{ color: categoryColor }}
          >
            <NodeIcon type={data.nodeType} size={20} />
          </div>
        </div>

        {/* Content Section */}
        <div className="flex-1 min-w-0 px-3 py-2 flex flex-col justify-center border-l border-border/50 h-full bg-card/50">
          <div className="flex items-center justify-between gap-2">
            <span className="font-semibold text-sm truncate text-foreground/90">
              {data.label}
            </span>
            {getStatusIcon()}
          </div>
          <span className="text-[10px] text-muted-foreground uppercase tracking-wider truncate mt-0.5 font-medium">
            {data.nodeType.replace(/_/g, ' ')}
          </span>
        </div>
      </Card>

      {/* Input Handles */}
      {data.inputs.map((input, index) => (
        <Handle
          key={input.id}
          type="target"
          position={Position.Left}
          id={input.id}
          className={`
            !w-3 !h-3 !bg-background !border-[2.5px]
            transition-colors z-10
            !left-[-6px]
          `}
          style={{
            top: '50%',
            transform: 'translateY(-50%)',
            borderColor: categoryColor
          }}
        />
      ))}

      {/* Output Handles */}
      {data.outputs.map((output, index) => (
        <Handle
          key={output.id}
          type="source"
          position={Position.Right}
          id={output.id}
          className={`
            !w-3 !h-3 !bg-background !border-[2.5px]
            transition-colors z-10
            !right-[-6px]
          `}
          style={{
            top: '50%',
            transform: 'translateY(-50%)',
            borderColor: categoryColor
          }}
        />
      ))}
    </div>
  );
});

BaseNodeComponent.displayName = 'BaseNode';

export default BaseNodeComponent;