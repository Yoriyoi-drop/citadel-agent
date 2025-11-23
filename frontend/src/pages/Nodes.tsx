import { useState } from 'react';
import { MagnifyingGlassIcon, FunnelIcon } from '@heroicons/react/24/outline';

const Nodes = () => {
    const [nodes] = useState<any[]>([
        { id: 'http-request', name: 'HTTP Request', category: 'Network', description: 'Make HTTP requests to external APIs', color: 'from-blue-500 to-cyan-500' },
        { id: 'postgres-query', name: 'PostgreSQL Query', category: 'Database', description: 'Execute SQL queries on PostgreSQL', color: 'from-emerald-500 to-teal-500' },
        { id: 'openai-chat', name: 'OpenAI Chat', category: 'AI', description: 'Generate text using GPT models', color: 'from-purple-500 to-pink-500' },
        { id: 'json-transform', name: 'JSON Transform', category: 'Utility', description: 'Parse and stringify JSON data', color: 'from-orange-500 to-red-500' },
        { id: 'slack-message', name: 'Slack Message', category: 'Communication', description: 'Send messages to Slack channels', color: 'from-violet-500 to-purple-500' },
        { id: 'email-send', name: 'Send Email', category: 'Communication', description: 'Send emails via SMTP', color: 'from-blue-500 to-indigo-500' },
    ]);
    const [search, setSearch] = useState('');
    const [selectedCategory, setSelectedCategory] = useState('all');

    const categories = ['all', ...Array.from(new Set(nodes.map(node => node.category)))];

    const filteredNodes = nodes.filter(node => {
        const matchesSearch = node.name.toLowerCase().includes(search.toLowerCase()) ||
            node.category.toLowerCase().includes(search.toLowerCase());
        const matchesCategory = selectedCategory === 'all' || node.category === selectedCategory;
        return matchesSearch && matchesCategory;
    });

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900/20 to-slate-900 p-8">
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-4xl font-bold text-white mb-2 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                        Node Library
                    </h1>
                    <p className="text-gray-400">Discover and use powerful workflow nodes</p>
                </div>

                {/* Filters */}
                <div className="flex flex-col md:flex-row gap-4 mb-8">
                    <div className="relative flex-1">
                        <MagnifyingGlassIcon className="h-5 w-5 absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-400" />
                        <input
                            type="text"
                            placeholder="Search nodes..."
                            className="w-full pl-12 pr-4 py-3 bg-slate-800/50 backdrop-blur-xl border border-slate-700/50 rounded-xl text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent transition-all"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                    </div>
                    <div className="flex items-center space-x-2 bg-slate-800/50 backdrop-blur-xl border border-slate-700/50 rounded-xl px-4 py-3">
                        <FunnelIcon className="h-5 w-5 text-gray-400" />
                        <select
                            className="bg-transparent text-white focus:outline-none cursor-pointer"
                            value={selectedCategory}
                            onChange={(e) => setSelectedCategory(e.target.value)}
                        >
                            {categories.map(cat => (
                                <option key={cat} value={cat} className="bg-slate-800">
                                    {cat.charAt(0).toUpperCase() + cat.slice(1)}
                                </option>
                            ))}
                        </select>
                    </div>
                </div>

                {/* Nodes Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {filteredNodes.map(node => (
                        <div key={node.id} className="relative group cursor-pointer">
                            <div className={`absolute inset-0 bg-gradient-to-r ${node.color} opacity-20 rounded-2xl blur-xl group-hover:blur-2xl transition-all duration-300`} />
                            <div className="relative bg-slate-800/50 backdrop-blur-xl border border-slate-700/50 rounded-2xl p-6 hover:border-slate-600/50 transition-all duration-300 h-full">
                                <div className="flex items-start justify-between mb-4">
                                    <div className={`p-3 bg-gradient-to-br ${node.color} rounded-xl`}>
                                        <svg className="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                                        </svg>
                                    </div>
                                    <span className="px-3 py-1 text-xs font-semibold rounded-full bg-slate-900/50 text-gray-300 border border-slate-700/30">
                                        {node.category}
                                    </span>
                                </div>
                                <h3 className="text-xl font-bold text-white mb-2 group-hover:text-blue-400 transition-colors">{node.name}</h3>
                                <p className="text-gray-400 text-sm leading-relaxed">{node.description}</p>

                                <div className="mt-4 pt-4 border-t border-slate-700/30">
                                    <button className="text-sm text-blue-400 hover:text-blue-300 font-medium transition-colors">
                                        Add to workflow →
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>

                {filteredNodes.length === 0 && (
                    <div className="text-center py-12">
                        <p className="text-gray-400 text-lg">No nodes found matching your criteria</p>
                    </div>
                )}
            </div>
        </div>
    );
};

export default Nodes;
