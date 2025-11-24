"use client"

import { useState, useEffect } from 'react';
import { useWorkflowStore } from '@/stores/workflowStore';
import { useNodeStore } from '@/stores/nodeStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  Settings, 
  Trash2, 
  Copy, 
  Play, 
  Pause,
  CheckCircle,
  XCircle,
  Clock,
  Zap
} from 'lucide-react';

interface NodeEditorProps {
  nodeId: string;
}

export function NodeEditor({ nodeId }: NodeEditorProps) {
  const { currentWorkflow, updateNode } = useWorkflowStore();
  const { nodeTypes } = useNodeStore();
  const [configValues, setConfigValues] = useState<Record<string, any>>({});

  const node = currentWorkflow?.nodes.find(n => n.id === nodeId);
  const nodeType = node ? nodeTypes.find(nt => nt.id === node.type) : null;

  useEffect(() => {
    if (node) {
      setConfigValues(node.data.config || {});
    }
  }, [node]);

  if (!node || !nodeType) {
    return (
      <div className="p-4">
        <div className="text-center text-muted-foreground">
          <Settings className="w-8 h-8 mx-auto mb-2" />
          <p>Select a node to edit</p>
        </div>
      </div>
    );
  }

  const handleConfigChange = (name: string, value: any) => {
    const newConfig = { ...configValues, [name]: value };
    setConfigValues(newConfig);
    updateNode(nodeId, {
      data: {
        ...node.data,
        config: newConfig
      }
    });
  };

  const getStatusIcon = () => {
    switch (node.data.status) {
      case 'success': return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'error': return <XCircle className="w-4 h-4 text-red-500" />;
      case 'running': return <Clock className="w-4 h-4 text-blue-500 animate-spin" />;
      default: return <Zap className="w-4 h-4 text-gray-500" />;
    }
  };

  const getStatusText = () => {
    switch (node.data.status) {
      case 'success': return 'Success';
      case 'error': return 'Error';
      case 'running': return 'Running';
      default: return 'Idle';
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-lg font-semibold">{node.data.label}</h3>
          <div className="flex items-center space-x-2">
            {getStatusIcon()}
            <span className="text-sm text-muted-foreground">{getStatusText()}</span>
          </div>
        </div>
        <p className="text-sm text-muted-foreground">{nodeType.description}</p>
        <Badge variant="outline" className="mt-2">
          {nodeType.name}
        </Badge>
      </div>

      {/* Configuration */}
      <ScrollArea className="flex-1">
        <div className="p-4 space-y-6">
          {/* Basic Settings */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Basic Settings</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="nodeLabel">Node Label</Label>
                <Input
                  id="nodeLabel"
                  value={node.data.label}
                  onChange={(e) => updateNode(nodeId, {
                    data: { ...node.data, label: e.target.value }
                  })}
                />
              </div>
              
              <div>
                <Label htmlFor="nodeDescription">Description</Label>
                <Textarea
                  id="nodeDescription"
                  value={node.data.description || ''}
                  onChange={(e) => updateNode(nodeId, {
                    data: { ...node.data, description: e.target.value }
                  })}
                  rows={3}
                />
              </div>
            </CardContent>
          </Card>

          {/* Node Configuration */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Configuration</CardTitle>
              <CardDescription>
                Configure the node's behavior and parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {nodeType.config.map((config) => (
                <div key={config.name} className="space-y-2">
                  <Label htmlFor={config.name}>
                    {config.label}
                    {config.required && <span className="text-red-500 ml-1">*</span>}
                  </Label>
                  
                  {config.type === 'string' && (
                    <Input
                      id={config.name}
                      value={configValues[config.name] || config.default || ''}
                      onChange={(e) => handleConfigChange(config.name, e.target.value)}
                      placeholder={config.description}
                    />
                  )}
                  
                  {config.type === 'textarea' && (
                    <Textarea
                      id={config.name}
                      value={configValues[config.name] || config.default || ''}
                      onChange={(e) => handleConfigChange(config.name, e.target.value)}
                      placeholder={config.description}
                      rows={4}
                    />
                  )}
                  
                  {config.type === 'number' && (
                    <Input
                      id={config.name}
                      type="number"
                      value={configValues[config.name] || config.default || ''}
                      onChange={(e) => handleConfigChange(config.name, Number(e.target.value))}
                      min={config.validation?.min}
                      max={config.validation?.max}
                    />
                  )}
                  
                  {config.type === 'boolean' && (
                    <div className="flex items-center space-x-2">
                      <Switch
                        id={config.name}
                        checked={configValues[config.name] || config.default || false}
                        onCheckedChange={(checked) => handleConfigChange(config.name, checked)}
                      />
                      <Label htmlFor={config.name}>{config.description}</Label>
                    </div>
                  )}
                  
                  {config.type === 'select' && (
                    <Select
                      value={configValues[config.name] || config.default || ''}
                      onValueChange={(value) => handleConfigChange(config.name, value)}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={`Select ${config.label}`} />
                      </SelectTrigger>
                      <SelectContent>
                        {config.options?.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                  
                  {config.type === 'password' && (
                    <Input
                      id={config.name}
                      type="password"
                      value={configValues[config.name] || ''}
                      onChange={(e) => handleConfigChange(config.name, e.target.value)}
                      placeholder={config.description}
                    />
                  )}
                  
                  {config.description && (
                    <p className="text-xs text-muted-foreground">{config.description}</p>
                  )}
                </div>
              ))}
            </CardContent>
          </Card>

          {/* Input/Output Ports */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Ports</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <h4 className="text-sm font-medium mb-2">Inputs</h4>
                <div className="space-y-2">
                  {nodeType.inputs.map((input) => (
                    <div key={input.id} className="flex items-center justify-between p-2 bg-muted rounded">
                      <span className="text-sm">{input.name}</span>
                      <div className="flex items-center space-x-2">
                        <Badge variant="outline" className="text-xs">
                          {input.type}
                        </Badge>
                        {input.required && (
                          <Badge variant="destructive" className="text-xs">
                            Required
                          </Badge>
                        )}
                      </div>
                    </div>
                  ))}
                  {nodeType.inputs.length === 0 && (
                    <p className="text-sm text-muted-foreground">No inputs</p>
                  )}
                </div>
              </div>
              
              <Separator />
              
              <div>
                <h4 className="text-sm font-medium mb-2">Outputs</h4>
                <div className="space-y-2">
                  {nodeType.outputs.map((output) => (
                    <div key={output.id} className="flex items-center justify-between p-2 bg-muted rounded">
                      <span className="text-sm">{output.name}</span>
                      <Badge variant="outline" className="text-xs">
                        {output.type}
                      </Badge>
                    </div>
                  ))}
                  {nodeType.outputs.length === 0 && (
                    <p className="text-sm text-muted-foreground">No outputs</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </ScrollArea>

      {/* Actions */}
      <div className="p-4 border-t">
        <div className="flex space-x-2">
          <Button variant="outline" size="sm">
            <Play className="w-4 h-4 mr-2" />
            Test Node
          </Button>
          <Button variant="outline" size="sm">
            <Copy className="w-4 h-4 mr-2" />
            Duplicate
          </Button>
          <Button variant="destructive" size="sm">
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>
    </div>
  );
}