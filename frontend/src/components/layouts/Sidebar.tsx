"use client"

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { useUIStore } from '@/stores/uiStore';
import {
  LayoutDashboard,
  GitBranch,
  PlayCircle,
  Box,
  Settings,
  Package,
  Store,
  BarChart3,
  ChevronLeft,
  ChevronRight,
  Plus,
  Search,
  Zap,
  Database,
  Brain,
  Globe,
  MessageSquare,
  FileText,
  Clock,
  X
} from 'lucide-react';

interface SidebarProps {
  collapsed: boolean;
}

const navigation = [
  {
    name: 'Dashboard',
    href: '/',
    icon: LayoutDashboard,
    badge: null
  },
  {
    name: 'Workflows',
    href: '/workflows',
    icon: GitBranch,
    badge: '12'
  },
  {
    name: 'Executions',
    href: '/executions',
    icon: PlayCircle,
    badge: null
  },
  {
    name: 'Nodes',
    href: '/nodes',
    icon: Box,
    badge: '45+'
  },
  {
    name: 'Templates',
    href: '/templates',
    icon: Package,
    badge: 'New'
  },
  {
    name: 'Marketplace',
    href: '/marketplace',
    icon: Store,
    badge: null
  },
  {
    name: 'Analytics',
    href: '/analytics',
    icon: BarChart3,
    badge: null
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
    badge: null
  }
];

const nodeCategories = [
  { name: 'AI', icon: Brain, count: 8 },
  { name: 'Database', icon: Database, count: 12 },
  { name: 'Communication', icon: MessageSquare, count: 15 },
  { name: 'Utility', icon: Clock, count: 10 }
];

export function Sidebar({ collapsed }: SidebarProps) {
  const pathname = usePathname();
  const [searchQuery, setSearchQuery] = useState('');
  const { setMobileMenuOpen } = useUIStore();

  return (
    <div className={cn(
      "flex flex-col bg-card transition-all duration-300 h-screen shadow-elegant-md",
      collapsed ? "w-16" : "w-64 md:w-56 lg:w-64"
    )}>
      {/* Logo */}
      <div className="flex items-center justify-between px-3 py-3">
        {!collapsed && (
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center">
              <Zap className="w-5 h-5" />
            </div>
            <span className="font-semibold text-lg">FlowForge</span>
          </div>
        )}

        {/* Close button for mobile */}
        <Button
          variant="ghost"
          size="sm"
          className="md:hidden ml-auto h-8 w-8 p-0"
          onClick={() => setMobileMenuOpen(false)}
        >
          <X className="w-4 h-4" />
        </Button>

        {/* Collapse button for desktop */}
        <Button
          variant="ghost"
          size="sm"
          className="hidden md:flex ml-auto h-8 w-8 p-0"
          onClick={() => {/* Toggle sidebar handled by parent */ }}
        >
          {collapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
        </Button>
      </div>

      {/* Quick Actions */}
      {!collapsed && (
        <div className="px-3 py-2 space-y-2">
          <Link href="/workflows/new" className="w-full">
            <Button className="w-full justify-start" size="sm">
              <Plus className="w-4 h-4 mr-2" />
              New Workflow
            </Button>
          </Link>
          <div className="relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <input
              placeholder="Search..."
              className="w-full pl-8 pr-3 py-2 text-sm bg-muted rounded-md border-0 focus:outline-none focus:ring-2 focus:ring-primary"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      )}

      <ScrollArea className="flex-1">
        {/* Main Navigation */}
        <div className="px-2 py-2">
          {navigation.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link key={item.name} href={item.href}>
                <Button
                  variant={isActive ? "secondary" : "ghost"}
                  className={cn(
                    "w-full justify-start mb-1",
                    collapsed && "justify-center px-2"
                  )}
                >
                  <item.icon className="w-4 h-4" />
                  {!collapsed && (
                    <>
                      <span className="ml-2">{item.name}</span>
                      {item.badge && (
                        <Badge variant="secondary" className="ml-auto text-xs">
                          {item.badge}
                        </Badge>
                      )}
                    </>
                  )}
                </Button>
              </Link>
            );
          })}
        </div>

        {/* Node Categories */}
        {!collapsed && (
          <>
            <Separator className="my-4" />
            <div className="px-4">
              <h3 className="text-sm font-medium text-muted-foreground mb-3">Node Categories</h3>
              <div className="space-y-2">
                {nodeCategories.map((category) => (
                  <div key={category.name} className="flex items-center justify-between p-2 rounded-lg hover:bg-muted cursor-pointer">
                    <div className="flex items-center space-x-2">
                      <category.icon className="w-4 h-4 text-muted-foreground" />
                      <span className="text-sm">{category.name}</span>
                    </div>
                    <Badge variant="outline" className="text-xs">
                      {category.count}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
      </ScrollArea>

      {/* Footer */}
      {!collapsed && (
        <div className="p-4">
          <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            <Globe className="w-4 h-4" />
            <span>v2.0.1</span>
          </div>
        </div>
      )}
    </div>
  );
}