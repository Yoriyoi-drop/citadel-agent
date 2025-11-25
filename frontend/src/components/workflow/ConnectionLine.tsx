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
    <g>
      {/* Glow effect */}
      <path
        d={edgePath}
        fill="none"
        stroke="#3b82f6"
        strokeWidth={6}
        opacity={0.2}
        className="pointer-events-none"
      />
      {/* Main line */}
      <BaseEdge
        path={edgePath}
        style={{
          ...connectionLineStyle,
          strokeWidth: 2.5,
          stroke: '#3b82f6',
          strokeDasharray: '8,4',
          animation: 'dash 0.5s linear infinite',
        }}
      />
    </g>
  );
});

ConnectionLineComponent.displayName = 'ConnectionLineComponent';

export default ConnectionLineComponent;