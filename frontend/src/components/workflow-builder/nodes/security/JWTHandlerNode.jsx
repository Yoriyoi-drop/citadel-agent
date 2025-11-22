// frontend/src/components/workflow-builder/nodes/security/JWTHandlerNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const JWTHandlerNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseSecurityNode data={data} color="#ec4899">
      <div className="flex items-center mb-2">
        <span className="text-pink-500 mr-2">.JWT</span>
        <span className="font-medium">JWT Handler</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Operation: {config.operation || 'token'}</div>
        <div className="truncate">Algorithm: {config.algorithm || 'HS256'}</div>
        <div className="truncate">Issuer: {config.issuer || 'Not set'}</div>
      </div>
    </BaseSecurityNode>
  );
};

export default JWTHandlerNode;