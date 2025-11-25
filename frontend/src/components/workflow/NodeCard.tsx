import React, { memo } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import { NodeType } from '@/types/workflow';
import { NodeIconBadge } from '@/components/NodeIcon';

interface NodeCardProps {
    nodeType: NodeType;
    onDragStart: (event: React.DragEvent, nodeType: NodeType) => void;
    onAdd?: (nodeType: NodeType) => void;
}

export const NodeCard = memo(({ nodeType, onDragStart, onAdd }: NodeCardProps) => {
    return (
        <Card
            className="cursor-grab active:cursor-grabbing hover:shadow-md transition-all duration-200 border-l-4 hover:border-l-4"
            style={{ borderLeftColor: nodeType.color || '#64748b' }}
            draggable
            onDragStart={(e) => onDragStart(e, nodeType)}
        >
            <CardHeader className="p-2 pb-0">
                <div className="flex items-start justify-between gap-2">
                    <div className="flex items-start gap-2 min-w-0 flex-1">
                        <NodeIconBadge
                            type={nodeType.id}
                            category={nodeType.category}
                            size={16}
                            className="shrink-0 mt-0.5"
                        />
                        <div className="min-w-0 flex-1">
                            <CardTitle className="text-xs font-semibold leading-tight line-clamp-1">
                                {nodeType.name}
                            </CardTitle>
                            <CardDescription className="text-[10px] mt-0.5 line-clamp-1">
                                {nodeType.description}
                            </CardDescription>
                        </div>
                    </div>
                    {onAdd && (
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-5 w-5 shrink-0 text-muted-foreground hover:text-primary"
                            onClick={() => onAdd(nodeType)}
                        >
                            <Plus className="h-3 w-3" />
                        </Button>
                    )}
                </div>
            </CardHeader>
            <CardContent className="p-2 pt-1">
                <div className="flex flex-wrap gap-1">
                    {nodeType.tags?.slice(0, 2).map((tag) => (
                        <Badge key={tag} variant="secondary" className="text-[9px] px-1 py-0 h-3.5 leading-none">
                            {tag}
                        </Badge>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
});

NodeCard.displayName = 'NodeCard';
