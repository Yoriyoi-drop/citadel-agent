// frontend/src/components/workflow-builder/nodes/DelayNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const DelayNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseNode data={data} color="#ec4899">
      <div className="flex items-center">
        <span className="text-pink-500 mr-2">⏱️</span>
        <span className="font-medium">Delay</span>
      </div>
      <div className="mt-1 text-xs text-gray-500">
        <div>Duration: {config.duration || 'Not set'} seconds</div>
      </div>
    </BaseNode>
  );
};

export default DelayNode;