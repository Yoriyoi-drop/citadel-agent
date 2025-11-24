"use client"

import { memo } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Settings,
  Play,
  Copy,
  Trash2,
  CheckCircle,
  XCircle,
  Clock,
  Zap
} from 'lucide-react';
import { BaseNode as BaseNodeType } from '@/types/workflow';
import { NodeIcon } from '@/components/NodeIcon';

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
      case 'running': return 'border-blue-500 bg-blue-50';
      case 'success': return 'border-green-500 bg-green-50';
      case 'error': return 'border-red-500 bg-red-50';
      default: return 'border-gray-300 bg-white';
    }
  };

  const getStatusIcon = () => {
    switch (data.status) {
      case 'success': return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'error': return <XCircle className="w-4 h-4 text-red-500" />;
      case 'running': return <Clock className="w-4 h-4 text-blue-500 animate-spin" />;
      default: return <Zap className="w-4 h-4 text-gray-500" />;
    }
  };


  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    data.onDelete();
  };

  const handleDuplicate = (e: React.MouseEvent) => {
    e.stopPropagation();
    // Handle duplication logic
  };

  const handleRun = (e: React.MouseEvent) => {
    e.stopPropagation();
    // Handle run logic
  };

  return (
    <Card
      className={`min-w-[200px] transition-all duration-200 ${getStatusColor()} ${selected ? 'ring-2 ring-primary ring-offset-2' : ''
        }`}
    >
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <NodeIcon type={data.nodeType} size={18} />
            <h3 className="font-semibold text-sm">{data.label}</h3>
          </div>
          <div className="flex items-center space-x-1">
            {getStatusIcon()}
            <div className="opacity-0 hover:opacity-100 transition-opacity flex space-x-1">
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0"
                onClick={handleRun}
              >
                <Play className="w-3 h-3" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0"
                onClick={handleDuplicate}
              >
                <Copy className="w-3 h-3" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0 text-red-500 hover:text-red-700"
                onClick={handleDelete}
              >
                <Trash2 className="w-3 h-3" />
              </Button>
            </div>
          </div>
        </div>
        {data.description && (
          <p className="text-xs text-muted-foreground">{data.description}</p>
        )}
      </CardHeader>

      <CardContent className="pt-0">
        {/* Input Handles */}
        {data.inputs.map((input, index) => (
          <Handle
            key={`input-${index}`}
            type="target"
            position={Position.Left}
            id={input.id}
            style={{ top: `${30 + index * 20}px` }}
            className="w-3 h-3 bg-gray-400 border-2 border-white"
          />
        ))}

        {/* Output Handles */}
        {data.outputs.map((output, index) => (
          <Handle
            key={`output-${index}`}
            type="source"
            position={Position.Right}
            id={output.id}
            style={{ top: `${30 + index * 20}px` }}
            className="w-3 h-3 bg-blue-500 border-2 border-white"
          />
        ))}

        {/* Port Labels */}
        <div className="flex justify-between text-xs">
          <div className="space-y-1">
            {data.inputs.slice(0, 2).map((input, index) => (
              <div key={index} className="text-muted-foreground">
                {input.name}
              </div>
            ))}
            {data.inputs.length > 2 && (
              <div className="text-muted-foreground">+{data.inputs.length - 2}</div>
            )}
          </div>

          <div className="space-y-1 text-right">
            {data.outputs.slice(0, 2).map((output, index) => (
              <div key={index} className="text-muted-foreground">
                {output.name}
              </div>
            ))}
            {data.outputs.length > 2 && (
              <div className="text-muted-foreground">+{data.outputs.length - 2}</div>
            )}
          </div>
        </div>

        {/* Configuration Summary */}
        {Object.keys(data.config).length > 0 && (
          <div className="mt-3 pt-2 border-t">
            <div className="flex flex-wrap gap-1">
              {Object.entries(data.config).slice(0, 3).map(([key, value]) => (
                <Badge key={key} variant="secondary" className="text-xs">
                  {key}: {String(value).length > 10 ? `${String(value).slice(0, 10)}...` : value}
                </Badge>
              ))}
              {Object.keys(data.config).length > 3 && (
                <Badge variant="outline" className="text-xs">
                  +{Object.keys(data.config).length - 3}
                </Badge>
              )}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
});

BaseNodeComponent.displayName = 'BaseNodeComponent';

export default BaseNodeComponent;