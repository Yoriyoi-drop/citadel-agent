// packages/plugin-system/src/index.ts
import { PluginRegistry } from './registry/plugin-registry';
import { PluginLoader } from './loader/plugin-loader';
import { PluginSandbox } from './sandbox/plugin-sandbox';
import { MarketplaceAPI } from './marketplace/marketplace-api';
import { PluginManifest, PluginMetadata, PluginType, PluginStatus, PluginGrade } from './types/plugin';

export {
  // Core classes
  PluginRegistry,
  PluginLoader,
  PluginSandbox,
  MarketplaceAPI,
  
  // Types
  PluginManifest,
  PluginMetadata,
  PluginType,
  PluginStatus,
  PluginGrade,
};

export default {
  PluginRegistry,
  PluginLoader,
  PluginSandbox,
  MarketplaceAPI,
};

// Convenience function to initialize the plugin system
export function initializePluginSystem(pluginsDir: string = 'plugins', apiPort?: number) {
  const loader = new PluginLoader(pluginsDir);
  
  if (apiPort) {
    const marketplace = new MarketplaceAPI(pluginsDir);
    marketplace.start(apiPort);
    return { loader, marketplace };
  }
  
  return { loader };
}