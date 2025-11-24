"use client"

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layouts/MainLayout';
import { WorkflowBuilder } from '@/components/workflow/WorkflowBuilder';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { 
  Save, 
  Play, 
  ArrowLeft,
  GitBranch,
  CheckCircle
} from 'lucide-react';

export default function NewWorkflowPage() {
  const router = useRouter();
  const [workflowName, setWorkflowName] = useState('New Workflow');
  const [workflowDescription, setWorkflowDescription] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [showSettings, setShowSettings] = useState(false);

  const handleSave = async () => {
    setIsSaving(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      console.log('Workflow saved');
    } catch (error) {
      console.error('Failed to save workflow:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleRun = async () => {
    setIsRunning(true);
    try {
      // Simulate workflow execution
      await new Promise(resolve => setTimeout(resolve, 3000));
      console.log('Workflow executed');
    } catch (error) {
      console.error('Failed to run workflow:', error);
    } finally {
      setIsRunning(false);
    }
  };

  const handleBack = () => {
    router.push('/workflows');
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
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
                value={workflowName}
                onChange={(e) => setWorkflowName(e.target.value)}
                className="text-lg font-semibold border-0 bg-transparent p-0 h-auto focus-visible:ring-0"
              />
              <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <span>Just now</span>
                <span>•</span>
                <span>Version 1</span>
                <span>•</span>
                <Badge variant="secondary">Draft</Badge>
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
            disabled={isRunning}
            className="bg-green-600 hover:bg-green-700"
          >
            <Play className="w-4 h-4 mr-2" />
            {isRunning ? 'Running...' : 'Run Workflow'}
          </Button>
        </div>
      </header>

      {/* Settings Panel */}
      {showSettings && (
        <div className="border-b bg-muted/50 p-4">
          <div className="max-w-4xl mx-auto">
            <h3 className="text-lg font-semibold mb-4">Workflow Settings</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={workflowDescription}
                  onChange={(e) => setWorkflowDescription(e.target.value)}
                  rows={3}
                  placeholder="Describe what this workflow does..."
                />
              </div>
              
              <div>
                <Label htmlFor="errorHandling">Error Handling</Label>
                <select
                  id="errorHandling"
                  className="w-full p-2 border rounded-md"
                  defaultValue="stop"
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
                  defaultValue="3"
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
        <WorkflowBuilder />
      </div>

      {/* Status Bar */}
      <div className="border-t bg-muted/50 px-6 py-2">
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center space-x-4">
            <span>Nodes: 0</span>
            <span>Connections: 0</span>
            <span>Status: Draft</span>
          </div>
          
          <div className="flex items-center space-x-4">
            {isRunning && (
              <div className="flex items-center space-x-2">
                <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
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