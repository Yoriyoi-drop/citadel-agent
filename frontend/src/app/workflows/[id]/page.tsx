"use client"

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layouts/MainLayout';
import { Header } from '@/components/layouts/Header';
import { WorkflowBuilder } from '@/components/workflow/WorkflowBuilder';
import { useWorkflowStore } from '@/stores/workflowStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { 
  Save, 
  Play, 
  Settings, 
  ArrowLeft,
  GitBranch,
  Clock,
  CheckCircle,
  XCircle
} from 'lucide-react';

export default function WorkflowEditor() {
  const params = useParams();
  const router = useRouter();
  const { currentWorkflow, setCurrentWorkflow, updateWorkflow } = useWorkflowStore();
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [showSettings, setShowSettings] = useState(false);

  const workflowId = params.id as string;

  useEffect(() => {
    // Load workflow data based on ID
    if (workflowId) {
      // Mock data - in real app, this would be an API call
      const mockWorkflow = {
        id: workflowId,
        name: 'Customer Data Processing',
        description: 'Process customer data from CRM to database',
        nodes: [],
        edges: [],
        settings: {
          autoSave: true,
          errorHandling: 'stop' as const,
          retryCount: 3
        },
        createdAt: new Date('2024-01-10'),
        updatedAt: new Date('2024-01-15'),
        version: 1,
        isActive: true
      };
      
      setCurrentWorkflow(mockWorkflow);
    }
  }, [workflowId, setCurrentWorkflow]);

  const handleSave = async () => {
    if (!currentWorkflow) return;
    
    setIsSaving(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      updateWorkflow(currentWorkflow.id, {
        updatedAt: new Date()
      });
    } catch (error) {
      console.error('Failed to save workflow:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleRun = async () => {
    if (!currentWorkflow) return;
    
    setIsRunning(true);
    try {
      // Simulate workflow execution
      await new Promise(resolve => setTimeout(resolve, 3000));
      console.log('Workflow executed successfully');
    } catch (error) {
      console.error('Failed to run workflow:', error);
    } finally {
      setIsRunning(false);
    }
  };

  const handleBack = () => {
    router.push('/workflows');
  };

  if (!currentWorkflow) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-semibold mb-2">Workflow not found</h2>
            <p className="text-muted-foreground mb-4">The workflow you're looking for doesn't exist.</p>
            <Button onClick={handleBack}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Workflows
            </Button>
          </div>
        </div>
      </MainLayout>
    );
  }

  return (
    <div className="flex flex-col h-screen">
      {/* Custom Header for Workflow Editor */}
      <header className="flex items-center justify-between px-6 py-3 border-b bg-background/95 backdrop-blur">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={handleBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          
          <div className="flex items-center space-x-3">
            <GitBranch className="w-5 h-5 text-primary" />
            <div>
              <Input
                value={currentWorkflow.name}
                onChange={(e) => updateWorkflow(currentWorkflow.id, { name: e.target.value })}
                className="text-lg font-semibold border-0 bg-transparent p-0 h-auto focus-visible:ring-0"
              />
              <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <span>Last updated: {currentWorkflow.updatedAt.toLocaleDateString()}</span>
                <span>•</span>
                <span>Version {currentWorkflow.version}</span>
                <span>•</span>
                <Badge variant={currentWorkflow.isActive ? 'default' : 'secondary'}>
                  {currentWorkflow.isActive ? 'Active' : 'Inactive'}
                </Badge>
              </div>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowSettings(!showSettings)}
          >
            <Settings className="w-4 h-4 mr-2" />
            Settings
          </Button>
          
          <Button
            variant="outline"
            size="sm"
            onClick={handleSave}
            disabled={isSaving}
          >
            <Save className="w-4 h-4 mr-2" />
            {isSaving ? 'Saving...' : 'Save'}
          </Button>
          
          <Button
            size="sm"
            onClick={handleRun}
            disabled={isRunning || currentWorkflow.nodes.length === 0}
            className="bg-green-600 hover:bg-green-700"
          >
            <Play className="w-4 h-4 mr-2" />
            {isRunning ? 'Running...' : 'Run Workflow'}
          </Button>
        </div>
      </header>

      {/* Workflow Settings Panel */}
      {showSettings && (
        <div className="border-b bg-muted/50 p-4">
          <div className="max-w-4xl mx-auto">
            <h3 className="text-lg font-semibold mb-4">Workflow Settings</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={currentWorkflow.description || ''}
                  onChange={(e) => updateWorkflow(currentWorkflow.id, { description: e.target.value })}
                  rows={3}
                />
              </div>
              
              <div>
                <Label htmlFor="errorHandling">Error Handling</Label>
                <select
                  id="errorHandling"
                  value={currentWorkflow.settings.errorHandling}
                  onChange={(e) => updateWorkflow(currentWorkflow.id, {
                    settings: {
                      ...currentWorkflow.settings,
                      errorHandling: e.target.value as 'stop' | 'continue' | 'retry'
                    }
                  })}
                  className="w-full p-2 border rounded-md"
                >
                  <option value="stop">Stop on Error</option>
                  <option value="continue">Continue on Error</option>
                  <option value="retry">Retry on Error</option>
                </select>
              </div>
              
              <div>
                <Label htmlFor="retryCount">Retry Count</Label>
                <Input
                  id="retryCount"
                  type="number"
                  value={currentWorkflow.settings.retryCount}
                  onChange={(e) => updateWorkflow(currentWorkflow.id, {
                    settings: {
                      ...currentWorkflow.settings,
                      retryCount: parseInt(e.target.value) || 0
                    }
                  })}
                  min="0"
                  max="10"
                />
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Workflow Builder */}
      <div className="flex-1 overflow-hidden">
        <WorkflowBuilder workflowId={workflowId} />
      </div>

      {/* Status Bar */}
      <div className="border-t bg-muted/50 px-6 py-2">
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center space-x-4">
            <span>Nodes: {currentWorkflow.nodes.length}</span>
            <span>Connections: {currentWorkflow.edges.length}</span>
            <span>Status: {currentWorkflow.isActive ? 'Active' : 'Inactive'}</span>
          </div>
          
          <div className="flex items-center space-x-4">
            {isRunning && (
              <div className="flex items-center space-x-2">
                <Clock className="w-4 h-4 animate-spin" />
                <span>Running workflow...</span>
              </div>
            )}
            
            {isSaving && (
              <div className="flex items-center space-x-2">
                <Save className="w-4 h-4" />
                <span>Saving...</span>
              </div>
            )}
            
            <div className="flex items-center space-x-2">
              <CheckCircle className="w-4 h-4 text-green-500" />
              <span>All changes saved</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}