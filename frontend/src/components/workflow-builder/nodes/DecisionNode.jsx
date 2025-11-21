// frontend/src/components/workflow-builder/nodes/DecisionNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const DecisionNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseNode data={data} color="#f59e0b">
      <div className="flex items-center">
        <span className="text-yellow-500 mr-2">❓</span>
        <span className="font-medium">Conditional Logic</span>
      </div>
      <div className="mt-1 text-xs text-gray-500">
        <div>Condition: {config.condition || 'Not set'}</div>
        <div>Operator: {config.operator || '='}</div>
      </div>
    </BaseNode>
  );
};

export default DecisionNode;