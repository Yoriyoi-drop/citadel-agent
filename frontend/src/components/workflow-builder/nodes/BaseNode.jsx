// frontend/src/components/workflow-builder/nodes/BaseNode.jsx
import React from 'react';
import { Handle, Position } from 'reactflow';

const BaseNode = ({ data, children, color = '#3b82f6' }) => {
  return (
    <div className={`px-4 py-3 min-w-[200px] rounded-lg shadow-md border-2 border-${color.replace('#', '')} bg-white`} style={{borderColor: color}}>
      <Handle type="target" position={Position.Top} />
      <div className="text-sm font-semibold text-gray-800 mb-2">{data.label}</div>
      <div className="text-xs text-gray-600">
        {children}
      </div>
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
};

export default BaseNode;