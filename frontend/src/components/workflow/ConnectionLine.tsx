"use client"

import { memo } from 'react';
import { BaseEdge, EdgeLabelRenderer, getBezierPath } from '@xyflow/react';
import { Badge } from '@/components/ui/badge';

interface ConnectionLineProps {
  id: string;
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: any;
  targetPosition: any;
  style?: React.CSSProperties;
  markerEnd?: string;
  data?: {
    label?: string;
    type?: string;
    animated?: boolean;
  };
}

const ConnectionLineComponent = memo(({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  data
}: ConnectionLineProps) => {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          ...style,
          strokeWidth: 2,
          stroke: data?.type === 'error' ? '#ef4444' : '#3b82f6',
          strokeDasharray: data?.type === 'async' ? '5,5' : undefined,
        }}
      />
      
      {data?.label && (
        <EdgeLabelRenderer>
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              pointerEvents: 'all',
            }}
            className="nodrag nopan"
          >
            <Badge variant="secondary" className="text-xs">
              {data.label}
            </Badge>
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
});

ConnectionLineComponent.displayName = 'ConnectionLineComponent';

export default ConnectionLineComponent;