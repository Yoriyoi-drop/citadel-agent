// frontend/src/components/workflow-builder/nodes/EndNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const EndNode = ({ data }) => {
  return (
    <BaseNode data={data} color="#ef4444">
      <div className="flex items-center">
        <span className="text-red-500 mr-2">⏹️</span>
        <span>End of workflow</span>
      </div>
    </BaseNode>
  );
};

export default EndNode;