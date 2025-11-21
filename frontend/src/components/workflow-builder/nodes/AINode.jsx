// frontend/src/components/workflow-builder/nodes/AINode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const AINode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseNode data={data} color="#6366f1">
      <div className="flex items-center">
        <span className="text-indigo-500 mr-2">🤖</span>
        <span className="font-medium">AI Agent</span>
      </div>
      <div className="mt-1 text-xs text-gray-500">
        <div>Model: {config.model || 'Not set'}</div>
        <div>Provider: {config.provider || 'OpenAI'}</div>
      </div>
    </BaseNode>
  );
};

export default AINode;