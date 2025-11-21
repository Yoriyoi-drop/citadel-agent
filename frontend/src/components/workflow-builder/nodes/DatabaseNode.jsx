// frontend/src/components/workflow-builder/nodes/DatabaseNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const DatabaseNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseNode data={data} color="#8b5cf6">
      <div className="flex items-center">
        <span className="text-purple-500 mr-2">🗄️</span>
        <span className="font-medium">Database Query</span>
      </div>
      <div className="mt-1 text-xs text-gray-500">
        <div>Type: {config.databaseType || 'Unknown'}</div>
        <div>Query: {config.query ? config.query.substring(0, 30) + '...' : 'Not set'}</div>
      </div>
    </BaseNode>
  );
};

export default DatabaseNode;