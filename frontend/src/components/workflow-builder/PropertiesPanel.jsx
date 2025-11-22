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
      case 'firewallManager':
        return (
          <div>
            <h4 className="font-semibold mb-3">Firewall Manager Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operation</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operation || 'check'}
                  onChange={(e) => handleInputChange('operation', e.target.value)}
                >
                  <option value="check">Check Access</option>
                  <option value="validate">Validate IP</option>
                  <option value="filter">Filter Request</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Default Action</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.default_action || 'block'}
                  onChange={(e) => handleInputChange('default_action', e.target.value)}
                >
                  <option value="allow">Allow</option>
                  <option value="block">Block</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Whitelist IPs (comma separated)</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="192.168.1.1, 10.0.0.0/24"
                  value={config.whitelist_ips?.join(', ') || ''}
                  onChange={(e) => handleInputChange('whitelist_ips', e.target.value.split(',').map(ip => ip.trim()))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Blacklist IPs (comma separated)</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="192.168.1.10, 203.0.113.0/24"
                  value={config.blacklist_ips?.join(', ') || ''}
                  onChange={(e) => handleInputChange('blacklist_ips', e.target.value.split(',').map(ip => ip.trim()))}
                />
              </div>
            </div>
          </div>
        );
      case 'encryption':
        return (
          <div>
            <h4 className="font-semibold mb-3">Encryption/Decryption Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operation</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operation || 'encrypt'}
                  onChange={(e) => handleInputChange('operation', e.target.value)}
                >
                  <option value="encrypt">Encrypt</option>
                  <option value="decrypt">Decrypt</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Algorithm</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.algorithm || 'AES256'}
                  onChange={(e) => handleInputChange('algorithm', e.target.value)}
                >
                  <option value="AES256">AES-256</option>
                  <option value="AES128">AES-128</option>
                  <option value="RSA">RSA</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Encryption Key</label>
                <input
                  type="password"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter encryption key"
                  value={config.encryption_key || ''}
                  onChange={(e) => handleInputChange('encryption_key', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'accessControl':
        return (
          <div>
            <h4 className="font-semibold mb-3">Access Control Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Control Type</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.control_type || 'rbac'}
                  onChange={(e) => handleInputChange('control_type', e.target.value)}
                >
                  <option value="rbac">Role-Based Access Control (RBAC)</option>
                  <option value="ldap">LDAP Authentication</option>
                  <option value="ad">Active Directory</option>
                  <option value="custom">Custom Access Control</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Default Allow</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.default_allow ? 'true' : 'false'}
                  onChange={(e) => handleInputChange('default_allow', e.target.value === 'true')}
                >
                  <option value="true">Allow by Default</option>
                  <option value="false">Deny by Default</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">LDAP Server</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="ldap://ldap.example.com:389"
                  value={config.ldap_server || ''}
                  onChange={(e) => handleInputChange('ldap_server', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'apiKeyManager':
        return (
          <div>
            <h4 className="font-semibold mb-3">API Key Manager Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operation</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operation || 'create'}
                  onChange={(e) => handleInputChange('operation', e.target.value)}
                >
                  <option value="create">Create Key</option>
                  <option value="validate">Validate Key</option>
                  <option value="revoke">Revoke Key</option>
                  <option value="list">List Keys</option>
                  <option value="rotate">Rotate Key</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Default Expiry (hours)</label>
                <input
                  type="number"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="24"
                  value={config.default_expiry ? Math.floor(config.default_expiry / 1000 / 60 / 60) : 24}
                  onChange={(e) => handleInputChange('default_expiry', parseInt(e.target.value) * 1000 * 60 * 60)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Max Keys Per User</label>
                <input
                  type="number"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="5"
                  value={config.max_keys_per_user || ''}
                  onChange={(e) => handleInputChange('max_keys_per_user', parseInt(e.target.value))}
                />
              </div>
            </div>
          </div>
        );
      case 'jwtHandler':
        return (
          <div>
            <h4 className="font-semibold mb-3">JWT Handler Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operation</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operation || 'create'}
                  onChange={(e) => handleInputChange('operation', e.target.value)}
                >
                  <option value="create">Create Token</option>
                  <option value="validate">Validate Token</option>
                  <option value="refresh">Refresh Token</option>
                  <option value="decode">Decode Token</option>
                  <option value="revoke">Revoke Token</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Algorithm</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.algorithm || 'HS256'}
                  onChange={(e) => handleInputChange('algorithm', e.target.value)}
                >
                  <option value="HS256">HS256 (HMAC)</option>
                  <option value="HS384">HS384 (HMAC)</option>
                  <option value="HS512">HS512 (HMAC)</option>
                  <option value="RS256">RS256 (RSA)</option>
                  <option value="RS384">RS384 (RSA)</option>
                  <option value="RS512">RS512 (RSA)</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Issuer</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="your-app-name"
                  value={config.issuer || ''}
                  onChange={(e) => handleInputChange('issuer', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Secret Key</label>
                <input
                  type="password"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter secret key"
                  value={config.secret_key || ''}
                  onChange={(e) => handleInputChange('secret_key', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'oauth2Provider':
        return (
          <div>
            <h4 className="font-semibold mb-3">OAuth2 Provider Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operation</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operation || 'authorize'}
                  onChange={(e) => handleInputChange('operation', e.target.value)}
                >
                  <option value="authorize">Authorize</option>
                  <option value="token">Token</option>
                  <option value="refresh">Refresh Token</option>
                  <option value="user_info">User Info</option>
                  <option value="revoke">Revoke Token</option>
                  <option value="introspect">Introspect Token</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Enable PKCE</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.enable_pkce ? 'true' : 'false'}
                  onChange={(e) => handleInputChange('enable_pkce', e.target.value === 'true')}
                >
                  <option value="true">Yes</option>
                  <option value="false">No</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Issuer URL</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="https://auth.example.com"
                  value={config.issuer || ''}
                  onChange={(e) => handleInputChange('issuer', e.target.value)}
                />
              </div>
            </div>
          </div>
        );
      case 'securityOperation':
        return (
          <div>
            <h4 className="font-semibold mb-3">Security Operations Node</h4>
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Operation</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.operation || 'hash'}
                  onChange={(e) => handleInputChange('operation', e.target.value)}
                >
                  <option value="hash">Hash Data</option>
                  <option value="encrypt">Encrypt Data</option>
                  <option value="decrypt">Decrypt Data</option>
                  <option value="sign">Sign Data</option>
                  <option value="verify">Verify Signature</option>
                  <option value="validate">Validate Data</option>
                  <option value="mask">Mask Data</option>
                  <option value="generate_token">Generate Token</option>
                  <option value="validate_token">Validate Token</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Algorithm</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  value={config.algorithm || 'SHA256'}
                  onChange={(e) => handleInputChange('algorithm', e.target.value)}
                >
                  <option value="SHA256">SHA-256</option>
                  <option value="SHA512">SHA-512</option>
                  <option value="AES256">AES-256</option>
                  <option value="HMACSHA256">HMAC-SHA256</option>
                  <option value="BCrypt">BCrypt</option>
                  <option value="Scrypt">Scrypt</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Secret Key</label>
                <input
                  type="password"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Enter secret key if required"
                  value={config.secret_key || ''}
                  onChange={(e) => handleInputChange('secret_key', e.target.value)}
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