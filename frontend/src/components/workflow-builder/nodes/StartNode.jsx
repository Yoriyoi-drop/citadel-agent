// frontend/src/components/workflow-builder/nodes/StartNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const StartNode = ({ data }) => {
  return (
    <BaseNode data={data} color="#10b981">
      <div className="flex items-center">
        <span className="text-green-500 mr-2">▶️</span>
        <span>Start of workflow</span>
      </div>
    </BaseNode>
  );
};

export default StartNode;