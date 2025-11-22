// frontend/src/components/workflow-builder/nodes/security/SecurityOperationNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const SecurityOperationNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseSecurityNode data={data} color="#6366f1">
      <div className="flex items-center mb-2">
        <span className="text-indigo-500 mr-2">🛡️</span>
        <span className="font-medium">Security Operations</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Operation: {config.operation || 'process'}</div>
        <div className="truncate">Algorithm: {config.algorithm || 'SHA-256'}</div>
        <div className="truncate">Action: {config.operation || 'hash/encrypt/validate'}</div>
      </div>
    </BaseSecurityNode>
  );
};

export default SecurityOperationNode;