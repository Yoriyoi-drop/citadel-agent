import { memo } from 'react';
import { BaseEdge, EdgeProps, getBezierPath } from '@xyflow/react';

const CustomEdge = memo(({
    id,
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    style = {},
    markerEnd,
    selected,
}: EdgeProps) => {
    const [edgePath] = getBezierPath({
        sourceX,
        sourceY,
        sourcePosition,
        targetX,
        targetY,
        targetPosition,
    });

    return (
        <g>
            {/* Glow effect when selected */}
            {selected && (
                <path
                    d={edgePath}
                    fill="none"
                    stroke="hsl(var(--primary))"
                    strokeWidth={8}
                    opacity={0.15}
                    className="pointer-events-none"
                />
            )}

            {/* Main edge path */}
            <BaseEdge
                id={id}
                path={edgePath}
                markerEnd={markerEnd}
                style={{
                    ...style,
                    strokeWidth: selected ? 2.5 : 2,
                    stroke: selected ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground))',
                    transition: 'all 0.2s ease',
                }}
            />
        </g>
    );
});

CustomEdge.displayName = 'CustomEdge';

export default CustomEdge;
