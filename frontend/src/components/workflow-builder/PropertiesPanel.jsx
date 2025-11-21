// frontend/src/components/workflow-builder/PropertiesPanel.jsx
import React, { useState, useEffect } from 'react';

const PropertiesPanel = ({ selectedNode, onSave, onCancel }) => {
  const [config, setConfig] = useState({});
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    if (selectedNode) {
      setConfig(selectedNode.data.config || {});
      setDirty(false);
    } else {
      setConfig({});
      setDirty(false);
    }
  }, [selectedNode]);

  const handleInputChange = (field, value) => {
    setConfig(prev => ({
      ...prev,
      [field]: value
    }));
    setDirty(true);
  };

  const handleSave = () => {
    if (selectedNode && onSave) {
      onSave({
        ...selectedNode,
        data: {
          ...selectedNode.data,
          config: { ...config }
        }
      });
      setDirty(false);
    }
  };

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    }
    setConfig(selectedNode?.data.config || {});
    setDirty(false);
  };

  if (!selectedNode) {
    return (
      <div className="h-full flex items-center justify-center text-gray-500">
        <div className="text-center">
          <div className="text-4xl mb-2">🔍</div>
          <div>Select a node to edit properties</div>
        </div>
      </div>
    );
  }

  const renderPropertyForm = () => {
    switch (selectedNode.type) {
      case 'start':
        return (
          <div>
            <h4 className="font-semibold mb-3">Start Node</h4>
            <p className="text-sm text-gray-600">Beginning of the workflow. No configuration needed.</p>
          </div>
        );
      case 'end':
        return (
          <div>
            <h4 className="font-semibold mb-3">End Node</h4>
            <p className="text-sm text-gray-600">End of the workflow. No configuration needed.</p>
          </div>
        );
      case 'http':
        return (
          <div>
            <h4 className="font-semibold mb-3">HTTP Request Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Method</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.method || 'GET'}
                  onChange={(e) => handleInputChange('method', e.target.value)}
                >
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                  <option value="PATCH">PATCH</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">URL</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://api.example.com/data"
                  value={config.url || ''}
                  onChange={(e) => handleInputChange('url', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Headers (JSON)</label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                  placeholder='{\n  "Authorization": "Bearer ...",\n  "Content-Type": "application/json"\n}'
                  rows="4"
                  value={config.headers || ''}
                  onChange={(e) => handleInputChange('headers', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Body (JSON)</label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                  placeholder='{\n  "key": "value"\n}'
                  rows="4"
                  value={config.body || ''}
                  onChange={(e) => handleInputChange('body', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'database':
        return (
          <div>
            <h4 className="font-semibold mb-3">Database Query Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Database Type</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.databaseType || 'postgresql'}
                  onChange={(e) => handleInputChange('databaseType', e.target.value)}
                >
                  <option value="postgresql">PostgreSQL</option>
                  <option value="mysql">MySQL</option>
                  <option value="sqlite">SQLite</option>
                  <option value="mongodb">MongoDB</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Connection Name</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="my_database_connection"
                  value={config.connectionName || ''}
                  onChange={(e) => handleInputChange('connectionName', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Query</label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                  placeholder="SELECT * FROM users WHERE id = ?"
                  rows="6"
                  value={config.query || ''}
                  onChange={(e) => handleInputChange('query', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Parameters (JSON)</label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                  placeholder="[1, \"user@example.com\"]"
                  rows="3"
                  value={config.parameters || ''}
                  onChange={(e) => handleInputChange('parameters', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'decision':
        return (
          <div>
            <h4 className="font-semibold mb-3">Decision Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Condition</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="value > 10"
                  value={config.condition || ''}
                  onChange={(e) => handleInputChange('condition', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operator</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operator || '=='}
                  onChange={(e) => handleInputChange('operator', e.target.value)}
                >
                  <option value="==">Equals (==)</option>
                  <option value="!=">Not Equals (!=)</option>
                  <option value=">">Greater Than (>)</option>
                  <option value="<">Less Than (<)</option>
                  <option value=">=">Greater or Equal (>=)</option>
                  <option value="<=">Less or Equal (<=)</option>
                  <option value="contains">Contains</option>
                  <option value="startsWith">Starts With</option>
                  <option value="endsWith">Ends With</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="expected value"
                  value={config.value || ''}
                  onChange={(e) => handleInputChange('value', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'delay':
        return (
          <div>
            <h4 className="font-semibold mb-3">Delay Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Duration (seconds)</label>
                <input
                  type="number"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="5"
                  value={config.duration || ''}
                  onChange={(e) => handleInputChange('duration', parseInt(e.target.value))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Delay Type</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.delayType || 'fixed'}
                  onChange={(e) => handleInputChange('delayType', e.target.value)}
                >
                  <option value="fixed">Fixed Duration</option>
                  <option value="random">Random Duration</option>
                  <option value="until">Until Specific Time</option>
                </select>
              </div>
            </div>
          </div>
        );
      case 'ai':
        return (
          <div>
            <h4 className="font-semibold mb-3">AI Agent Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Provider</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.provider || 'openai'}
                  onChange={(e) => handleInputChange('provider', e.target.value)}
                >
                  <option value="openai">OpenAI</option>
                  <option value="anthropic">Anthropic</option>
                  <option value="google">Google</option>
                  <option value="local">Local Model</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Model</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="gpt-4, claude-3, etc."
                  value={config.model || ''}
                  onChange={(e) => handleInputChange('model', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Prompt</label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="What would you like the AI to do?"
                  rows="4"
                  value={config.prompt || ''}
                  onChange={(e) => handleInputChange('prompt', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Temperature</label>
                <input
                  type="number"
                  min="0"
                  max="1"
                  step="0.1"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="0.7"
                  value={config.temperature || 0.7}
                  onChange={(e) => handleInputChange('temperature', parseFloat(e.target.value))}
                />
              </div>
            </div>
          </div>
        );
      case 'notification':
        return (
          <div>
            <h4 className="font-semibold mb-3">Notification Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Channel</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.channel || 'email'}
                  onChange={(e) => handleInputChange('channel', e.target.value)}
                >
                  <option value="email">Email</option>
                  <option value="slack">Slack</option>
                  <option value="webhook">Webhook</option>
                  <option value="push">Push Notification</option>
                  <option value="sms">SMS</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Recipients</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="user@example.com, #channel, etc."
                  value={config.recipients || ''}
                  onChange={(e) => handleInputChange('recipients', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Subject/Title</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Notification Subject"
                  value={config.subject || ''}
                  onChange={(e) => handleInputChange('subject', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Message</label>
                <textarea
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus-500"
                  placeholder="Your notification message here..."
                  rows="4"
                  value={config.message || ''}
                  onChange={(e) => handleInputChange('message', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      default:
        return (
          <div>
            <h4 className="font-semibold mb-3">Properties</h4>
            <p className="text-sm text-gray-600">No specific configuration for this node type.</p>
            <div className="mt-3">
              <label className="block text-sm font-medium text-gray-700 mb-1">Label</label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Node label"
                value={selectedNode.data.label}
                onChange={(e) => {
                  if (onSave) {
                    onSave({
                      ...selectedNode,
                      data: {
                        ...selectedNode.data,
                        label: e.target.value
                      }
                    });
                  }
                }}
              />
            </div>
          </div>
        );
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-gray-200">
        <h3 className="text-lg font-semibold text-gray-800">
          {selectedNode?.data?.label || selectedNode?.type || 'Properties'} Properties
        </h3>
        <p className="text-sm text-gray-600 mt-1">
          Configure the selected node
        </p>
      </div>
      
      <div className="flex-1 p-4 overflow-y-auto">
        {renderPropertyForm()}
      </div>
      
      <div className="p-4 border-t border-gray-200 flex justify-end space-x-2">
        <button
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          onClick={handleCancel}
        >
          Cancel
        </button>
        <button
          className={`px-4 py-2 text-sm font-medium text-white rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 ${
            dirty 
              ? 'bg-blue-600 hover:bg-blue-700 focus:ring-blue-500' 
              : 'bg-gray-400 cursor-not-allowed'
          }`}
          onClick={handleSave}
          disabled={!dirty}
        >
          Save
        </button>
      </div>
    </div>
  );
};

export default PropertiesPanel;