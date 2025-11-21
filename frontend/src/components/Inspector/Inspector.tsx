// frontend/src/components/Inspector/Inspector.tsx
import React, { useState } from 'react';
import { 
  CogIcon, 
  CodeIcon, 
  DocumentTextIcon, 
  HashtagIcon, 
  TerminalIcon,
  PencilIcon,
  TrashIcon,
  DotsVerticalIcon,
  ChipIcon,
  DocumentSearchIcon,
  AnnotationIcon
} from '@heroicons/react/solid';
import { NodeData } from '../../types';

interface InspectorProps {
  node: NodeData;
}

const Inspector: React.FC<InspectorProps> = ({ node }) => {
  const [activeTab, setActiveTab] = useState('properties');
  const [nodeProperties, setNodeProperties] = useState({
    label: node.label,
    description: node.description,
    ...node.parameters
  });

  // Mock available nodes for the selected node type
  const availableNodes = [
    { id: 'http-request', name: 'HTTP Request', type: 'action', category: 'Integrations' },
    { id: 'data-transform', name: 'Data Transform', type: 'utility', category: 'Data Processing' },
    { id: 'condition', name: 'Condition', type: 'logic', category: 'Logic' },
    { id: 'loop', name: 'Loop', type: 'logic', category: 'Logic' },
    { id: 'ai-agent', name: 'AI Agent', type: 'ai', category: 'AI/ML' },
  ];

  const handlePropertyChange = (key: string, value: any) => {
    setNodeProperties({
      ...nodeProperties,
      [key]: value
    });
  };

  const renderPropertyField = (key: string, value: any) => {
    switch (typeof value) {
      case 'boolean':
        return (
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={value}
              onChange={(e) => handlePropertyChange(key, e.target.checked)}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
          </div>
        );
      case 'number':
        return (
          <input
            type="number"
            value={value}
            onChange={(e) => handlePropertyChange(key, Number(e.target.value))}
            className="w-full px-3 py-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          />
        );
      default:
        return (
          <input
            type="text"
            value={value}
            onChange={(e) => handlePropertyChange(key, e.target.value)}
            className="w-full px-3 py-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          />
        );
    }
  };

  return (
    <div className="w-80 bg-white shadow-lg border-l h-full flex flex-col">
      {/* Header */}
      <div className="border-b p-4">
        <div className="flex justify-between items-center">
          <h3 className="text-lg font-semibold text-gray-900">Inspector</h3>
          <button className="text-gray-500 hover:text-gray-700">
            <DotsVerticalIcon className="h-5 w-5" />
          </button>
        </div>
        <p className="text-sm text-gray-500 mt-1">{node.type} node</p>
      </div>

      {/* Tabs */}
      <div className="border-b">
        <nav className="flex -mb-px">
          <button
            onClick={() => setActiveTab('properties')}
            className={`${
              activeTab === 'properties'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-2 px-4 text-sm font-medium border-b-2`}
          >
            Properties
          </button>
          <button
            onClick={() => setActiveTab('data')}
            className={`${
              activeTab === 'data'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-2 px-4 text-sm font-medium border-b-2`}
          >
            Data
          </button>
          <button
            onClick={() => setActiveTab('config')}
            className={`${
              activeTab === 'config'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-2 px-4 text-sm font-medium border-b-2`}
          >
            Configuration
          </button>
        </nav>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {activeTab === 'properties' && (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                type="text"
                value={nodeProperties.label}
                onChange={(e) => handlePropertyChange('label', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                value={nodeProperties.description}
                onChange={(e) => handlePropertyChange('description', e.target.value)}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Node Type</label>
              <div className="relative">
                <select
                  value={node.type}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white"
                >
                  <option value="trigger">Trigger</option>
                  <option value="action">Action</option>
                  <option value="condition">Condition</option>
                  <option value="loop">Loop</option>
                  <option value="ai">AI Agent</option>
                </select>
              </div>
            </div>

            {/* Parameters based on node type */}
            <div className="space-y-4">
              <h4 className="text-sm font-medium text-gray-700 flex items-center">
                <CogIcon className="h-4 w-4 mr-1" />
                Parameters
              </h4>
              
              {Object.keys(nodeProperties).filter(key => !['label', 'description', 'type'].includes(key)).map(key => (
                <div key={key}>
                  <label className="block text-sm font-medium text-gray-700 mb-1 capitalize">{key.replace(/([A-Z])/g, ' $1').trim()}</label>
                  {renderPropertyField(key, nodeProperties[key as keyof typeof nodeProperties])}
                </div>
              ))}
              
              <button className="w-full mt-2 px-3 py-1.5 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                + Add Parameter
              </button>
            </div>
          </div>
        )}

        {activeTab === 'data' && (
          <div className="space-y-6">
            <div>
              <h4 className="text-sm font-medium text-gray-700 flex items-center mb-2">
                <DocumentSearchIcon className="h-4 w-4 mr-1" />
                Input Data
              </h4>
              <div className="bg-gray-50 p-3 rounded-md">
                <pre className="text-xs overflow-x-auto">
                  {JSON.stringify(node.parameters, null, 2)}
                </pre>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-700 flex items-center mb-2">
                <AnnotationIcon className="h-4 w-4 mr-1" />
                Sample Output
              </h4>
              <div className="bg-gray-50 p-3 rounded-md">
                <pre className="text-xs overflow-x-auto">
                  {`{
  "status": "success",
  "data": {},
  "timestamp": "2023-06-15T10:30:00Z"
}`}
                </pre>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-700 mb-2">Data Mapping</h4>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Input</span>
                  <span className="text-gray-500">Output</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>request.url</span>
                  <span className="text-blue-600">response.data.url</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>request.method</span>
                  <span className="text-blue-600">response.data.method</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'config' && (
          <div className="space-y-6">
            <div>
              <h4 className="text-sm font-medium text-gray-700 flex items-center mb-2">
                <ChipIcon className="h-4 w-4 mr-1" />
                Execution Settings
              </h4>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm">Retry on failure</span>
                  <input
                    type="checkbox"
                    defaultChecked
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Max retries</label>
                  <input
                    type="number"
                    defaultValue="3"
                    className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Timeout (seconds)</label>
                  <input
                    type="number"
                    defaultValue="30"
                    className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                  />
                </div>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-700 mb-2">Error Handling</h4>
              <div className="space-y-2">
                <div className="flex items-center">
                  <input
                    type="radio"
                    name="errorHandling"
                    defaultChecked
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                  />
                  <label className="ml-2 block text-sm text-gray-700">
                    Stop workflow
                  </label>
                </div>
                <div className="flex items-center">
                  <input
                    type="radio"
                    name="errorHandling"
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                  />
                  <label className="ml-2 block text-sm text-gray-700">
                    Continue with fallback value
                  </label>
                </div>
                <div className="flex items-center">
                  <input
                    type="radio"
                    name="errorHandling"
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                  />
                  <label className="ml-2 block text-sm text-gray-700">
                    Execute error handler
                  </label>
                </div>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-700 mb-2">Metadata</h4>
              <div className="space-y-2">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Version</label>
                  <input
                    type="text"
                    defaultValue="1.0.0"
                    className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Author</label>
                  <input
                    type="text"
                    defaultValue="Current User"
                    className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                  />
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="border-t p-4 space-y-3">
        <button className="w-full flex justify-center items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
          <TrashIcon className="h-4 w-4 mr-2" />
          Delete Node
        </button>
        <button className="w-full flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
          <PencilIcon className="h-4 w-4 mr-2" />
          Apply Changes
        </button>
      </div>
    </div>
  );
};

export default Inspector;