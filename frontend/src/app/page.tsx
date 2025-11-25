"use client"

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useWorkflowStore } from '@/stores/workflowStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Input } from '@/components/ui/input';
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
  Trash2,
  Search,
  Filter,
  ArrowUpRight,
  Loader2,
  AlertCircle
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
    trend: '+8.3%',
    icon: GitBranch,
    color: 'text-blue-600',
    bgColor: 'bg-blue-50 dark:bg-blue-950',
    iconColor: 'text-blue-600'
  },
  {
    title: 'Active Workflows',
    value: '18',
    change: '+3 from yesterday',
    trend: '+16.7%',
    icon: PlayCircle,
    color: 'text-green-600',
    bgColor: 'bg-green-50 dark:bg-green-950',
    iconColor: 'text-green-600'
  },
  {
    title: 'Total Executions',
    value: '1,247',
    change: '+127 from yesterday',
    trend: '+11.3%',
    icon: Activity,
    color: 'text-purple-600',
    bgColor: 'bg-purple-50 dark:bg-purple-950',
    iconColor: 'text-purple-600'
  },
  {
    title: 'Success Rate',
    value: '96.2%',
    change: '+1.2% from last week',
    trend: '+1.2%',
    icon: TrendingUp,
    color: 'text-emerald-600',
    bgColor: 'bg-emerald-50 dark:bg-emerald-950',
    iconColor: 'text-emerald-600',
    showProgress: true,
    progressValue: 96.2
  }
];

// Mock workflow data with nodes and edges (keeping existing data)
const mockWorkflowData: Record<string, { nodes: any[], edges: any[] }> = {
  '1': {
    nodes: [
      {
        id: 'trigger-1',
        type: 'webhook',
        position: { x: 100, y: 100 },
        data: { label: 'Webhook Trigger', inputs: [], outputs: [{ id: 'out-1', name: 'output', type: 'object' }], config: { method: 'POST', path: '/customer-data' } }
      },
      {
        id: 'transform-1',
        type: 'transform',
        position: { x: 400, y: 100 },
        data: { label: 'Format Data', inputs: [{ id: 'in-1', name: 'input', type: 'object' }], outputs: [{ id: 'out-1', name: 'output', type: 'object' }], config: { script: 'return { ...input, processed: true };' } }
      },
      {
        id: 'db-1',
        type: 'database',
        position: { x: 700, y: 100 },
        data: { label: 'Save to DB', inputs: [{ id: 'in-1', name: 'input', type: 'object' }], outputs: [], config: { table: 'customers', operation: 'insert' } }
      }
    ],
    edges: [
      { id: 'e1-2', source: 'trigger-1', target: 'transform-1', sourceHandle: 'out-1', targetHandle: 'in-1' },
      { id: 'e2-3', source: 'transform-1', target: 'db-1', sourceHandle: 'out-1', targetHandle: 'in-1' }
    ]
  },
  '2': {
    nodes: [
      {
        id: 'schedule-1',
        type: 'schedule',
        position: { x: 100, y: 100 },
        data: { label: 'Daily Schedule', inputs: [], outputs: [{ id: 'out-1', name: 'output', type: 'object' }], config: { cron: '0 9 * * *' } }
      },
      {
        id: 'db-read-1',
        type: 'database',
        position: { x: 400, y: 100 },
        data: { label: 'Get Subscribers', inputs: [{ id: 'in-1', name: 'input', type: 'object' }], outputs: [{ id: 'out-1', name: 'output', type: 'array' }], config: { query: 'SELECT * FROM subscribers WHERE active = true' } }
      },
      {
        id: 'email-1',
        type: 'email',
        position: { x: 700, y: 100 },
        data: { label: 'Send Email', inputs: [{ id: 'in-1', name: 'recipients', type: 'array' }], outputs: [], config: { subject: 'Daily Update', template: 'newsletter' } }
      }
    ],
    edges: [
      { id: 'e1-2', source: 'schedule-1', target: 'db-read-1', sourceHandle: 'out-1', targetHandle: 'in-1' },
      { id: 'e2-3', source: 'db-read-1', target: 'email-1', sourceHandle: 'out-1', targetHandle: 'in-1' }
    ]
  },
  '3': {
    nodes: [
      {
        id: 'trigger-1',
        type: 'webhook',
        position: { x: 100, y: 100 },
        data: { label: 'Generate Request', inputs: [], outputs: [{ id: 'out-1', name: 'output', type: 'object' }], config: { method: 'POST', path: '/generate-report' } }
      },
      {
        id: 'api-1',
        type: 'http-request',
        position: { x: 400, y: 50 },
        data: { label: 'Fetch Analytics', inputs: [{ id: 'in-1', name: 'input', type: 'object' }], outputs: [{ id: 'out-1', name: 'output', type: 'object' }], config: { url: 'https://api.analytics.com/data', method: 'GET' } }
      },
      {
        id: 'api-2',
        type: 'http-request',
        position: { x: 400, y: 200 },
        data: { label: 'Fetch Sales', inputs: [{ id: 'in-1', name: 'input', type: 'object' }], outputs: [{ id: 'out-1', name: 'output', type: 'object' }], config: { url: 'https://api.sales.com/data', method: 'GET' } }
      },
      {
        id: 'transform-1',
        type: 'transform',
        position: { x: 700, y: 125 },
        data: { label: 'Merge Data', inputs: [{ id: 'in-1', name: 'analytics', type: 'object' }, { id: 'in-2', name: 'sales', type: 'object' }], outputs: [{ id: 'out-1', name: 'report', type: 'object' }], config: { script: 'return { ...analytics, ...sales };' } }
      },
      {
        id: 'file-1',
        type: 'file-save',
        position: { x: 1000, y: 125 },
        data: { label: 'Save Report', inputs: [{ id: 'in-1', name: 'content', type: 'object' }], outputs: [], config: { filename: 'report.pdf', format: 'pdf' } }
      }
    ],
    edges: [
      { id: 'e1-2', source: 'trigger-1', target: 'api-1', sourceHandle: 'out-1', targetHandle: 'in-1' },
      { id: 'e1-3', source: 'trigger-1', target: 'api-2', sourceHandle: 'out-1', targetHandle: 'in-1' },
      { id: 'e2-4', source: 'api-1', target: 'transform-1', sourceHandle: 'out-1', targetHandle: 'in-1' },
      { id: 'e3-4', source: 'api-2', target: 'transform-1', sourceHandle: 'out-1', targetHandle: 'in-2' },
      { id: 'e4-5', source: 'transform-1', target: 'file-1', sourceHandle: 'out-1', targetHandle: 'in-1' }
    ]
  }
};

