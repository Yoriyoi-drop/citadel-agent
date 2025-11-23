import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
    PlusIcon,
    PlayIcon,
    PencilIcon,
    TrashIcon,
    ClockIcon,
    BoltIcon
} from '@heroicons/react/24/outline';

const Workflows = () => {
    const navigate = useNavigate();
    const [workflows, setWorkflows] = useState([
        {
            id: 1,
            name: 'Data Sync Pipeline',
            description: 'Syncs data from multiple sources',
            status: 'active',
            lastRun: '2 hours ago',
            executions: 145,
            nodes: 8
        },
        {
            id: 2,
            name: 'Email Campaign',
            description: 'Automated email marketing workflow',
            status: 'active',
            lastRun: '1 day ago',
            executions: 89,
            nodes: 12
        },
        {
            id: 3,
            name: 'API Monitor',
            description: 'Monitors API health and performance',
            status: 'inactive',
            lastRun: '3 days ago',
            executions: 234,
            nodes: 5
        },
    ]);

    const handleDelete = (id: number) => {
        if (confirm('Are you sure you want to delete this workflow?')) {
            setWorkflows(workflows.filter(w => w.id !== id));
        }
    };

    const handleRun = (name: string) => {
        alert(`Started execution for workflow: ${name}`);
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900/20 to-slate-900 p-4 md:p-8">
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                    <div>
                        <h1 className="text-4xl font-bold text-white mb-2 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                            Workflows
                        </h1>
                        <p className="text-gray-400">Create and manage your automation workflows</p>
                    </div>
                    <Link to="/workflows/new" className="w-full md:w-auto flex justify-center items-center space-x-2 px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-xl hover:from-blue-600 hover:to-purple-600 transition-all duration-300 font-medium shadow-lg shadow-blue-500/50">
                        <PlusIcon className="h-5 w-5" />
                        <span>New Workflow</span>
                    </Link>
                </div>

                {/* Workflows Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {workflows.map((workflow) => (
                        <div key={workflow.id} className="relative group">
                            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-2xl blur-xl group-hover:blur-2xl transition-all duration-300" />
                            <div className="relative bg-slate-800/50 backdrop-blur-xl border border-slate-700/50 rounded-2xl p-6 hover:border-slate-600/50 transition-all duration-300">
                                {/* Status Badge */}
                                <div className="flex items-center justify-between mb-4">
                                    <span className={`px-3 py-1 text-xs font-semibold rounded-full ${workflow.status === 'active'
                                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                        : 'bg-gray-500/10 text-gray-400 border border-gray-500/20'
                                        }`}>
                                        {workflow.status}
                                    </span>
                                    <div className="flex space-x-2">
                                        <button
                                            onClick={() => handleRun(workflow.name)}
                                            className="p-2 hover:bg-slate-700/50 rounded-lg transition-colors"
                                            title="Run Workflow"
                                        >
                                            <PlayIcon className="h-4 w-4 text-blue-400" />
                                        </button>
                                        <button
                                            onClick={() => navigate(`/workflows/${workflow.id}`)}
                                            className="p-2 hover:bg-slate-700/50 rounded-lg transition-colors"
                                            title="Edit Workflow"
                                        >
                                            <PencilIcon className="h-4 w-4 text-gray-400" />
                                        </button>
                                        <button
                                            onClick={() => handleDelete(workflow.id)}
                                            className="p-2 hover:bg-slate-700/50 rounded-lg transition-colors"
                                            title="Delete Workflow"
                                        >
                                            <TrashIcon className="h-4 w-4 text-red-400" />
                                        </button>
                                    </div>
                                </div>

                                {/* Workflow Info */}
                                <div className="mb-4">
                                    <h3 className="text-xl font-bold text-white mb-2 group-hover:text-blue-400 transition-colors">
                                        {workflow.name}
                                    </h3>
                                    <p className="text-gray-400 text-sm line-clamp-2">{workflow.description}</p>
                                </div>

                                {/* Stats */}
                                <div className="grid grid-cols-2 gap-4 mb-4">
                                    <div className="bg-slate-900/50 rounded-lg p-3">
                                        <div className="flex items-center space-x-2 mb-1">
                                            <BoltIcon className="h-4 w-4 text-blue-400" />
                                            <span className="text-xs text-gray-400">Executions</span>
                                        </div>
                                        <p className="text-lg font-bold text-white">{workflow.executions}</p>
                                    </div>
                                    <div className="bg-slate-900/50 rounded-lg p-3">
                                        <div className="flex items-center space-x-2 mb-1">
                                            <svg className="h-4 w-4 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16m-7 6h7" />
                                            </svg>
                                            <span className="text-xs text-gray-400">Nodes</span>
                                        </div>
                                        <p className="text-lg font-bold text-white">{workflow.nodes}</p>
                                    </div>
                                </div>

                                {/* Last Run */}
                                <div className="flex items-center text-sm text-gray-400">
                                    <ClockIcon className="h-4 w-4 mr-2" />
                                    Last run {workflow.lastRun}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default Workflows;
