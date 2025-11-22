// frontend/src/components/workflow-builder/nodes/security/FirewallManagerNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const FirewallManagerNode = ({ data }) => {
  const config = data.config || {};
  const rules = config.rules || [];
  
  return (
    <BaseSecurityNode data={data} color="#ef4444">
      <div className="flex items-center mb-2">
        <span className="text-red-500 mr-2">🛡️</span>
        <span className="font-medium">Firewall Manager</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Rules: {rules.length || 0}</div>
        <div className="truncate">Default: {config.default_action || 'Deny'}</div>
        <div className="truncate">IPs: {config.whitelist_ips?.length || 0} whitelist, {config.blacklist_ips?.length || 0} blacklist</div>
      </div>
    </BaseSecurityNode>
  );
};

export default FirewallManagerNode;