// packages/plugin-system/src/loader/plugin-loader.ts
import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';
import { PluginManifest, PluginMetadata, PluginStatus, SecurityPolicy } from '../types/plugin';
import { PluginRegistry } from '../registry/plugin-registry';

export class PluginLoader {
  private pluginRegistry = PluginRegistry.getInstance();
  private pluginsDir: string;

  constructor(pluginsDir: string = 'plugins') {
    this.pluginsDir = pluginsDir;
    if (!fs.existsSync(pluginsDir)) {
      fs.mkdirSync(pluginsDir, { recursive: true });
    }
  }

  async loadPlugin(pluginPath: string): Promise<boolean> {
    try {
      const manifestPath = path.join(pluginPath, 'plugin.json');
      if (!fs.existsSync(manifestPath)) {
        console.error(`Plugin manifest not found at ${manifestPath}`);
        return false;
      }

      const manifestContent = fs.readFileSync(manifestPath, 'utf8');
      const manifest: PluginManifest = JSON.parse(manifestContent);

      // Validate manifest
      if (!this.validateManifest(manifest)) {
        console.error(`Invalid manifest for plugin: ${manifest.metadata.id}`);
        return false;
      }

      // Verify plugin files
      if (!this.verifyPluginFiles(pluginPath, manifest)) {
        console.error(`Plugin file verification failed for: ${manifest.metadata.id}`);
        return false;
      }

      // Check security policy
      if (!this.checkSecurityPolicy(manifest)) {
        console.error(`Security policy violation for plugin: ${manifest.metadata.id}`);
        return false;
      }

      // Register the plugin
      const success = this.pluginRegistry.registerPlugin({
        id: manifest.metadata.id,
        metadata: manifest.metadata,
        version: manifest.metadata.version,
        type: manifest.metadata.category as any, // Simplified mapping
        status: PluginStatus.INSTALLED,
        installPath: pluginPath,
        dependencies: manifest.metadata.dependencies?.map(dep => dep.id) || [],
        installedAt: new Date(),
        lastUpdated: new Date()
      });

      if (success) {
        console.log(`Successfully loaded plugin: ${manifest.metadata.name}`);
        return true;
      } else {
        console.error(`Failed to register plugin: ${manifest.metadata.id}`);
        return false;
      }
    } catch (error) {
      console.error(`Error loading plugin from ${pluginPath}:`, error);
      return false;
    }
  }

  async installPluginFromUrl(url: string): Promise<boolean> {
    // In a real implementation, this would download and install a plugin from a URL
    // For now, we'll simulate this functionality
    console.log(`Installing plugin from URL: ${url}`);
    // This would download, verify, and install the plugin
    return true;
  }

  async installPluginFromFile(filePath: string): Promise<boolean> {
    // In a real implementation, this would install a plugin from a local file
    console.log(`Installing plugin from file: ${filePath}`);
    // This would extract, verify, and install the plugin
    return true;
  }

  async uninstallPlugin(pluginId: string): Promise<boolean> {
    const entry = this.pluginRegistry.getPlugin(pluginId);
    if (!entry) {
      console.error(`Plugin not found: ${pluginId}`);
      return false;
    }

    try {
      // Execute uninstall hook if defined
      // const manifest = this.loadManifest(entry.installPath);
      // if (manifest.hooks?.uninstall) {
      //   await this.executeHook(manifest.hooks.uninstall, entry.installPath);
      // }

      // Remove from registry
      const success = this.pluginRegistry.unregisterPlugin(pluginId);
      
      if (success) {
        console.log(`Successfully uninstalled plugin: ${pluginId}`);
        return true;
      } else {
        return false;
      }
    } catch (error) {
      console.error(`Error uninstalling plugin ${pluginId}:`, error);
      return false;
    }
  }

  async updatePlugin(pluginId: string, newVersionPath: string): Promise<boolean> {
    const currentEntry = this.pluginRegistry.getPlugin(pluginId);
    if (!currentEntry) {
      console.error(`Plugin not found: ${pluginId}`);
      return false;
    }

    try {
      this.pluginRegistry.updatePluginStatus(pluginId, PluginStatus.UPDATING);
      
      // Load and validate new version
      const manifestPath = path.join(newVersionPath, 'plugin.json');
      const manifestContent = fs.readFileSync(manifestPath, 'utf8');
      const newManifest: PluginManifest = JSON.parse(manifestContent);

      if (!this.validateManifest(newManifest)) {
        console.error(`Invalid manifest for new version of plugin: ${pluginId}`);
        this.pluginRegistry.updatePluginStatus(pluginId, currentEntry.status);
        return false;
      }

      // Verify new version files
      if (!this.verifyPluginFiles(newVersionPath, newManifest)) {
        console.error(`New version file verification failed: ${pluginId}`);
        this.pluginRegistry.updatePluginStatus(pluginId, currentEntry.status);
        return false;
      }

      // Update the registry entry
      const success = this.pluginRegistry.registerPlugin({
        id: newManifest.metadata.id,
        metadata: newManifest.metadata,
        version: newManifest.metadata.version,
        type: newManifest.metadata.category as any,
        status: PluginStatus.INSTALLED,
        installPath: newVersionPath,
        dependencies: newManifest.metadata.dependencies?.map(dep => dep.id) || [],
        installedAt: currentEntry.installedAt,
        lastUpdated: new Date()
      });

      if (success) {
        console.log(`Successfully updated plugin: ${pluginId}`);
        return true;
      } else {
        this.pluginRegistry.updatePluginStatus(pluginId, currentEntry.status);
        return false;
      }
    } catch (error) {
      console.error(`Error updating plugin ${pluginId}:`, error);
      this.pluginRegistry.updatePluginStatus(pluginId, currentEntry.status);
      return false;
    }
  }

  private validateManifest(manifest: PluginManifest): boolean {
    // Check required fields
    if (!manifest.metadata.id || !manifest.metadata.name || !manifest.metadata.version) {
      console.error('Missing required manifest fields');
      return false;
    }

    // Check semantic versioning
    const semVerRegex = /^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/;
    if (!semVerRegex.test(manifest.metadata.version)) {
      console.error('Invalid semantic version in manifest');
      return false;
    }

    return true;
  }

  private verifyPluginFiles(pluginPath: string, manifest: PluginManifest): boolean {
    // Verify each file in the manifest exists and matches hash
    for (const file of manifest.files) {
      const fullPath = path.join(pluginPath, file.path);
      if (!fs.existsSync(fullPath)) {
        console.error(`Plugin file does not exist: ${fullPath}`);
        return false;
      }

      // Calculate file hash and compare with manifest
      const fileContent = fs.readFileSync(fullPath);
      const fileHash = crypto.createHash('sha256').update(fileContent).digest('hex');
      
      if (fileHash !== file.hash) {
        console.error(`File hash mismatch for: ${fullPath}`);
        return false;
      }
    }

    return true;
  }

  private checkSecurityPolicy(manifest: PluginManifest): boolean {
    // Check if the plugin is requesting too many permissions
    const criticalPermissions = manifest.permissions.filter(
      p => p.type === 'system' || (p.type === 'file_system' && p.access === 'write')
    );

    // For now, just log critical permissions - in a real system you'd have more complex checks
    if (criticalPermissions.length > 5) {
      console.warn(`Plugin ${manifest.metadata.id} requests many critical permissions`);
    }

    return true;
  }

  getInstalledPlugins(): PluginMetadata[] {
    return this.pluginRegistry.getAllPlugins()
      .filter(plugin => plugin.status === PluginStatus.INSTALLED)
      .map(plugin => plugin.metadata);
  }
}