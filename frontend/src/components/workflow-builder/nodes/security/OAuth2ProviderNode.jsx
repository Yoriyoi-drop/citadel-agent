// frontend/src/components/workflow-builder/nodes/security/OAuth2ProviderNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const OAuth2ProviderNode = ({ data }) => {
  const config = data.config || {};
  const clients = config.clients || [];
  
  return (
    <BaseSecurityNode data={data} color="#3b82f6">
      <div className="flex items-center mb-2">
        <span className="text-blue-500 mr-2">OAuth</span>
        <span className="font-medium">OAuth2 Provider</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Operation: {config.operation || 'authorize'}</div>
        <div className="truncate">Clients: {clients.length || 0}</div>
        <div className="truncate">PKCE: {config.enable_pkce ? 'Enabled' : 'Disabled'}</div>
      </div>
    </BaseSecurityNode>
  );
};

export default OAuth2ProviderNode;