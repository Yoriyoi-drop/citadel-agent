// frontend/src/components/Dashboard/Dashboard.tsx
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ChartBarIcon,
  CogIcon,
  ClockIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
  RectangleStackIcon,
  UserGroupIcon,
  CloudIcon,
  DocumentTextIcon,
  ArrowTrendingUpIcon
} from '@heroicons/react/24/solid';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { Line } from 'react-chartjs-2';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

interface Stats {
  totalWorkflows: number;
  activeWorkflows: number;
  executionCount: number;
  successRate: number;
  avgExecutionTime: number;
}

interface Execution {
  id: string;
  workflow: string;
  status: 'success' | 'failed' | 'running' | 'cancelled';
  startedAt: string;
  duration: number; // in seconds
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [stats] = useState<Stats>({
    totalWorkflows: 12,
    activeWorkflows: 8,
    executionCount: 1248,
    successRate: 96.2,
    avgExecutionTime: 4.2
  });

  const [executions] = useState<Execution[]>([
    {
      id: 'exec-1',
      workflow: 'Email Notification Workflow',
      status: 'success',
      startedAt: '2023-06-15T10:30:00Z',
      duration: 2.5
    },
    {
      id: 'exec-2',
      workflow: 'Data Processing Pipeline',
      status: 'success',
      startedAt: '2023-06-15T10:25:00Z',
      duration: 8.1
    },
    {
      id: 'exec-3',
      workflow: 'API Integration',
      status: 'failed',
      startedAt: '2023-06-15T10:20:00Z',
      duration: 1.2
    },
    {
      id: 'exec-4',
      workflow: 'Email Notification Workflow',
      status: 'success',
      startedAt: '2023-06-15T10:15:00Z',
      duration: 1.8
    },
    {
      id: 'exec-5',
      workflow: 'Report Generation',
      status: 'running',
      startedAt: '2023-06-15T10:10:00Z',
      duration: 15.0
    }
  ]);

  const [chartData, setChartData] = useState<any>(null);

  useEffect(() => {
    // Mock chart data - in real implementation this would come from API
    const labels = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    const data = [65, 59, 80, 81, 56, 55, 40];

    setChartData({
      labels,
      datasets: [
        {
          label: 'Executions',
          data,
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.5)',
        },
      ],
    });
  }, []);

  const getStatusIcon = (status: Execution['status']) => {
    switch (status) {
      case 'success':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
      case 'failed':
        return <XCircleIcon className="h-5 w-5 text-red-500" />;
      case 'running':
        return <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />;
      case 'cancelled':
        return <XCircleIcon className="h-5 w-5 text-gray-500" />;
      default:
        return <ExclamationTriangleIcon className="h-5 w-5 text-gray-500" />;
    }
  };

  const getStatusColor = (status: Execution['status']) => {
    switch (status) {
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'running':
        return 'bg-yellow-100 text-yellow-800';
      case 'cancelled':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="flex-1 overflow-y-auto p-4 md:p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600">Monitor your workflows and automation performance</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-blue-100">
              <RectangleStackIcon className="h-6 w-6 text-blue-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Workflows</p>
              <p className="text-2xl font-semibold text-gray-900">{stats.totalWorkflows}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-green-100">
              <CheckCircleIcon className="h-6 w-6 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Active Workflows</p>
              <p className="text-2xl font-semibold text-gray-900">{stats.activeWorkflows}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-purple-100">
              <ChartBarIcon className="h-6 w-6 text-purple-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Executions</p>
              <p className="text-2xl font-semibold text-gray-900">{stats.executionCount}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-yellow-100">
              <ArrowTrendingUpIcon className="h-6 w-6 text-yellow-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Success Rate</p>
              <p className="text-2xl font-semibold text-gray-900">{stats.successRate}%</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-3 rounded-full bg-indigo-100">
              <ClockIcon className="h-6 w-6 text-indigo-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Avg. Time</p>
              <p className="text-2xl font-semibold text-gray-900">{stats.avgExecutionTime}s</p>
            </div>
          </div>
        </div>
      </div>

      {/* Charts and Recent Executions */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Execution Trend Chart */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Execution Trend</h2>
          {chartData && (
            <div className="h-64">
              <Line data={chartData} options={{ responsive: true, maintainAspectRatio: false }} />
            </div>
          )}
        </div>

        {/* Recent Executions */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Recent Executions</h2>
          <div className="overflow-hidden">
            <ul className="divide-y divide-gray-200">
              {executions.map((execution) => (
                <li key={execution.id} className="py-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      {getStatusIcon(execution.status)}
                      <div className="ml-3">
                        <p className="text-sm font-medium text-gray-900">{execution.workflow}</p>
                        <p className="text-sm text-gray-500">{formatDate(execution.startedAt)}</p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-4">
                      <span className={`px - 2 inline - flex text - xs leading - 5 font - semibold rounded - full ${getStatusColor(execution.status)} `}>
                        {execution.status}
                      </span>
                      <span className="text-sm text-gray-500">{execution.duration}s</span>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg shadow p-6 mb-8">
        <h2 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button
            onClick={() => navigate('/workflows/new')}
            className="flex flex-col items-center justify-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <RectangleStackIcon className="h-8 w-8 text-blue-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">New Workflow</span>
          </button>
          <button className="flex flex-col items-center justify-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
            <CloudIcon className="h-8 w-8 text-green-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">Integrations</span>
          </button>
          <button className="flex flex-col items-center justify-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
            <UserGroupIcon className="h-8 w-8 text-purple-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">AI Agents</span>
          </button>
          <button className="flex flex-col items-center justify-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
            <DocumentTextIcon className="h-8 w-8 text-yellow-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">Documentation</span>
          </button>
        </div>
      </div>

      {/* Node Categories */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-medium text-gray-900 mb-4">Node Categories</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <div className="flex flex-col items-center p-4 border border-gray-200 rounded-lg">
            <CloudIcon className="h-8 w-8 text-blue-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">Integrations</span>
            <span className="text-xs text-gray-500">24 nodes</span>
          </div>
          <div className="flex flex-col items-center p-4 border border-gray-200 rounded-lg">
            <CogIcon className="h-8 w-8 text-green-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">Utilities</span>
            <span className="text-xs text-gray-500">18 nodes</span>
          </div>
          <div className="flex flex-col items-center p-4 border border-gray-200 rounded-lg">
            <DocumentTextIcon className="h-8 w-8 text-yellow-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">Data Processing</span>
            <span className="text-xs text-gray-500">15 nodes</span>
          </div>
          <div className="flex flex-col items-center p-4 border border-gray-200 rounded-lg">
            <UserGroupIcon className="h-8 w-8 text-purple-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">AI Agents</span>
            <span className="text-xs text-gray-500">12 nodes</span>
          </div>
          <div className="flex flex-col items-center p-4 border border-gray-200 rounded-lg">
            <ArrowTrendingUpIcon className="h-8 w-8 text-red-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">Analytics</span>
            <span className="text-xs text-gray-500">9 nodes</span>
          </div>
          <div className="flex flex-col items-center p-4 border border-gray-200 rounded-lg">
            <XCircleIcon className="h-8 w-8 text-gray-600 mb-2" />
            <span className="text-sm font-medium text-gray-900">System</span>
            <span className="text-xs text-gray-500">8 nodes</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;