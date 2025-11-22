// frontend/src/components/workflow-builder/nodes/security/AccessControlNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const AccessControlNode = ({ data }) => {
  const config = data.config || {};
  const roles = config.roles || [];
  
  return (
    <BaseSecurityNode data={data} color="#10b981">
      <div className="flex items-center mb-2">
        <span className="text-green-500 mr-2">🔐</span>
        <span className="font-medium">Access Control</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Control Type: {config.control_type || 'rbac'}</div>
        <div className="truncate">Roles: {roles.length || 0}</div>
        <div className="truncate">Policy: {config.default_allow ? 'Allow by default' : 'Deny by default'}</div>
      </div>
    </BaseSecurityNode>
  );
};

export default AccessControlNode;