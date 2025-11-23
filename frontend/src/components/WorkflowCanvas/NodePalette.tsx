import React, { useState } from 'react';
import { MagnifyingGlassIcon, XMarkIcon } from '@heroicons/react/24/outline';
import {
    BoltIcon,
    CommandLineIcon,
    ArrowsRightLeftIcon,
    ArrowPathIcon,
    CpuChipIcon
} from '@heroicons/react/24/solid';

interface NodePaletteProps {
    isOpen: boolean;
    onClose: () => void;
    onAddNode: (type: string, label: string) => void;
}

const NodePalette: React.FC<NodePaletteProps> = ({ isOpen, onClose, onAddNode }) => {
    const [searchQuery, setSearchQuery] = useState('');
    const [activeCategory, setActiveCategory] = useState('all');

    if (!isOpen) return null;

    const categories = [
        { id: 'all', name: 'All Nodes' },
        { id: 'trigger', name: 'Triggers' },
        { id: 'action', name: 'Actions' },
        { id: 'logic', name: 'Logic' },
        { id: 'ai', name: 'AI & ML' },
    ];

    const nodeTypes = [
        {
            type: 'trigger',
            label: 'Webhook',
            description: 'Starts workflow on HTTP request',
            icon: BoltIcon,
            category: 'trigger'
        },
        {
            type: 'trigger',
            label: 'Schedule',
            description: 'Runs workflow on a schedule',
            icon: BoltIcon,
            category: 'trigger'
        },
        {
            type: 'action',
            label: 'HTTP Request',
            description: 'Make an API call',
            icon: CommandLineIcon,
            category: 'action'
        },
        {
            type: 'action',
            label: 'Send Email',
            description: 'Send an email via SMTP',
            icon: CommandLineIcon,
            category: 'action'
        },
        {
            type: 'condition',
            label: 'If / Else',
            description: 'Branch based on conditions',
            icon: ArrowsRightLeftIcon,
            category: 'logic'
        },
        {
            type: 'loop',
            label: 'Loop',
            description: 'Iterate over items',
            icon: ArrowPathIcon,
            category: 'logic'
        },
        {
            type: 'ai',
            label: 'AI Agent',
            description: 'Process with LLM',
            icon: CpuChipIcon,
            category: 'ai'
        },
    ];

    const filteredNodes = nodeTypes.filter(node => {
        const matchesSearch = node.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
            node.description.toLowerCase().includes(searchQuery.toLowerCase());
        const matchesCategory = activeCategory === 'all' || node.category === activeCategory;
        return matchesSearch && matchesCategory;
    });

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="bg-white w-[600px] h-[500px] rounded-2xl shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in duration-200">
                {/* Header */}
                <div className="p-4 border-b border-gray-100 flex items-center justify-between">
                    <div className="relative flex-1 mr-4">
                        <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
                        <input
                            type="text"
                            placeholder="Search nodes..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-10 pr-4 py-2 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                            autoFocus
                        />
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-gray-100 rounded-lg text-gray-400 hover:text-gray-600 transition-colors"
                    >
                        <XMarkIcon className="h-6 w-6" />
                    </button>
                </div>

                <div className="flex flex-1 overflow-hidden">
                    {/* Sidebar Categories */}
                    <div className="w-40 bg-gray-50 border-r border-gray-100 p-2 overflow-y-auto">
                        {categories.map(category => (
                            <button
                                key={category.id}
                                onClick={() => setActiveCategory(category.id)}
                                className={`w-full text-left px-3 py-2 rounded-lg text-sm font-medium mb-1 transition-colors ${activeCategory === category.id
                                    ? 'bg-white text-blue-600 shadow-sm'
                                    : 'text-gray-600 hover:bg-gray-100'
                                    }`}
                            >
                                {category.name}
                            </button>
                        ))}
                    </div>

                    {/* Node List */}
                    <div className="flex-1 p-4 overflow-y-auto">
                        <div className="grid grid-cols-1 gap-2">
                            {filteredNodes.map((node, index) => (
                                <button
                                    key={index}
                                    onClick={() => {
                                        onAddNode(node.type, node.label);
                                        onClose();
                                    }}
                                    className="flex items-start p-3 hover:bg-blue-50 border border-transparent hover:border-blue-100 rounded-xl transition-all group text-left"
                                >
                                    <div className={`p-2 rounded-lg mr-3 ${node.category === 'trigger' ? 'bg-purple-100 text-purple-600' :
                                        node.category === 'action' ? 'bg-blue-100 text-blue-600' :
                                            node.category === 'logic' ? 'bg-orange-100 text-orange-600' :
                                                'bg-green-100 text-green-600'
                                        }`}>
                                        <node.icon className="h-5 w-5" />
                                    </div>
                                    <div>
                                        <h4 className="font-semibold text-gray-900 group-hover:text-blue-700">{node.label}</h4>
                                        <p className="text-xs text-gray-500 mt-0.5">{node.description}</p>
                                    </div>
                                </button>
                            ))}

                            {filteredNodes.length === 0 && (
                                <div className="text-center py-12 text-gray-400">
                                    <p>No nodes found matching "{searchQuery}"</p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NodePalette;
