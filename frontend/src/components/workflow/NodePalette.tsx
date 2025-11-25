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
            All
          </Button>
          {categories.map((category) => {
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
              <Filter className="w-8 h-8 mx-auto mb-2" />
              <p>No nodes found</p>
              <p className="text-sm mb-4">Try adjusting your search or filter</p>
              <Button variant="outline" size="sm" onClick={() => fetchNodes()}>
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
                  className="w-full mt-2"
                  onClick={handleLoadMore}
                >
                  Load More ({filteredNodes.length - visibleCount} remaining)
                </Button>
              )}
            </>
          )}
        </div>
      </ScrollArea >
    </div >
  );
}