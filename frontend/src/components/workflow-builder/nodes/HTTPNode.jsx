// frontend/src/components/workflow-builder/nodes/HTTPNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const HTTPNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseNode data={data} color="#3b82f6">
      <div className="flex items-center">
        <span className="text-blue-500 mr-2">🌐</span>
        <span className="font-medium">HTTP Request</span>
      </div>
      <div className="mt-1 text-xs text-gray-500">
        <div>Method: {config.method || 'GET'}</div>
        <div>URL: {config.url || 'Not set'}</div>
      </div>
    </BaseNode>
  );
};

export default HTTPNode;