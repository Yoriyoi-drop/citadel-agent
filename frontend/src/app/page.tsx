"use client"

import { useState } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { MainLayout } from '@/components/layouts/MainLayout';
import {
  Activity,
  PlayCircle,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  Users,
  Zap,
  BarChart3,
  GitBranch,
  Plus,
  Eye,
  Edit,
  Trash2
} from 'lucide-react';

// Mock data
const recentWorkflows = [
  {
    id: '1',
    name: 'Customer Data Processing',
    description: 'Process customer data from CRM to database',
    status: 'active',
    lastRun: '2 hours ago',
    executions: 156,
    successRate: 98.5
  },
  {
    id: '2',
    name: 'Email Campaign Automation',
    description: 'Send automated emails to subscribers',
    status: 'active',
    lastRun: '30 minutes ago',
    executions: 89,
    successRate: 95.2
  },
  {
    id: '3',
    name: 'Report Generation',
    description: 'Generate daily reports from multiple sources',
    status: 'inactive',
    lastRun: '2 days ago',
    executions: 45,
    successRate: 87.3
  }
];

const recentExecutions = [
  {
    id: '1',
    workflowName: 'Customer Data Processing',
    status: 'completed',
    startTime: '2024-01-15T10:30:00Z',
    duration: '2m 34s',
    nodes: 12
  },
  {
    id: '2',
    workflowName: 'Email Campaign Automation',
    status: 'failed',
    startTime: '2024-01-15T10:15:00Z',
    duration: '45s',
    nodes: 8,
    error: 'SMTP connection timeout'
  },
  {
    id: '3',
    workflowName: 'Report Generation',
    status: 'running',
    startTime: '2024-01-15T10:00:00Z',
    duration: '3m 12s',
    nodes: 15
  }
];

const stats = [
  {
    title: 'Total Workflows',
    value: '24',
    change: '+2 from last week',
    icon: GitBranch,
    color: 'text-blue-600'
  },
  {
    title: 'Active Workflows',
    value: '18',
    change: '+3 from yesterday',
    icon: PlayCircle,
    color: 'text-green-600'
  },
  {
    title: 'Total Executions',
    value: '1,247',
    change: '+127 from yesterday',
    icon: Activity,
    color: 'text-purple-600'
  },
  {
    title: 'Success Rate',
    value: '96.2%',
    change: '+1.2% from last week',
    icon: TrendingUp,
    color: 'text-emerald-600'
  }
];

