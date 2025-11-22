// frontend/src/components/workflow-builder/nodes/security/EncryptionNode.jsx
import React from 'react';
import BaseSecurityNode from './BaseSecurityNode';

const EncryptionNode = ({ data }) => {
  const config = data.config || {};
  
  return (
    <BaseSecurityNode data={data} color="#8b5cf6">
      <div className="flex items-center mb-2">
        <span className="text-purple-500 mr-2">🔒</span>
        <span className="font-medium">Encryption/Decryption</span>
      </div>
      <div className="text-xs text-gray-600">
        <div className="mb-1">Operation: {config.operation || 'encrypt/decrypt'}</div>
        <div className="truncate">Algorithm: {config.algorithm || 'AES-256'}</div>
        <div className="truncate">Key: {config.encryption_key ? 'Set' : 'Not configured'}</div>
      </div>
    </BaseSecurityNode>
  );
};

export default EncryptionNode;