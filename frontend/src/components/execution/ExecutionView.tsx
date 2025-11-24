"use client"

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Play, 
  Pause, 
  Square, 
  RotateCcw,
  CheckCircle,
  XCircle,
  Clock,
  Zap,
  Terminal,
  FileText,
  Settings
} from 'lucide-react';

interface ExecutionViewProps {
  executionId: string;
}

export function ExecutionView({ executionId }: ExecutionViewProps) {
  const [isRunning, setIsRunning] = useState(false);
  const [selectedTab, setSelectedTab] = useState('overview');

  // Mock execution data
  const execution = {
    id: executionId,
    status: 'running',
    startTime: new Date('2024-01-15T10:30:00Z'),
    duration: '2m 34s',
    nodes: [
      {
        id: '1',
        name: 'HTTP Request',
        status: 'completed',
        duration: '1.2s',
        startTime: '10:30:00',
        endTime: '10:30:01'
      },
      {
        id: '2',
        name: 'Data Transform',
        status: 'running',
        duration: '0.8s',
        startTime: '10:30:01',
        endTime: null
      },
      {
        id: '3',
        name: 'Database Query',
        status: 'pending',
        duration: null,
        startTime: null,
        endTime: null
      }
    ],
    logs: [
      {
        id: '1',
        level: 'info',
        message: 'Starting workflow execution',
        timestamp: '10:30:00',
        nodeId: null
      },
      {
        id: '2',
        level: 'info',
        message: 'HTTP Request: GET https://api.example.com/data',
        timestamp: '10:30:00',
        nodeId: '1'
      },
      {
        id: '3',
        level: 'success',
        message: 'HTTP Request completed successfully (200)',
        timestamp: '10:30:01',
        nodeId: '1'
      },
      {
        id: '4',
        level: 'info',
        message: 'Starting data transformation',
        timestamp: '10:30:01',
        nodeId: '2'
      }
    ]
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'failed': return <XCircle className="w-4 h-4 text-red-500" />;
      case 'running': return <Clock className="w-4 h-4 text-blue-500 animate-spin" />;
      case 'pending': return <Clock className="w-4 h-4 text-gray-500" />;
      default: return <Clock className="w-4 h-4 text-gray-500" />;
    }
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
      completed: 'default',
      failed: 'destructive',
      running: 'outline',
      pending: 'secondary'
    };
    
    return (
      <Badge variant={variants[status] || 'outline'}>
        {status}
      </Badge>
    );
  };

  const getLogLevelColor = (level: string) => {
    switch (level) {
      case 'error': return 'text-red-500';
      case 'warn': return 'text-yellow-500';
      case 'success': return 'text-green-500';
      case 'info': return 'text-blue-500';
      default: return 'text-gray-500';
    }
  };

  const handleStop = () => {
    setIsRunning(false);
  };

  const handlePause = () => {
    setIsRunning(!isRunning);
  };

  const handleRestart = () => {
    setIsRunning(true);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b">
        <div>
          <h2 className="text-lg font-semibold">Execution #{execution.id}</h2>
          <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            <span>Started: {execution.startTime.toLocaleString()}</span>
            <span>•</span>
            <span>Duration: {execution.duration}</span>
            <span>•</span>
            {getStatusBadge(execution.status)}
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          {isRunning ? (
            <>
              <Button variant="outline" size="sm" onClick={handlePause}>
                <Pause className="w-4 h-4 mr-2" />
                Pause
              </Button>
              <Button variant="destructive" size="sm" onClick={handleStop}>
                <Square className="w-4 h-4 mr-2" />
                Stop
              </Button>
            </>
          ) : (
            <Button variant="outline" size="sm" onClick={handleRestart}>
              <RotateCcw className="w-4 h-4 mr-2" />
              Restart
            </Button>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        <Tabs value={selectedTab} onValueChange={setSelectedTab} className="h-full flex flex-col">
          <div className="border-b px-4">
            <TabsList>
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="nodes">Node Execution</TabsTrigger>
              <TabsTrigger value="logs">Logs</TabsTrigger>
              <TabsTrigger value="debug">Debug</TabsTrigger>
            </TabsList>
          </div>

          <div className="flex-1 overflow-hidden">
            <TabsContent value="overview" className="h-full m-0">
              <div className="p-4 h-full">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 h-full">
                  {/* Execution Progress */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">Execution Progress</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div>
                          <div className="flex justify-between text-sm mb-2">
                            <span>Overall Progress</span>
                            <span>33%</span>
                          </div>
                          <Progress value={33} />
                        </div>
                        
                        <div className="space-y-2">
                          <div className="flex justify-between text-sm">
                            <span>Completed Nodes</span>
                            <span>1 / 3</span>
                          </div>
                          <div className="flex justify-between text-sm">
                            <span>Running Nodes</span>
                            <span>1</span>
                          </div>
                          <div className="flex justify-between text-sm">
                            <span>Pending Nodes</span>
                            <span>1</span>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Performance Metrics */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">Performance Metrics</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div>
                          <div className="flex justify-between text-sm mb-2">
                            <span>CPU Usage</span>
                            <span>45%</span>
                          </div>
                          <Progress value={45} />
                        </div>
                        
                        <div>
                          <div className="flex justify-between text-sm mb-2">
                            <span>Memory Usage</span>
                            <span>2.1GB</span>
                          </div>
                          <Progress value={65} />
                        </div>
                        
                        <div>
                          <div className="flex justify-between text-sm mb-2">
                            <span>Network I/O</span>
                            <span>23%</span>
                          </div>
                          <Progress value={23} />
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="nodes" className="h-full m-0">
              <div className="p-4 h-full">
                <ScrollArea className="h-full">
                  <div className="space-y-3">
                    {execution.nodes.map((node) => (
                      <Card key={node.id}>
                        <CardContent className="p-4">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-3">
                              {getStatusIcon(node.status)}
                              <div>
                                <h4 className="font-medium">{node.name}</h4>
                                <div className="text-sm text-muted-foreground">
                                  {node.startTime && (
                                    <span>Started: {node.startTime}</span>
                                  )}
                                  {node.endTime && (
                                    <span> • Ended: {node.endTime}</span>
                                  )}
                                  {node.duration && (
                                    <span> • Duration: {node.duration}</span>
                                  )}
                                </div>
                              </div>
                            </div>
                            {getStatusBadge(node.status)}
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                </ScrollArea>
              </div>
            </TabsContent>

            <TabsContent value="logs" className="h-full m-0">
              <div className="p-4 h-full">
                <ScrollArea className="h-full">
                  <div className="space-y-2 font-mono text-sm">
                    {execution.logs.map((log) => (
                      <div key={log.id} className="flex items-start space-x-2 p-2 rounded hover:bg-muted">
                        <span className="text-muted-foreground text-xs w-20">
                          {log.timestamp}
                        </span>
                        <span className={`font-medium ${getLogLevelColor(log.level)} w-16`}>
                          {log.level.toUpperCase()}
                        </span>
                        <span className="flex-1">{log.message}</span>
                        {log.nodeId && (
                          <Badge variant="outline" className="text-xs">
                            Node {log.nodeId}
                          </Badge>
                        )}
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              </div>
            </TabsContent>

            <TabsContent value="debug" className="h-full m-0">
              <div className="p-4 h-full">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 h-full">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center">
                        <Terminal className="w-4 h-4 mr-2" />
                        Console Output
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <ScrollArea className="h-[300px]">
                        <div className="font-mono text-xs space-y-1">
                          <div className="text-green-500">$ Starting workflow execution...</div>
                          <div className="text-blue-500">$ Initializing node: HTTP Request</div>
                          <div className="text-blue-500">$ Making request to: https://api.example.com/data</div>
                          <div className="text-green-500">$ Response received: 200 OK</div>
                          <div className="text-blue-500">$ Initializing node: Data Transform</div>
                          <div className="text-yellow-500">$ Processing data...</div>
                        </div>
                      </ScrollArea>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center">
                        <FileText className="w-4 h-4 mr-2" />
                        Variable Inspector
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <ScrollArea className="h-[300px]">
                        <div className="space-y-3">
                          <div>
                            <h5 className="font-medium text-sm">Global Variables</h5>
                            <div className="mt-2 space-y-1">
                              <div className="flex justify-between text-sm">
                                <span>workflow_id</span>
                                <span className="font-mono">"{execution.id}"</span>
                              </div>
                              <div className="flex justify-between text-sm">
                                <span>execution_time</span>
                                <span className="font-mono">"2m 34s"</span>
                              </div>
                            </div>
                          </div>
                          
                          <div>
                            <h5 className="font-medium text-sm">Node 1 Output</h5>
                            <div className="mt-2 space-y-1">
                              <div className="flex justify-between text-sm">
                                <span>status_code</span>
                                <span className="font-mono">200</span>
                              </div>
                              <div className="flex justify-between text-sm">
                                <span>response_size</span>
                                <span className="font-mono">1024</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </ScrollArea>
                    </CardContent>
                  </Card>
                </div>
              </div>
            </TabsContent>
          </div>
        </Tabs>
      </div>
    </div>
  );
}