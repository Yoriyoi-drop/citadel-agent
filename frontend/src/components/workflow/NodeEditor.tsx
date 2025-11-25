"use client"

import { memo, useEffect, useState } from 'react';
import { useWorkflowStore } from '@/stores/workflowStore';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { X } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Play, Loader2 } from 'lucide-react';
import { ConfigField } from './ConfigField';

interface NodeEditorProps {
  nodeId: string;
}

export const NodeEditor = memo(function NodeEditor({ nodeId }: NodeEditorProps) {
  const { currentWorkflow, updateNode, selectNodes } = useWorkflowStore();
  const node = currentWorkflow?.nodes.find(n => n.id === nodeId);

  const [label, setLabel] = useState('');
  const [description, setDescription] = useState('');
  const [config, setConfig] = useState<Record<string, any>>({});
  const [isExecuting, setIsExecuting] = useState(false);
  const [activeTab, setActiveTab] = useState('parameters');
  // New state for error handling
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    if (node) {
      setLabel(node.data.label);
      setDescription(node.data.description || '');
      setConfig(node.data.config || {});
    }
  }, [node]);

  if (!node) return null;

  const handleExecute = async () => {
    setIsExecuting(true);
    setErrorMessage(null);
    updateNode(nodeId, { data: { ...node.data, status: 'running' } });

    // Simulate execution delay with possible failure
    setTimeout(() => {
      // 20% chance to simulate an error
      if (Math.random() < 0.2) {
        const error = 'Execution failed due to network error.';
        updateNode(nodeId, {
          data: {
            ...node.data,
            status: 'error',
            error,
          }
        });
        setErrorMessage(error);
        setIsExecuting(false);
        return;
      }

      const mockOutput = {
        timestamp: new Date().toISOString(),
        nodeId: nodeId,
        status: 'success',
        data: {
          message: "Execution successful",
          result: Math.random() > 0.5 ? "Success data" : { complex: "object", value: 123 }
        }
      };

      updateNode(nodeId, {
        data: {
          ...node.data,
          status: 'success',
          outputData: mockOutput
        }
      });
      setIsExecuting(false);
      setActiveTab('output');
    }, 1500);
  };

  const handleConfigChange = (key: string, value: any) => {
    setConfig(prev => ({
      ...prev,
      [key]: value
    }));
    updateNode(nodeId, {
      data: {
        ...node.data,
        config: {
          ...config,
          [key]: value
        }
      }
    });
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b flex items-center justify-between bg-muted/30">
        <div className="flex items-center gap-2">
          <h3 className="font-semibold text-lg flex items-center gap-2">
            <span className="uppercase text-xs font-bold bg-primary/10 text-primary px-2 py-1 rounded">
              {node.type.replace('-', ' ')}
            </span>
            {label}
          </h3>
        </div>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            className="h-8 gap-1.5 bg-green-600 hover:bg-green-700 text-white"
            onClick={handleExecute}
            disabled={isExecuting}
            aria-label="Execute node"
          >
            {isExecuting ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5 fill-current" />}
            {isExecuting ? 'Running...' : 'Execute Node'}
          </Button>
          {errorMessage && (
            <p className="mt-2 text-sm text-red-500" role="alert">
              {errorMessage}
            </p>
          )}
        </div>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="flex-1 flex flex-col overflow-hidden"
        aria-label="Node editor tabs"
      >
        <div className="px-4 pt-2 border-b bg-background">
          <TabsList className="w-full justify-start h-9 bg-transparent p-0 gap-6" role="tablist">
            <TabsTrigger
              value="parameters"
              className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none px-1 pb-2"
              role="tab"
              aria-selected={activeTab === 'parameters'}
            >
              Parameters
            </TabsTrigger>
            <TabsTrigger
              value="output"
              className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none px-1 pb-2"
              role="tab"
              aria-selected={activeTab === 'output'}
            >
              Output Data
            </TabsTrigger>
          </TabsList>
        </div>

        <ScrollArea className="flex-1">
          <div className="p-6">
            <TabsContent value="parameters" className="mt-0 space-y-4">
              <div className="space-y-2">
                <Label>Label</Label>
                <Input
                  value={label}
                  onChange={(e) => {
                    setLabel(e.target.value);
                    updateNode(nodeId, { data: { ...node.data, label: e.target.value } });
                  }}
                />
              </div>

              <div className="space-y-2">
                <Label>Description</Label>
                <Textarea
                  value={description}
                  onChange={(e) => {
                    setDescription(e.target.value);
                    updateNode(nodeId, { data: { ...node.data, description: e.target.value } });
                  }}
                />
              </div>

              <div className="border-t pt-4 mt-4">
                <h4 className="font-medium mb-3">Configuration</h4>
                <div className="space-y-4">
                  {/* Dynamic config fields based on node type */}
                  {node.type === 'http-request' && (
                    <>
                      <ConfigField
                        label="URL"
                        value={config.url || ''}
                        onChange={(value) => handleConfigChange('url', value)}
                        placeholder="https://api.example.com"
                        required
                      />
                      <ConfigField
                        label="Method"
                        value={config.method || 'GET'}
                        onChange={(value) => handleConfigChange('method', value)}
                      />
                    </>
                  )}

                  {node.type === 'webhook' && (
                    <>
                      <ConfigField
                        label="Path"
                        value={config.path || ''}
                        onChange={(value) => handleConfigChange('path', value)}
                        placeholder="/webhook/path"
                        required
                      />
                      <ConfigField
                        label="Method"
                        value={config.method || 'POST'}
                        onChange={(value) => handleConfigChange('method', value)}
                      />
                    </>
                  )}

                  {node.type === 'database' && (
                    <>
                      <ConfigField
                        label="Table"
                        value={config.table || ''}
                        onChange={(value) => handleConfigChange('table', value)}
                        required
                      />
                      <ConfigField
                        label="Operation"
                        value={config.operation || 'SELECT'}
                        onChange={(value) => handleConfigChange('operation', value)}
                      />
                    </>
                  )}

                  {/* Generic config fallback */}
                  {!['http-request', 'webhook', 'database'].includes(node.type) && (
                    <div className="text-sm text-muted-foreground">
                      No specific configuration available for this node type.
                    </div>
                  )}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="output" className="mt-0 h-full">
              {node.data.outputData ? (
                <div className="bg-muted/50 rounded-md p-3 font-mono text-xs overflow-auto border">
                  <pre>{JSON.stringify(node.data.outputData, null, 2)}</pre>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-40 text-muted-foreground text-sm border-2 border-dashed rounded-md">
                  <Play className="w-8 h-8 mb-2 opacity-20" />
                  <p>No output data available</p>
                  <p className="text-xs mt-1">Run the node to see results</p>
                </div>
              )}
            </TabsContent>
          </div>
        </ScrollArea>
      </Tabs>
    </div>

  );
});