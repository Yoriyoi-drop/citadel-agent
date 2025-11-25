"use client"

import { memo } from 'react';
import { BaseEdge, getBezierPath, ConnectionLineComponentProps } from '@xyflow/react';

const ConnectionLineComponent = memo(({
  fromX,
  fromY,
  toX,
  toY,
  fromPosition,
  toPosition,
  connectionLineStyle,
}: ConnectionLineComponentProps) => {
  const [edgePath] = getBezierPath({
    sourceX: fromX,
    sourceY: fromY,
    sourcePosition: fromPosition,
    targetX: toX,
    targetY: toY,
    targetPosition: toPosition,
  });

  return (
    <BaseEdge
      path={edgePath}
      style={{
        ...connectionLineStyle,
        strokeWidth: 2,
        stroke: '#3b82f6',
        strokeDasharray: '5,5',
      }}
    />
  );
});

ConnectionLineComponent.displayName = 'ConnectionLineComponent';

export default ConnectionLineComponent;