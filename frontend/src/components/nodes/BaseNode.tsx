"use client"

import { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Play,
  Copy,
  Trash2,
  MoreHorizontal,
  AlertCircle,
  CheckCircle2,
  Loader2
} from 'lucide-react';
import { BaseNode as BaseNodeType } from '@/types/workflow';
import { NodeIcon } from '@/components/NodeIcon';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface BaseNodeComponentProps extends NodeProps {
  data: {
    label: string;
    description?: string;
    nodeType: string;
    inputs: any[];
    outputs: any[];
    config: Record<string, any>;
    status?: 'idle' | 'running' | 'success' | 'error';
    onUpdate: (updates: Partial<BaseNodeType>) => void;
    onDelete: () => void;
    isSelected: boolean;
  };
}

const BaseNodeComponent = memo(({ data, selected }: BaseNodeComponentProps) => {
  const getStatusColor = () => {
    switch (data.status) {
      case 'running': return 'border-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.3)]';
      case 'success': return 'border-green-500 shadow-[0_0_10px_rgba(34,197,94,0.3)]';
      case 'error': return 'border-red-500 shadow-[0_0_10px_rgba(239,68,68,0.3)]';
      default: return 'border-border hover:border-primary/50';
    }
  };

  const getStatusIcon = () => {
    switch (data.status) {
      case 'success': return <CheckCircle2 className="w-3 h-3 text-green-500" />;
      case 'error': return <AlertCircle className="w-3 h-3 text-red-500" />;
      case 'running': return <Loader2 className="w-3 h-3 text-blue-500 animate-spin" />;
      default: return null;
    }
  };

  return (
    <div className="relative group">
      {/* Node Card */}
      <Card
        className={`
          min-w-[180px] max-w-[220px] h-[50px]
          flex items-center px-3 gap-3
          transition-all duration-200
          bg-card/95 backdrop-blur-sm
          border-2
          ${getStatusColor()}
          ${selected ? 'ring-2 ring-primary ring-offset-2 border-primary' : ''}
        `}
      >
        {/* Icon Box */}
        <div className={`
          flex items-center justify-center w-8 h-8 rounded-md
          bg-muted/50 border border-border/50
        `}>
          <NodeIcon type={data.nodeType} size={16} />
        </div>

        {/* Label & Status */}
        <div className="flex-1 min-w-0 flex flex-col justify-center">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm truncate leading-none">
              {data.label}
            </span>
            {getStatusIcon()}
          </div>
          <span className="text-[10px] text-muted-foreground uppercase tracking-wider truncate leading-tight mt-0.5">
            {data.nodeType.replace('-', ' ')}
          </span>
        </div>

        {/* Actions Menu (Visible on Hover/Selected) */}
        <div className={`
          opacity-0 group-hover:opacity-100 transition-opacity
          ${selected ? 'opacity-100' : ''}
        `}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-6 w-6 -mr-1">
                <MoreHorizontal className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={(e) => { e.stopPropagation(); /* Run logic */ }}>
                <Play className="w-4 h-4 mr-2" /> Run Node
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => { e.stopPropagation(); /* Duplicate logic */ }}>
                <Copy className="w-4 h-4 mr-2" /> Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-red-600 focus:text-red-600"
                onClick={(e) => { e.stopPropagation(); data.onDelete(); }}
              >
                <Trash2 className="w-4 h-4 mr-2" /> Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </Card>

      {/* Input Handles - Larger & Better Positioned */}
      {data.inputs.map((input, index) => (
        <Handle
          key={input.id}
          type="target"
          position={Position.Left}
          id={input.id}
          className={`
            !w-4 !h-4 !bg-background !border-2 !border-muted-foreground
            hover:!border-primary hover:!bg-primary/20 transition-colors
            !left-[-9px]
          `}
          style={{ top: '50%', transform: 'translateY(-50%)' }}
        />
      ))}

      {/* Output Handles - Larger & Better Positioned */}
      {data.outputs.map((output, index) => (
        <Handle
          key={output.id}
          type="source"
          position={Position.Right}
          id={output.id}
          className={`
            !w-4 !h-4 !bg-background !border-2 !border-muted-foreground
            hover:!border-primary hover:!bg-primary/20 transition-colors
            !right-[-9px]
          `}
          style={{ top: '50%', transform: 'translateY(-50%)' }}
        />
      ))}
    </div>
  );
});

BaseNodeComponent.displayName = 'BaseNode';

export default BaseNodeComponent;