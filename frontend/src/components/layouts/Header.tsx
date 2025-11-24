"use client"

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Bell,
  Search,
  Settings,
  User,
  HelpCircle,
  Play,
  Pause,
  Save,
  Undo,
  Redo,
  Copy,
  Trash2,
  Download,
  Upload,
  Menu
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import Link from "next/link";
import { useUIStore } from '@/stores/uiStore';

interface HeaderProps {
  title?: string;
  showActions?: boolean;
}

export function Header({ title = "FlowForge", showActions = false }: HeaderProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const { toggleMobileMenu } = useUIStore();

  return (
    <header className="flex items-center justify-between px-3 md:px-4 py-2 md:py-3 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      {/* Left Section */}
      <div className="flex items-center space-x-2 md:space-x-4 flex-1">
        {/* Mobile Menu Button */}
        <Button
          variant="ghost"
          size="sm"
          className="md:hidden h-9 w-9 p-0"
          onClick={toggleMobileMenu}
        >
          <Menu className="h-5 w-5" />
        </Button>

        <h1 className="text-lg md:text-xl font-semibold mr-2 md:mr-4">{title}</h1>

        {/* Search - Hidden on small mobile, icon on mobile, full on tablet+ */}
        <div className="hidden sm:flex relative max-w-xl flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search workflows, nodes, executions..."
            className="w-full pl-10 bg-muted/50 border-0 focus-visible:ring-1"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {/* Mobile Search Icon */}
        <Button variant="ghost" size="sm" className="sm:hidden h-9 w-9 p-0">
          <Search className="h-4 w-4" />
        </Button>
      </div>

      {/* Center Section - Workflow Actions - Hidden on mobile and tablet */}
      {showActions && (
        <div className="hidden lg:flex items-center space-x-2">
          <Button variant="outline" size="sm">
            <Undo className="w-4 h-4 mr-1" />
            Undo
          </Button>
          <Button variant="outline" size="sm">
            <Redo className="w-4 h-4 mr-1" />
            Redo
          </Button>
          <Separator orientation="vertical" className="h-6" />
          <Button variant="outline" size="sm">
            <Copy className="w-4 h-4 mr-1" />
            Copy
          </Button>
          <Button variant="outline" size="sm">
            <Trash2 className="w-4 h-4 mr-1" />
            Delete
          </Button>
          <Separator orientation="vertical" className="h-6" />
          <Button variant="outline" size="sm">
            <Upload className="w-4 h-4 mr-1" />
            Import
          </Button>
          <Button variant="outline" size="sm">
            <Download className="w-4 h-4 mr-1" />
            Export
          </Button>
          <Separator orientation="vertical" className="h-6" />
          <Button variant="default" size="sm">
            <Save className="w-4 h-4 mr-1" />
            Save
          </Button>
          <Button variant="default" size="sm">
            <Play className="w-4 h-4 mr-1" />
            Run
          </Button>
        </div>
      )}

      {/* Right Section */}
      <div className="flex items-center space-x-1 sm:space-x-2 md:space-x-3">
        {/* Notifications */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="relative h-9 w-9 p-0">
              <Bell className="w-4 h-4" />
              <span className="absolute -top-1 -right-1 w-2 h-2 bg-foreground rounded-full"></span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-80">
            <DropdownMenuLabel>Notifications</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <div className="flex items-start space-x-2">
                <div className="w-2 h-2 bg-foreground rounded-full mt-2"></div>
                <div>
                  <p className="text-sm font-medium">Workflow completed successfully</p>
                  <p className="text-xs text-muted-foreground">2 minutes ago</p>
                </div>
              </div>
            </DropdownMenuItem>
            <DropdownMenuItem>
              <div className="flex items-start space-x-2">
                <div className="w-2 h-2 bg-muted-foreground rounded-full mt-2"></div>
                <div>
                  <p className="text-sm font-medium">Node execution failed</p>
                  <p className="text-xs text-muted-foreground">5 minutes ago</p>
                </div>
              </div>
            </DropdownMenuItem>
            <DropdownMenuItem>
              <div className="flex items-start space-x-2">
                <div className="w-2 h-2 bg-foreground rounded-full mt-2"></div>
                <div>
                  <p className="text-sm font-medium">New workflow template available</p>
                  <p className="text-xs text-muted-foreground">1 hour ago</p>
                </div>
              </div>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Help - Hidden on small mobile */}
        <Button variant="ghost" size="sm" className="hidden sm:flex h-9 w-9 p-0">
          <HelpCircle className="w-4 h-4" />
        </Button>

        {/* Settings - Hidden on small mobile */}
        <Link href="/settings" className="hidden sm:inline-flex">
          <Button variant="ghost" size="sm" className="h-9 w-9 p-0">
            <Settings className="w-4 h-4" />
          </Button>
        </Link>

        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="flex items-center space-x-2 h-9">
              <div className="w-7 h-7 md:w-8 md:h-8 bg-primary rounded-full flex items-center justify-center">
                <User className="w-3 h-3 md:w-4 md:h-4 text-primary-foreground" />
              </div>
              <span className="hidden md:inline text-sm">John Doe</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Profile</DropdownMenuItem>
            <DropdownMenuItem>Settings</DropdownMenuItem>
            <DropdownMenuItem>Billing</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Sign out</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Theme Toggle */}
        <ThemeToggle />
      </div>
    </header>
  );
}