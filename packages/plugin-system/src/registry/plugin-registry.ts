// packages/plugin-system/src/registry/plugin-registry.ts
import { PluginMetadata, PluginType, PluginStatus } from '../types/plugin';

export interface PluginRegistryEntry {
  id: string;
  metadata: PluginMetadata;
  version: string;
  type: PluginType;
  status: PluginStatus;
  installPath: string;
  dependencies: string[];
  installedAt: Date;
  lastUpdated: Date;
}

export class PluginRegistry {
  private static instance: PluginRegistry;
  private plugins: Map<string, PluginRegistryEntry> = new Map();
  private pluginSearchIndex: Map<string, string[]> = new Map(); // keyword -> [pluginIds]

  private constructor() {}

  static getInstance(): PluginRegistry {
    if (!PluginRegistry.instance) {
      PluginRegistry.instance = new PluginRegistry();
    }
    return PluginRegistry.instance;
  }

  registerPlugin(entry: PluginRegistryEntry): boolean {
    if (this.plugins.has(entry.id)) {
      console.warn(`Plugin ${entry.id} already registered`);
      return false;
    }

    this.plugins.set(entry.id, entry);
    
    // Update search index
    this.updateSearchIndex(entry);
    
    return true;
  }

  unregisterPlugin(pluginId: string): boolean {
    if (!this.plugins.has(pluginId)) {
      return false;
    }

    const entry = this.plugins.get(pluginId)!;
    
    // Remove from search index
    this.removeFromSearchIndex(entry);
    
    return this.plugins.delete(pluginId);
  }

  getPlugin(pluginId: string): PluginRegistryEntry | undefined {
    return this.plugins.get(pluginId);
  }

  getAllPlugins(): PluginRegistryEntry[] {
    return Array.from(this.plugins.values());
  }

  searchPlugins(query: string): PluginRegistryEntry[] {
    const queryLower = query.toLowerCase();
    const results: PluginRegistryEntry[] = [];

    // Search by ID
    for (const [id, entry] of this.plugins) {
      if (id.toLowerCase().includes(queryLower)) {
        results.push(entry);
      }
    }

    // Search by name
    for (const entry of this.plugins.values()) {
      if (entry.metadata.name.toLowerCase().includes(queryLower)) {
        if (!results.some(r => r.id === entry.id)) {
          results.push(entry);
        }
      }
    }

    // Search by description
    for (const entry of this.plugins.values()) {
      if (entry.metadata.description.toLowerCase().includes(queryLower)) {
        if (!results.some(r => r.id === entry.id)) {
          results.push(entry);
        }
      }
    }

    // Search by tags
    for (const entry of this.plugins.values()) {
      if (entry.metadata.tags?.some(tag => tag.toLowerCase().includes(queryLower))) {
        if (!results.some(r => r.id === entry.id)) {
          results.push(entry);
        }
      }
    }

    return results;
  }

  getPluginsByType(type: PluginType): PluginRegistryEntry[] {
    return Array.from(this.plugins.values()).filter(plugin => plugin.type === type);
  }

  getPluginsByStatus(status: PluginStatus): PluginRegistryEntry[] {
    return Array.from(this.plugins.values()).filter(plugin => plugin.status === status);
  }

  getPluginsByCategory(category: string): PluginRegistryEntry[] {
    return Array.from(this.plugins.values()).filter(plugin => 
      plugin.metadata.category.toLowerCase() === category.toLowerCase()
    );
  }

  updatePluginStatus(pluginId: string, status: PluginStatus): boolean {
    const entry = this.plugins.get(pluginId);
    if (!entry) {
      return false;
    }

    entry.status = status;
    entry.lastUpdated = new Date();
    return true;
  }

  private updateSearchIndex(entry: PluginRegistryEntry): void {
    // Index by name
    const nameTokens = this.tokenize(entry.metadata.name);
    for (const token of nameTokens) {
      if (!this.pluginSearchIndex.has(token)) {
        this.pluginSearchIndex.set(token, []);
      }
      const pluginList = this.pluginSearchIndex.get(token)!;
      if (!pluginList.includes(entry.id)) {
        pluginList.push(entry.id);
      }
    }

    // Index by description
    const descTokens = this.tokenize(entry.metadata.description);
    for (const token of descTokens) {
      if (!this.pluginSearchIndex.has(token)) {
        this.pluginSearchIndex.set(token, []);
      }
      const pluginList = this.pluginSearchIndex.get(token)!;
      if (!pluginList.includes(entry.id)) {
        pluginList.push(entry.id);
      }
    }

    // Index by tags
    if (entry.metadata.tags) {
      for (const tag of entry.metadata.tags) {
        const tagTokens = this.tokenize(tag);
        for (const token of tagTokens) {
          if (!this.pluginSearchIndex.has(token)) {
            this.pluginSearchIndex.set(token, []);
          }
          const pluginList = this.pluginSearchIndex.get(token)!;
          if (!pluginList.includes(entry.id)) {
            pluginList.push(entry.id);
          }
        }
      }
    }
  }

  private removeFromSearchIndex(entry: PluginRegistryEntry): void {
    // Remove all tokens related to this entry
    const allTokens = new Set<string>();
    
    // Collect all tokens from this entry
    const nameTokens = this.tokenize(entry.metadata.name);
    const descTokens = this.tokenize(entry.metadata.description);
    const tagTokens = entry.metadata.tags?.flatMap(tag => this.tokenize(tag)) || [];
    
    [...nameTokens, ...descTokens, ...tagTokens].forEach(token => allTokens.add(token));

    // Remove entry from all token indexes
    for (const token of allTokens) {
      const pluginList = this.pluginSearchIndex.get(token);
      if (pluginList) {
        const index = pluginList.indexOf(entry.id);
        if (index >= 0) {
          pluginList.splice(index, 1);
        }
        if (pluginList.length === 0) {
          this.pluginSearchIndex.delete(token);
        }
      }
    }
  }

  private tokenize(text: string): string[] {
    return text.toLowerCase()
      .split(/\W+/)
      .filter(token => token.length > 0);
  }
}