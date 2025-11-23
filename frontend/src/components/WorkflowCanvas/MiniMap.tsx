// MiniMap component wrapper for ReactFlow
import React from 'react';
import { MiniMap as ReactFlowMiniMap } from '@xyflow/react';

const MiniMapComponent: React.FC = () => {
    return (
        <ReactFlowMiniMap
            nodeColor={(node) => {
                switch (node.type) {
                    case 'trigger':
                        return '#3b82f6';
                    case 'action':
                        return '#10b981';
                    case 'condition':
                        return '#eab308';
                    case 'loop':
                        return '#a855f7';
                    default:
                        return '#6b7280';
                }
            }}
            nodeStrokeWidth={3}
            zoomable
            pannable
            position="bottom-left"
        />
    );
};

export default MiniMapComponent;
