"use client"

import { useState, useEffect, useMemo, useCallback } from 'react';
import { useNodeStore } from '@/stores/nodeStore';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { Search, Filter, Loader2 } from 'lucide-react';
import { NodeType } from '@/types/workflow';
import { CategoryIcon } from '@/components/NodeIcon';
import { NodeCard } from './NodeCard';

const ITEMS_PER_PAGE = 20;

export function NodePalette() {
  const { nodeTypes, searchNodeTypes, getNodeTypesByCategory, fetchNodes, isLoading } = useNodeStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [visibleCount, setVisibleCount] = useState(ITEMS_PER_PAGE);

  useEffect(() => {
    fetchNodes();
  }, [fetchNodes]);

  // Reset visible count when filter changes
  useEffect(() => {
    setVisibleCount(ITEMS_PER_PAGE);
  }, [searchQuery, selectedCategory]);

  // Filter nodes based on search and category
  const filteredNodes = useMemo(() => {
    return searchQuery
      ? searchNodeTypes(searchQuery)
      : selectedCategory === 'all'
        ? nodeTypes
        : getNodeTypesByCategory(selectedCategory);
  }, [searchQuery, selectedCategory, nodeTypes, searchNodeTypes, getNodeTypesByCategory]);

  // Get visible nodes
  const visibleNodes = useMemo(() => {
    return filteredNodes.slice(0, visibleCount);
  }, [filteredNodes, visibleCount]);

  // Get unique categories
  const categories = useMemo(() => {
    return Array.from(new Set(nodeTypes.map(node => node.category)));
  }, [nodeTypes]);

  const handleDragStart = useCallback((event: React.DragEvent, nodeType: NodeType) => {
    event.dataTransfer.setData('application/reactflow', JSON.stringify({
      nodeType: nodeType.id,
      label: nodeType.name,
      description: nodeType.description,
      inputs: nodeType.inputs,
      outputs: nodeType.outputs,
      config: nodeType.config
    }));
    event.dataTransfer.effectAllowed = 'move';
  }, []);

  const handleLoadMore = () => {
    setVisibleCount(prev => prev + ITEMS_PER_PAGE);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b shrink-0">
        <h2 className="text-base font-semibold mb-2">Node Palette</h2>

        {/* Search */}
        <div className="relative mb-2">
          <Search className="absolute left-2.5 top-1/2 transform -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Search nodes..."
            className="pl-8 h-8 text-xs"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {/* Category Filter */}
        <div className="flex flex-wrap gap-1">
          <Button
            variant={selectedCategory === 'all' ? 'default' : 'outline'}
            size="sm"
            className="h-6 text-[10px] px-2"
            onClick={() => setSelectedCategory('all')}
          >
            All
          </Button>
          {categories.slice(0, 5).map((category) => (
            <Button
              key={category}
              variant={selectedCategory === category ? 'default' : 'outline'}
              size="sm"
              className="h-6 text-[10px] px-2 flex items-center gap-1"
              onClick={() => setSelectedCategory(category)}
            >
              <CategoryIcon category={category} size={10} useColor={false} />
              <span className="capitalize">{category}</span>
            </Button>
          ))}
        </div>
      </div>

      {/* Node List */}
      <ScrollArea className="flex-1 overflow-y-auto">
        <div className="p-2 space-y-2">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Loader2 className="h-6 w-6 animate-spin mb-2" />
              <p className="text-xs">Loading nodes...</p>
            </div>
          ) : filteredNodes.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Filter className="w-6 h-6 mx-auto mb-2" />
              <p className="text-xs">No nodes found</p>
              <p className="text-[10px] mb-3">Try adjusting your search or filter</p>
              <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => fetchNodes()}>
                Refresh Nodes
              </Button>
            </div>
          ) : (
            <>
              {visibleNodes.map((nodeType) => (
                <NodeCard
                  key={nodeType.id}
                  nodeType={nodeType}
                  onDragStart={handleDragStart}
                />
              ))}

              {visibleCount < filteredNodes.length && (
                <Button
                  variant="ghost"
                  className="w-full mt-2 h-7 text-xs"
                  onClick={handleLoadMore}
                >
                  Load More ({filteredNodes.length - visibleCount} remaining)
                </Button>
              )}
            </>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}