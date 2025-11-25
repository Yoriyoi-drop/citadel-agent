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
            className="cursor-grab active:cursor-grabbing hover:shadow-md transition-all duration-200 border-l-4"
            style={{ borderLeftColor: nodeType.color || '#64748b' }}
            draggable
            onDragStart={(e) => onDragStart(e, nodeType)}
        >
            <CardHeader className="p-3 pb-0">
                <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                        <NodeIconBadge
                            type={nodeType.id}
                            category={nodeType.category}
                            size={18}
                            className="shrink-0"
                        />
                        <div>
                            <CardTitle className="text-sm font-medium leading-none">
                                {nodeType.name}
                            </CardTitle>
                            <CardDescription className="text-xs mt-1 line-clamp-2">
                                {nodeType.description}
                            </CardDescription>
                        </div>
                    </div>
                    {onAdd && (
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 -mr-1 -mt-1 text-muted-foreground hover:text-primary"
                            onClick={() => onAdd(nodeType)}
                        >
                            <Plus className="h-4 w-4" />
                        </Button>
                    )}
                </div>
            </CardHeader>
            <CardContent className="p-3 pt-2">
                <div className="flex flex-wrap gap-1">
                    {nodeType.tags?.slice(0, 3).map((tag) => (
                        <Badge key={tag} variant="secondary" className="text-[10px] px-1 h-4">
                            {tag}
                        </Badge>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
});

NodeCard.displayName = 'NodeCard';
