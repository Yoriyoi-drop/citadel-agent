// frontend/src/components/workflow-builder/nodes/security/APIKeyManagerNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const APIKeyManagerNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseSecurityNode data={data} color="#f59e0b">
      <div className="flex items-center mb-2">
        <span className="text-yellow-500 mr-2">🔑</span>
        <span className="font-medium">API Key Manager</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Operation: {config.operation || 'manage'}</div>
        <div className="truncate">Expiry: {config.default_expiry ? Math.floor(config.default_expiry / 1000 / 60 / 60 / 24) + ' days' : '30 days'}</div>
        <div className="truncate">Max: {config.max_keys_per_user || 'Unlimited'}</div>
      </div>
    </BaseSecurityNode>
  );
};

export default APIKeyManagerNode;