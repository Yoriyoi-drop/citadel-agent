"use client"

import { useState, useEffect } from 'react';
import { useNodeStore } from '@/stores/nodeStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import { Search, Filter, Loader2 } from 'lucide-react';
import { NodeType } from '@/types/workflow';
import { CategoryIcon, NodeIconBadge } from '@/components/NodeIcon';
import { getCategoryColor } from '@/config/nodeIcons';



export function NodePalette() {
  const { nodeTypes, searchNodeTypes, getNodeTypesByCategory, fetchNodes, isLoading } = useNodeStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  useEffect(() => {
    fetchNodes();
  }, [fetchNodes]);

  // Filter nodes based on search and category
  const filteredNodes = searchQuery
    ? searchNodeTypes(searchQuery)
    : selectedCategory === 'all'
      ? nodeTypes
      : getNodeTypesByCategory(selectedCategory);

  // Get unique categories
  const categories = Array.from(new Set(nodeTypes.map(node => node.category)));

  const handleDragStart = (event: React.DragEvent, nodeType: NodeType) => {
    event.dataTransfer.setData('application/reactflow', JSON.stringify({
      nodeType: nodeType.id,
      label: nodeType.name,
      description: nodeType.description,
      inputs: nodeType.inputs,
      outputs: nodeType.outputs,
      config: nodeType.config
    }));
    event.dataTransfer.effectAllowed = 'move';
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b">
        <h2 className="text-lg font-semibold mb-4">Node Palette</h2>

        {/* Search */}
        <div className="relative mb-4">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search nodes..."
            className="pl-10"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {/* Category Filter */}
        <div className="flex flex-wrap gap-2">
          <Button
            variant={selectedCategory === 'all' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSelectedCategory('all')}
          >
            All ({nodeTypes.length})
          </Button>
          {categories.map((category) => {
            const count = nodeTypes.filter(n => n.category === category).length;

            return (
              <Button
                key={category}
                variant={selectedCategory === category ? 'default' : 'outline'}
                size="sm"
                onClick={() => setSelectedCategory(category)}
                className="flex items-center space-x-1"
              >
                <CategoryIcon category={category} size={14} useColor={false} />
                <span className="capitalize">{category}</span>
                <span className="text-xs">({count})</span>
              </Button>
            );
          })}
        </div>
      </div>

      {/* Node List */}
      <ScrollArea className="flex-1">
        <div className="p-4 space-y-3">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Loader2 className="h-8 w-8 animate-spin mb-2" />
              <p>Loading nodes...</p>
            </div>
          ) : filteredNodes.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No nodes found</p>
            </div>
          ) : (
            filteredNodes.map((nodeType) => {
              return (
                <Card
                  key={nodeType.id}
                  className="cursor-grab active:cursor-grabbing hover:shadow-md transition-shadow"
                  draggable
                  onDragStart={(e) => handleDragStart(e, nodeType)}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center space-x-3">
                        <NodeIconBadge
                          type={nodeType.id}
                          category={nodeType.category}
                          size={20}
                          showBackground={true}
                        />
                        <CardTitle className="text-sm">{nodeType.name}</CardTitle>
                      </div>
                      <Badge
                        variant="outline"
                        className="text-xs"
                        style={{
                          borderColor: getCategoryColor(nodeType.category),
                          color: getCategoryColor(nodeType.category)
                        }}
                      >
                        {nodeType.category}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <CardDescription className="text-xs mb-3">
                      {nodeType.description}
                    </CardDescription>

                    {/* Inputs/Outputs */}
                    <div className="space-y-2">
                      {nodeType.inputs.length > 0 && (
                        <div className="flex items-center space-x-2">
                          <span className="text-xs text-muted-foreground">Inputs:</span>
                          <div className="flex space-x-1">
                            {nodeType.inputs.slice(0, 3).map((input, index) => (
                              <Badge key={index} variant="secondary" className="text-xs">
                                {input.name}
                              </Badge>
                            ))}
                            {nodeType.inputs.length > 3 && (
                              <Badge variant="secondary" className="text-xs">
                                +{nodeType.inputs.length - 3}
                              </Badge>
                            )}
                          </div>
                        </div>
                      )}

                      {nodeType.outputs.length > 0 && (
                        <div className="flex items-center space-x-2">
                          <span className="text-xs text-muted-foreground">Outputs:</span>
                          <div className="flex space-x-1">
                            {nodeType.outputs.slice(0, 3).map((output, index) => (
                              <Badge key={index} variant="secondary" className="text-xs">
                                {output.name}
                              </Badge>
                            ))}
                            {nodeType.outputs.length > 3 && (
                              <Badge variant="secondary" className="text-xs">
                                +{nodeType.outputs.length - 3}
                              </Badge>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              );
            })}

          {filteredNodes.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <Filter className="w-8 h-8 mx-auto mb-2" />
              <p>No nodes found</p>
              <p className="text-sm">Try adjusting your search or filter</p>
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}