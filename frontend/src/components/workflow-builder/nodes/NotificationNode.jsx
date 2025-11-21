// frontend/src/components/workflow-builder/nodes/NotificationNode.jsx
import React from 'react';
import BaseNode from './BaseNode';

const NotificationNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseNode data={data} color="#06b6d4">
      <div className="flex items-center">
        <span className="text-cyan-500 mr-2">🔔</span>
        <span className="font-medium">Notification</span>
      </div>
      <div className="mt-1 text-xs text-gray-500">
        <div>Channel: {config.channel || 'email'}</div>
        <div>Recipient: {config.recipient || 'Not set'}</div>
      </div>
    </BaseNode>
  );
};

export default NotificationNode;