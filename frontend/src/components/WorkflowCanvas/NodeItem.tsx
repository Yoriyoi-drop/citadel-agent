import { Handle, Position, NodeProps } from '@xyflow/react';
import { NodeData } from '../../types';

const NodeItem: React.FC<NodeProps> = ({ data, isConnectable }) => {
  const nodeData = data as unknown as NodeData;
  // Determine node color based on type
  const getNodeColor = (type: string) => {
    switch (type) {
      case 'trigger':
        return 'bg-blue-500';
      case 'action':
        return 'bg-green-500';
      case 'condition':
        return 'bg-yellow-500';
      case 'loop':
        return 'bg-purple-500';
      default:
        return 'bg-gray-500';
    }
  };

  return (
    <div className={`px-4 py-2 rounded shadow-md min-w-[200px] ${getNodeColor(nodeData.type)} text-white`}>
      <Handle
        type="target"
        position={Position.Left}
        isConnectable={isConnectable}
        className="w-3 h-3 bg-white"
      />

      <div className="text-center">
        <div className="font-bold">{nodeData.label}</div>
        <div className="text-xs mt-1 opacity-80">{nodeData.description}</div>
      </div>

      <Handle
        type="source"
        position={Position.Right}
        isConnectable={isConnectable}
        className="w-3 h-3 bg-white"
      />
    </div>
  );
};

export default NodeItem;