export default function Dashboard() {
  const [selectedTab, setSelectedTab] = useState('overview');
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const router = useRouter();
  const { setCurrentWorkflow } = useWorkflowStore();

  const handleEditWorkflow = (workflow: any) => {
    const mockData = mockWorkflowData[workflow.id] || { nodes: [], edges: [] };

    setCurrentWorkflow({
      id: workflow.id,
      name: workflow.name,
      description: workflow.description,
      nodes: mockData.nodes,
      edges: mockData.edges,
      settings: {
        autoSave: true,
        errorHandling: 'stop',
        retryCount: 3
      },
      createdAt: new Date(),
      updatedAt: new Date(),
      version: 1,
      isActive: workflow.status === 'active'
    });

    router.push(`/workflows/${workflow.id}`);
  };

  const getStatusBadge = (status: string) => {
    const variants = {
      active: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-green-200 dark:border-green-800',
      completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 border-green-200 dark:border-green-800',
      running: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border-blue-200 dark:border-blue-800',
      failed: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 border-red-200 dark:border-red-800',
      inactive: 'bg-gray-100 text-gray-700 dark:bg-gray-800/30 dark:text-gray-400 border-gray-200 dark:border-gray-700'
    };

    const labels = {
      active: 'Active',
      completed: 'Completed',
      running: 'Running',
      failed: 'Failed',
      inactive: 'Inactive'
    };

    return (
      <Badge className={`${variants[status as keyof typeof variants] || variants.inactive} border font-medium px-2.5 py-0.5`}>
        {labels[status as keyof typeof labels] || status}
      </Badge>
    );
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-green-600" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-600" />;
      case 'running':
        return <Loader2 className="w-4 h-4 text-blue-600 animate-spin" />;
      default:
        return <Clock className="w-4 h-4 text-gray-600" />;
    }
  };

  // Filter workflows based on search and status
  const filteredWorkflows = recentWorkflows.filter(workflow => {
    const matchesSearch = workflow.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      workflow.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesFilter = filterStatus === 'all' || workflow.status === filterStatus;
    return matchesSearch && matchesFilter;
  });

  return (
    <MainLayout>
      <div className="p-6 lg:p-8 space-y-8">
        {/* Header Section */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
            <p className="text-muted-foreground mt-1">Welcome back! Here's what's happening with your workflows.</p>
          </div>
          <Link href="/workflows/new">
            <Button size="lg" className="shadow-sm">
              <Plus className="w-4 h-4 mr-2" />
              New Workflow
            </Button>
          </Link>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {stats.map((stat, index) => (
            <Card key={index} className="border shadow-sm hover:shadow-md transition-all duration-200">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.title}
                </CardTitle>
                <div className={`p-2.5 rounded-lg ${stat.bgColor}`}>
                  <stat.icon className={`h-5 w-5 ${stat.iconColor}`} />
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="text-3xl font-bold">{stat.value}</div>
                {stat.showProgress && (
                  <div className="space-y-2">
                    <Progress value={stat.progressValue} className="h-2" />
                    <p className="text-xs text-muted-foreground">
                      {stat.change}
                    </p>
                  </div>
                )}
                {!stat.showProgress && (
                  <div className="flex items-center text-xs">
                    <ArrowUpRight className="w-3 h-3 text-green-600 mr-1" />
                    <span className="text-green-600 font-medium">{stat.trend}</span>
                    <span className="text-muted-foreground ml-1">from last week</span>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Main Content */}
        <Tabs value={selectedTab} onValueChange={setSelectedTab} className="space-y-6">
          <TabsList className="bg-muted/50">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="workflows">Recent Workflows</TabsTrigger>
            <TabsTrigger value="executions">Recent Executions</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Recent Workflows */}
              <Card className="shadow-sm border">
                <CardHeader className="flex flex-row items-center justify-between pb-4">
                  <div>
                    <CardTitle className="text-xl">Recent Workflows</CardTitle>
                    <CardDescription className="mt-1">Your most recently modified workflows</CardDescription>
                  </div>
                  <Link href="/workflows">
                    <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground">
                      View All
                      <ArrowUpRight className="w-4 h-4 ml-1" />
                    </Button>
                  </Link>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {recentWorkflows.length === 0 ? (
                      <div className="text-center py-12">
                        <GitBranch className="w-12 h-12 mx-auto text-muted-foreground/50 mb-3" />
                        <p className="text-muted-foreground">No workflows yet</p>
                        <Link href="/workflows/new">
                          <Button variant="outline" size="sm" className="mt-4">
                            <Plus className="w-4 h-4 mr-2" />
                            Create Your First Workflow
                          </Button>
                        </Link>
                      </div>
                    ) : (
                      recentWorkflows.map((workflow) => (
                        <div
                          key={workflow.id}
                          className="group flex items-start justify-between p-4 rounded-lg border bg-card hover:bg-accent/50 transition-all duration-200 cursor-pointer"
                          onClick={() => handleEditWorkflow(workflow)}
                        >
                          <div className="space-y-2 flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="font-semibold text-base truncate">{workflow.name}</span>
                              {getStatusBadge(workflow.status)}
                            </div>
                            <p className="text-sm text-muted-foreground line-clamp-1">
                              {workflow.description}
                            </p>
                            <div className="flex items-center gap-4 text-xs text-muted-foreground">
                              <span className="flex items-center gap-1">
                                <Clock className="w-3 h-3" />
                                {workflow.lastRun}
                              </span>
                              <span className="flex items-center gap-1">
                                <Activity className="w-3 h-3" />
                                {workflow.executions} runs
                              </span>
                              <span className="flex items-center gap-1">
                                <CheckCircle className="w-3 h-3" />
                                {workflow.successRate}%
                              </span>
                            </div>
                          </div>
                          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleEditWorkflow(workflow);
                              }}
                            >
                              <Edit className="w-4 h-4" />
                            </Button>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* Recent Executions */}
              <Card className="shadow-sm border">
                <CardHeader className="flex flex-row items-center justify-between pb-4">
                  <div>
                    <CardTitle className="text-xl">Recent Executions</CardTitle>
                    <CardDescription className="mt-1">Latest workflow executions</CardDescription>
                  </div>
                  <Link href="/executions">
                    <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground">
                      View All
                      <ArrowUpRight className="w-4 h-4 ml-1" />
                    </Button>
                  </Link>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {recentExecutions.map((execution) => (
                      <div
                        key={execution.id}
                        className="flex items-start justify-between p-4 rounded-lg border bg-card hover:bg-accent/50 transition-all duration-200"
                      >
                        <div className="space-y-2 flex-1">
                          <div className="flex items-center gap-2">
                            {getStatusIcon(execution.status)}
                            <span className="font-semibold text-base">{execution.workflowName}</span>
                            {getStatusBadge(execution.status)}
                          </div>
                          <div className="flex items-center gap-4 text-xs text-muted-foreground">
                            <span className="flex items-center gap-1">
                              <Clock className="w-3 h-3" />
                              {execution.duration}
                            </span>
                            <span className="flex items-center gap-1">
                              <GitBranch className="w-3 h-3" />
                              {execution.nodes} nodes
                            </span>
                            <span>{new Date(execution.startTime).toLocaleString()}</span>
                          </div>
                          {execution.error && (
                            <div className="flex items-start gap-2 text-xs text-red-600 bg-red-50 dark:bg-red-950/30 p-2 rounded">
                              <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
                              <span>{execution.error}</span>
                            </div>
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
            <Card className="shadow-sm border">
              <CardHeader>
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                  <div>
                    <CardTitle className="text-xl">All Workflows</CardTitle>
                    <CardDescription className="mt-1">Manage all your workflows</CardDescription>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="relative flex-1 md:w-64">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Search workflows..."
                        className="pl-9"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                      />
                    </div>
                    <select
                      className="px-3 py-2 border rounded-md text-sm bg-background"
                      value={filterStatus}
                      onChange={(e) => setFilterStatus(e.target.value)}
                    >
                      <option value="all">All Status</option>
                      <option value="active">Active</option>
                      <option value="inactive">Inactive</option>
                    </select>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {filteredWorkflows.length === 0 ? (
                    <div className="text-center py-12">
                      <Search className="w-12 h-12 mx-auto text-muted-foreground/50 mb-3" />
                      <p className="text-muted-foreground">No workflows found</p>
                      <p className="text-sm text-muted-foreground mt-1">Try adjusting your search or filter</p>
                    </div>
                  ) : (
                    filteredWorkflows.map((workflow) => (
                      <div key={workflow.id} className="flex items-center justify-between p-5 rounded-lg border bg-card hover:bg-accent/50 transition-all duration-200">
                        <div className="flex-1 space-y-2">
                          <div className="flex items-center gap-2">
                            <h3 className="font-semibold text-lg">{workflow.name}</h3>
                            {getStatusBadge(workflow.status)}
                          </div>
                          <p className="text-sm text-muted-foreground">{workflow.description}</p>
                          <div className="flex items-center gap-6 text-sm text-muted-foreground">
                            <span className="flex items-center gap-1.5">
                              <Activity className="w-4 h-4" />
                              {workflow.executions} executions
                            </span>
                            <span className="flex items-center gap-1.5">
                              <CheckCircle className="w-4 h-4" />
                              {workflow.successRate}% success
                            </span>
                            <span className="flex items-center gap-1.5">
                              <Clock className="w-4 h-4" />
                              Last run: {workflow.lastRun}
                            </span>
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditWorkflow(workflow)}
                          >
                            <Edit className="w-4 h-4 mr-2" />
                            Edit
                          </Button>
                          <Button variant="outline" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50">
                            <Trash2 className="w-4 h-4 mr-2" />
                            Delete
                          </Button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="executions">
            <Card className="shadow-sm border">
              <CardHeader>
                <CardTitle className="text-xl">Execution History</CardTitle>
                <CardDescription className="mt-1">View all workflow executions</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {recentExecutions.map((execution) => (
                    <div key={execution.id} className="flex items-center justify-between p-5 rounded-lg border bg-card hover:bg-accent/50 transition-all duration-200">
                      <div className="flex-1 space-y-2">
                        <div className="flex items-center gap-2">
                          {getStatusIcon(execution.status)}
                          <h3 className="font-semibold text-lg">{execution.workflowName}</h3>
                          {getStatusBadge(execution.status)}
                        </div>
                        <div className="flex items-center gap-6 text-sm text-muted-foreground">
                          <span>Started: {new Date(execution.startTime).toLocaleString()}</span>
                          <span>Duration: {execution.duration}</span>
                          <span>{execution.nodes} nodes</span>
                        </div>
                        {execution.error && (
                          <div className="flex items-start gap-2 text-sm text-red-600 bg-red-50 dark:bg-red-950/30 p-3 rounded">
                            <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
                            <span>{execution.error}</span>
                          </div>
                        )}
                      </div>
                      <Button variant="outline" size="sm">
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
              <Card className="shadow-sm border">
                <CardHeader>
                  <CardTitle className="text-xl">Execution Trends</CardTitle>
                  <CardDescription className="mt-1">Workflow execution trends over time</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="h-[300px] flex flex-col items-center justify-center text-muted-foreground">
                    <BarChart3 className="w-16 h-16 mb-4 opacity-50" />
                    <p className="text-lg font-medium">Analytics Chart</p>
                    <p className="text-sm mt-1">Chart visualization would be displayed here</p>
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-sm border">
                <CardHeader>
                  <CardTitle className="text-xl">Performance Metrics</CardTitle>
                  <CardDescription className="mt-1">Key performance indicators</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-6">
                    <div>
                      <div className="flex justify-between text-sm font-medium mb-3">
                        <span>Average Execution Time</span>
                        <span className="text-muted-foreground">2m 15s</span>
                      </div>
                      <Progress value={75} className="h-2" />
                    </div>
                    <div>
                      <div className="flex justify-between text-sm font-medium mb-3">
                        <span>Success Rate</span>
                        <span className="text-green-600">96.2%</span>
                      </div>
                      <Progress value={96} className="h-2 [&>div]:bg-green-600" />
                    </div>
                    <div>
                      <div className="flex justify-between text-sm font-medium mb-3">
                        <span>Resource Usage</span>
                        <span className="text-muted-foreground">68%</span>
                      </div>
                      <Progress value={68} className="h-2" />
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
