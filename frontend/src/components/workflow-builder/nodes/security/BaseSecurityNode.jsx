// frontend/src/components/workflow-builder/nodes/security/BaseSecurityNode.jsx
import React from 'react';
import { Handle, Position } from 'reactflow';

const BaseSecurityNode = ({ data, children, color = '#f59e0b' }) => {
  return (
    <div 
      className="px-4 py-3 min-w-[220px] rounded-xl shadow-lg border-2 bg-white flex flex-col"
      style={{ borderColor: color, minHeight: '100px' }}
    >
      <div className="flex items-center justify-between mb-2">
        <Handle 
          type="target" 
          position={Position.Left} 
          className="w-3 h-3 bg-gray-300"
          style={{ borderColor: color }}
        />
        <div className="flex-1 text-center">
          <div className="text-sm font-semibold text-gray-800">{data.label}</div>
          <div className="text-xs text-gray-500 mt-1">{data.description}</div>
        </div>
        <Handle 
          type="source" 
          position={Position.Right} 
          className="w-3 h-3 bg-gray-300"
          style={{ borderColor: color }}
        />
      </div>
      <div className="flex-1">
        {children}
      </div>
    </div>
  );
};

export default BaseSecurityNode;