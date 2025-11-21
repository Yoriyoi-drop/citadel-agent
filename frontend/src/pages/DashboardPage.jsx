// frontend/src/pages/DashboardPage.jsx
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { 
  ChartBarIcon, 
  CogIcon, 
  DocumentTextIcon, 
  UserGroupIcon,
  QueueListIcon,
  ArrowPathIcon
} from '@heroicons/react/24/outline';

const DashboardPage = () => {
  const [stats, setStats] = useState({
    totalWorkflows: 0,
    activeWorkflows: 0,
    totalExecutions: 0,
    successfulExecutions: 0
  });

  const [recentWorkflows, setRecentWorkflows] = useState([]);
  const [recentExecutions, setRecentExecutions] = useState([]);

  // Simulated data - in real app this would come from API
  useEffect(() => {
    // Simulate API calls to get dashboard data
    setStats({
      totalWorkflows: 12,
      activeWorkflows: 8,
      totalExecutions: 156,
      successfulExecutions: 142
    });

    setRecentWorkflows([
      { id: 1, name: 'Data Processing Pipeline', status: 'active', lastRun: '2 hours ago' },
      { id: 2, name: 'User Notification Workflow', status: 'active', lastRun: '5 hours ago' },
      { id: 3, name: 'Report Generation', status: 'draft', lastRun: '1 day ago' },
      { id: 4, name: 'API Integration', status: 'active', lastRun: '3 hours ago' }
    ]);

    setRecentExecutions([
      { id: 1, workflow: 'Data Processing', status: 'success', duration: '2.3s', timestamp: '2 hours ago' },
      { id: 2, workflow: 'User Notification', status: 'success', duration: '1.1s', timestamp: '5 hours ago' },
      { id: 3, workflow: 'Report Generation', status: 'failed', duration: '0.8s', timestamp: '1 day ago' },
      { id: 4, workflow: 'API Integration', status: 'success', duration: '3.2s', timestamp: '3 hours ago' }
    ]);
  }, []);

  const statCards = [
    {
      id: 1,
      name: 'Total Workflows',
      value: stats.totalWorkflows,
      icon: QueueListIcon,
      change: '+2',
      changeType: 'positive'
    },
    {
      id: 2,
      name: 'Active Workflows',
      value: stats.activeWorkflows,
      icon: ChartBarIcon,
      change: '+1',
      changeType: 'positive'
    },
    {
      id: 3,
      name: 'Total Executions',
      value: stats.totalExecutions,
      icon: ArrowPathIcon,
      change: '+24',
      changeType: 'positive'
    },
    {
      id: 4,
      name: 'Success Rate',
      value: `${Math.round((stats.successfulExecutions / Math.max(stats.totalExecutions, 1)) * 100)}%`,
      icon: DocumentTextIcon,
      change: '+3%',
      changeType: 'positive'
    }
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
          <div className="mb-6">
            <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
            <p className="mt-1 text-sm text-gray-500">
              Welcome back! Here's what's happening with your workflows.
            </p>
          </div>

          {/* Stats Grid */}
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
            {statCards.map((stat) => (
              <div
                key={stat.id}
                className="bg-white overflow-hidden shadow rounded-lg"
              >
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <stat.icon className="h-6 w-6 text-gray-400" aria-hidden="true" />
                    </div>
                    <div className="ml-5 w-0 flex-1">
                      <dl>
                        <dt className="text-sm font-medium text-gray-500 truncate">{stat.name}</dt>
                        <dd className="flex items-baseline">
                          <div className="text-2xl font-semibold text-gray-900">{stat.value}</div>
                        </dd>
                      </dl>
                    </div>
                  </div>
                </div>
                <div className="bg-gray-50 px-5 py-3">
                  <div className="text-sm">
                    <span className={`text-${stat.changeType === 'positive' ? 'green' : 'red'}-600`}>
                      {stat.changeType === 'positive' ? '+' : '-'}
                      {stat.change}
                    </span>{' '}
                    from last week
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Recent Activity */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Recent Workflows */}
            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
              <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Recent Workflows</h3>
                <p className="mt-1 max-w-2xl text-sm text-gray-500">Your recently created or modified workflows</p>
              </div>
              <ul className="divide-y divide-gray-200">
                {recentWorkflows.map((workflow) => (
                  <li key={workflow.id}>
                    <Link to={`/workflows/${workflow.id}`} className="block hover:bg-gray-50">
                      <div className="px-4 py-4 sm:px-6">
                        <div className="flex items-center justify-between">
                          <p className="text-sm font-medium text-blue-600 truncate">{workflow.name}</p>
                          <div className="ml-2 flex-shrink-0 flex">
                            <p className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              workflow.status === 'active' 
                                ? 'bg-green-100 text-green-800' 
                                : 'bg-yellow-100 text-yellow-800'
                            }`}>
                              {workflow.status}
                            </p>
                          </div>
                        </div>
                        <div className="mt-2 sm:flex sm:justify-between">
                          <div className="sm:flex">
                            <p className="flex items-center text-sm text-gray-500">
                              Last run: {workflow.lastRun}
                            </p>
                          </div>
                          <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                            <span>View</span>
                          </div>
                        </div>
                      </div>
                    </Link>
                  </li>
                ))}
              </ul>
              <div className="bg-gray-50 px-4 py-3 sm:px-6">
                <Link
                  to="/workflows"
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  View all workflows
                </Link>
              </div>
            </div>

            {/* Recent Executions */}
            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
              <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Recent Executions</h3>
                <p className="mt-1 max-w-2xl text-sm text-gray-500">Latest workflow execution results</p>
              </div>
              <ul className="divide-y divide-gray-200">
                {recentExecutions.map((execution) => (
                  <li key={execution.id}>
                    <div className="px-4 py-4 sm:px-6">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium text-gray-900 truncate">{execution.workflow}</p>
                        <div className="ml-2 flex-shrink-0 flex">
                          <p className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            execution.status === 'success' 
                              ? 'bg-green-100 text-green-800' 
                              : 'bg-red-100 text-red-800'
                          }`}>
                            {execution.status}
                          </p>
                        </div>
                      </div>
                      <div className="mt-2 sm:flex sm:justify-between">
                        <div className="sm:flex">
                          <p className="flex items-center text-sm text-gray-500">
                            Duration: {execution.duration}
                          </p>
                        </div>
                        <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                          <span>{execution.timestamp}</span>
                        </div>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
              <div className="bg-gray-50 px-4 py-3 sm:px-6">
                <Link
                  to="/analytics"
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  View execution analytics
                </Link>
              </div>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="mt-8">
            <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">Quick Actions</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <Link
                to="/workflows/new"
                className="bg-white p-6 rounded-lg shadow border border-gray-200 hover:shadow-md transition-shadow duration-200"
              >
                <QueueListIcon className="h-8 w-8 text-blue-600 mb-3" />
                <h4 className="text-lg font-medium text-gray-900">Create Workflow</h4>
                <p className="mt-1 text-sm text-gray-500">Build a new automation workflow</p>
              </Link>
              
              <Link
                to="/analytics"
                className="bg-white p-6 rounded-lg shadow border border-gray-200 hover:shadow-md transition-shadow duration-200"
              >
                <ChartBarIcon className="h-8 w-8 text-green-600 mb-3" />
                <h4 className="text-lg font-medium text-gray-900">Analytics</h4>
                <p className="mt-1 text-sm text-gray-500">View workflow performance metrics</p>
              </Link>
              
              <Link
                to="/settings"
                className="bg-white p-6 rounded-lg shadow border border-gray-200 hover:shadow-md transition-shadow duration-200"
              >
                <CogIcon className="h-8 w-8 text-gray-600 mb-3" />
                <h4 className="text-lg font-medium text-gray-900">Settings</h4>
                <p className="mt-1 text-sm text-gray-500">Configure your account settings</p>
              </Link>
              
              <Link
                to="/docs"
                className="bg-white p-6 rounded-lg shadow border border-gray-200 hover:shadow-md transition-shadow duration-200"
              >
                <DocumentTextIcon className="h-8 w-8 text-purple-600 mb-3" />
                <h4 className="text-lg font-medium text-gray-900">Documentation</h4>
                <p className="mt-1 text-sm text-gray-500">Learn more about Citadel Agent</p>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;