export default function Dashboard() {
  const [selectedTab, setSelectedTab] = useState('overview');

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle className="w-4 h-4" />;
      case 'failed': return <XCircle className="w-4 h-4" />;
      case 'running': return <Clock className="w-4 h-4 animate-spin" />;
      default: return <Clock className="w-4 h-4" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
      case 'completed':
        return <Badge className="bg-primary text-primary-foreground hover:bg-primary border-0 rounded-full">active</Badge>;
      case 'running':
        return <Badge className="bg-secondary text-secondary-foreground hover:bg-secondary border-0 rounded-full">running</Badge>;
      case 'failed':
        return <Badge className="bg-foreground text-background hover:bg-foreground border-0 rounded-full">failed</Badge>;
      default:
        return <Badge variant="secondary" className="rounded-full">inactive</Badge>;
    }
  };

  return (
    <MainLayout>
      <div className="p-4 space-y-6">
        {/* Header Section */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
            <p className="text-muted-foreground">Welcome back! Here's what's happening with your workflows.</p>
          </div>
          <Link href="/workflows/new">
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              New Workflow
            </Button>
          </Link>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {stats.map((stat, index) => (
            <Card key={index} className="bg-muted/30 border-0 shadow-sm hover:shadow-md transition-all">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.title}
                </CardTitle>
                <stat.icon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stat.value}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  {stat.change}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Main Content */}
        <Tabs value={selectedTab} onValueChange={setSelectedTab} className="space-y-6">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="workflows">Recent Workflows</TabsTrigger>
            <TabsTrigger value="executions">Recent Executions</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Recent Workflows */}
              <Card className="border-0 shadow-sm">
                <CardHeader className="flex flex-row items-center justify-between">
                  <div>
                    <CardTitle>Recent Workflows</CardTitle>
                    <CardDescription>Your most recently modified workflows</CardDescription>
                  </div>
                  <Button variant="ghost" size="sm" className="h-8 text-muted-foreground hover:text-foreground">View All</Button>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {recentWorkflows.map((workflow) => (
                      <div
                        key={workflow.id}
                        className="flex items-center justify-between p-4 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors"
                      >
                        <div className="space-y-1">
                          <div className="flex items-center space-x-2">
                            <span className="font-medium">{workflow.name}</span>
                            {getStatusBadge(workflow.status)}
                          </div>
                          <p className="text-sm text-muted-foreground line-clamp-1">
                            {workflow.description}
                          </p>
                          <div className="flex items-center text-xs text-muted-foreground space-x-4">
                            <span>Last run: {workflow.lastRun}</span>
                            <span>{workflow.executions} executions</span>
                            <span>{workflow.successRate}% success</span>
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Edit className="w-4 h-4" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              {/* Recent Executions */}
              <Card className="border-0 shadow-sm">
                <CardHeader className="flex flex-row items-center justify-between">
                  <div>
                    <CardTitle>Recent Executions</CardTitle>
                    <CardDescription>Latest workflow executions</CardDescription>
                  </div>
                  <Button variant="outline" size="sm" className="h-8">View All</Button>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {recentExecutions.map((execution) => (
                      <div
                        key={execution.id}
                        className="flex items-center justify-between p-4 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors"
                      >
                        <div className="space-y-1">
                          <div className="flex items-center space-x-2">
                            {execution.status === 'completed' ? (
                              <CheckCircle className="w-4 h-4" />
                            ) : execution.status === 'failed' ? (
                              <XCircle className="w-4 h-4" />
                            ) : (
                              <PlayCircle className="w-4 h-4" />
                            )}
                            <span className="font-medium">{execution.workflowName}</span>
                            {getStatusBadge(execution.status)}
                          </div>
                          <div className="flex items-center text-xs text-muted-foreground space-x-4">
                            <span>Duration: {execution.duration}</span>
                            <span>{execution.nodes} nodes</span>
                            <span>{new Date(execution.startTime).toLocaleString()}</span>
                          </div>
                          {execution.error && (
                            <p className="text-xs text-muted-foreground mt-1">
                              {execution.error}
                            </p>
                          )}
                        </div>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <Eye className="w-4 h-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="workflows">
            <Card>
              <CardHeader>
                <CardTitle>All Workflows</CardTitle>
                <CardDescription>Manage all your workflows</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {recentWorkflows.map((workflow) => (
                    <div key={workflow.id} className="flex items-center justify-between p-4 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          <h3 className="font-semibold">{workflow.name}</h3>
                          {getStatusBadge(workflow.status)}
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">{workflow.description}</p>
                        <div className="flex items-center space-x-6 mt-3 text-sm">
                          <span>{workflow.executions} executions</span>
                          <span>{workflow.successRate}% success rate</span>
                          <span>Last run: {workflow.lastRun}</span>
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <Button variant="ghost" size="sm">
                          <Eye className="w-4 h-4 mr-2" />
                          View
                        </Button>
                        <Button variant="ghost" size="sm">
                          <Edit className="w-4 h-4 mr-2" />
                          Edit
                        </Button>
                        <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-700 hover:bg-red-50">
                          <Trash2 className="w-4 h-4 mr-2" />
                          Delete
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="executions">
            <Card>
              <CardHeader>
                <CardTitle>Execution History</CardTitle>
                <CardDescription>View all workflow executions</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {recentExecutions.map((execution) => (
                    <div key={execution.id} className="flex items-center justify-between p-4 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          {getStatusIcon(execution.status)}
                          <h3 className="font-semibold">{execution.workflowName}</h3>
                          {getStatusBadge(execution.status)}
                        </div>
                        <div className="flex items-center space-x-6 mt-3 text-sm text-muted-foreground">
                          <span>Started: {new Date(execution.startTime).toLocaleString()}</span>
                          <span>Duration: {execution.duration}</span>
                          <span>{execution.nodes} nodes</span>
                        </div>
                        {execution.error && (
                          <p className="text-sm text-muted-foreground mt-2">{execution.error}</p>
                        )}
                      </div>
                      <Button variant="ghost" size="sm">
                        <Eye className="w-4 h-4 mr-2" />
                        View Details
                      </Button>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="analytics">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>Execution Trends</CardTitle>
                  <CardDescription>Workflow execution trends over time</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="h-[300px] flex items-center justify-center text-muted-foreground">
                    <BarChart3 className="w-12 h-12 mb-2" />
                    <p>Analytics chart would be displayed here</p>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Performance Metrics</CardTitle>
                  <CardDescription>Key performance indicators</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <div className="flex justify-between text-sm mb-2">
                        <span>Average Execution Time</span>
                        <span>2m 15s</span>
                      </div>
                      <Progress value={75} />
                    </div>
                    <div>
                      <div className="flex justify-between text-sm mb-2">
                        <span>Success Rate</span>
                        <span>96.2%</span>
                      </div>
                      <Progress value={96} />
                    </div>
                    <div>
                      <div className="flex justify-between text-sm mb-2">
                        <span>Resource Usage</span>
                        <span>68%</span>
                      </div>
                      <Progress value={68} />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </MainLayout>
  );
}