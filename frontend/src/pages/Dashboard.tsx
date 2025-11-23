import {
    ChartBarIcon,
    PlayIcon,
    CheckCircleIcon,
    XCircleIcon,
    ClockIcon,
    BoltIcon
} from '@heroicons/react/24/outline';

const Dashboard = () => {
    const stats = [
        { name: 'Total Workflows', value: '12', icon: BoltIcon, change: '+4.75%', changeType: 'positive' },
        { name: 'Active Executions', value: '3', icon: PlayIcon, change: '+54.02%', changeType: 'positive' },
        { name: 'Completed Today', value: '24', icon: CheckCircleIcon, change: '+12.5%', changeType: 'positive' },
        { name: 'Failed', value: '2', icon: XCircleIcon, change: '-3.2%', changeType: 'negative' },
    ];

    const recentExecutions = [
        { id: 1, workflow: 'Data Sync Pipeline', status: 'completed', time: '2 min ago', duration: '1.2s' },
        { id: 2, workflow: 'Email Campaign', status: 'running', time: '5 min ago', duration: 'Running' },
        { id: 3, workflow: 'API Monitor', status: 'completed', time: '10 min ago', duration: '0.8s' },
        { id: 4, workflow: 'Database Backup', status: 'failed', time: '15 min ago', duration: '2.1s' },
    ];

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'completed': return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
            case 'running': return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
            case 'failed': return 'bg-red-500/10 text-red-400 border-red-500/20';
            default: return 'bg-gray-500/10 text-gray-400 border-gray-500/20';
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900/20 to-slate-900 p-8">
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-4xl font-bold text-white mb-2 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                        Dashboard
                    </h1>
                    <p className="text-gray-400">Monitor your workflow executions and system health</p>
                </div>

                {/* Stats Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                    {stats.map((stat) => (
                        <div
                            key={stat.name}
                            className="relative group"
                        >
                            <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 rounded-2xl blur-xl group-hover:blur-2xl transition-all duration-300" />
                            <div className="relative bg-slate-800/50 backdrop-blur-xl border border-slate-700/50 rounded-2xl p-6 hover:border-slate-600/50 transition-all duration-300">
                                <div className="flex items-center justify-between mb-4">
                                    <div className="p-3 bg-gradient-to-br from-blue-500/20 to-purple-500/20 rounded-xl">
                                        <stat.icon className="h-6 w-6 text-blue-400" />
                                    </div>
                                    <span className={`text-sm font-medium ${stat.changeType === 'positive' ? 'text-emerald-400' : 'text-red-400'}`}>
                                        {stat.change}
                                    </span>
                                </div>
                                <div>
                                    <p className="text-gray-400 text-sm mb-1">{stat.name}</p>
                                    <p className="text-3xl font-bold text-white">{stat.value}</p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Recent Executions */}
                <div className="relative group">
                    <div className="absolute inset-0 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl blur-xl" />
                    <div className="relative bg-slate-800/50 backdrop-blur-xl border border-slate-700/50 rounded-2xl p-6">
                        <div className="flex items-center justify-between mb-6">
                            <h2 className="text-2xl font-bold text-white">Recent Executions</h2>
                            <button className="px-4 py-2 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-lg hover:from-blue-600 hover:to-purple-600 transition-all duration-300 font-medium">
                                View All
                            </button>
                        </div>

                        <div className="space-y-3">
                            {recentExecutions.map((execution) => (
                                <div
                                    key={execution.id}
                                    className="flex items-center justify-between p-4 bg-slate-900/50 rounded-xl border border-slate-700/30 hover:border-slate-600/50 transition-all duration-300 group/item"
                                >
                                    <div className="flex items-center space-x-4 flex-1">
                                        <div className="flex-shrink-0">
                                            <ChartBarIcon className="h-8 w-8 text-blue-400" />
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <p className="text-white font-medium truncate group-hover/item:text-blue-400 transition-colors">
                                                {execution.workflow}
                                            </p>
                                            <div className="flex items-center space-x-3 mt-1">
                                                <span className="flex items-center text-sm text-gray-400">
                                                    <ClockIcon className="h-4 w-4 mr-1" />
                                                    {execution.time}
                                                </span>
                                                <span className="text-sm text-gray-500">•</span>
                                                <span className="text-sm text-gray-400">{execution.duration}</span>
                                            </div>
                                        </div>
                                    </div>
                                    <span className={`px-3 py-1 text-xs font-semibold rounded-full border ${getStatusColor(execution.status)}`}>
                                        {execution.status}
                                    </span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
