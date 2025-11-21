// packages/plugin-system/src/marketplace/marketplace-api.ts
import express, { Request, Response } from 'express';
import { PluginRegistry } from '../registry/plugin-registry';
import { PluginLoader } from '../loader/plugin-loader';
import { PluginMetadata, PluginType, PluginGrade, PluginStatus } from '../types/plugin';

export class MarketplaceAPI {
  private app = express();
  private pluginRegistry = PluginRegistry.getInstance();
  private pluginLoader: PluginLoader;

  constructor(pluginsDir: string = 'plugins') {
    this.pluginLoader = new PluginLoader(pluginsDir);
    this.setupRoutes();
  }

  private setupRoutes(): void {
    // Middleware
    this.app.use(express.json());

    // Get all plugins
    this.app.get('/plugins', (req: Request, res: Response) => {
      try {
        const { type, category, grade, status, search } = req.query;
        
        let plugins = this.pluginRegistry.getAllPlugins().map(entry => entry.metadata);
        
        // Apply filters
        if (type) {
          plugins = plugins.filter(p => p.category === type);
        }
        
        if (category) {
          plugins = plugins.filter(p => p.category.toLowerCase().includes((category as string).toLowerCase()));
        }
        
        if (grade) {
          plugins = plugins.filter(p => p.grade === grade);
        }
        
        if (status) {
          // This would require checking installed status rather than registry status
          const installedPlugins = this.pluginLoader.getInstalledPlugins().map(p => p.id);
          if (status === 'installed') {
            plugins = plugins.filter(p => installedPlugins.includes(p.id));
          } else if (status === 'available') {
            plugins = plugins.filter(p => !installedPlugins.includes(p.id));
          }
        }
        
        if (search) {
          const query = search as string;
          plugins = this.pluginRegistry.searchPlugins(query).map(entry => entry.metadata);
        }

        res.json({
          success: true,
          data: plugins,
          count: plugins.length
        });
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });

    // Get plugin by ID
    this.app.get('/plugins/:id', (req: Request, res: Response) => {
      try {
        const plugin = this.pluginRegistry.getPlugin(req.params.id);
        
        if (!plugin) {
          return res.status(404).json({
            success: false,
            error: 'Plugin not found'
          });
        }

        res.json({
          success: true,
          data: plugin.metadata
        });
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });

    // Install plugin
    this.app.post('/plugins/:id/install', async (req: Request, res: Response) => {
      try {
        const { id } = req.params;
        const { source } = req.body; // URL, file path, or marketplace ID
        
        let success = false;
        
        if (source?.startsWith('http')) {
          // Install from URL
          success = await this.pluginLoader.installPluginFromUrl(source);
        } else if (source) {
          // Install from local file
          success = await this.pluginLoader.installPluginFromFile(source);
        } else {
          // Install from marketplace (default)
          // This would fetch from a marketplace service
          success = await this.installFromMarketplace(id);
        }

        if (success) {
          res.json({
            success: true,
            message: `Plugin ${id} installed successfully`
          });
        } else {
          res.status(400).json({
            success: false,
            error: `Failed to install plugin ${id}`
          });
        }
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });

    // Uninstall plugin
    this.app.delete('/plugins/:id', async (req: Request, res: Response) => {
      try {
        const { id } = req.params;
        
        const success = await this.pluginLoader.uninstallPlugin(id);
        
        if (success) {
          res.json({
            success: true,
            message: `Plugin ${id} uninstalled successfully`
          });
        } else {
          res.status(400).json({
            success: false,
            error: `Failed to uninstall plugin ${id}`
          });
        }
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });

    // Update plugin
    this.app.put('/plugins/:id/update', async (req: Request, res: Response) => {
      try {
        const { id } = req.params;
        const { source } = req.body;
        
        if (!source) {
          return res.status(400).json({
            success: false,
            error: 'Update source is required'
          });
        }
        
        const success = await this.pluginLoader.updatePlugin(id, source);
        
        if (success) {
          res.json({
            success: true,
            message: `Plugin ${id} updated successfully`
          });
        } else {
          res.status(400).json({
            success: false,
            error: `Failed to update plugin ${id}`
          });
        }
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });

    // Get installed plugins
    this.app.get('/plugins/installed', (req: Request, res: Response) => {
      try {
        const installedPlugins = this.pluginLoader.getInstalledPlugins();
        
        res.json({
          success: true,
          data: installedPlugins,
          count: installedPlugins.length
        });
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });

    // Get plugin statistics
    this.app.get('/plugins/stats', (req: Request, res: Response) => {
      try {
        const allPlugins = this.pluginRegistry.getAllPlugins();
        const installed = this.pluginLoader.getInstalledPlugins();
        
        const stats = {
          total: allPlugins.length,
          installed: installed.length,
          byType: allPlugins.reduce((acc, entry) => {
            acc[entry.metadata.category] = (acc[entry.metadata.category] || 0) + 1;
            return acc;
          }, {} as Record<string, number>),
          byGrade: allPlugins.reduce((acc, entry) => {
            acc[entry.metadata.grade] = (acc[entry.metadata.grade] || 0) + 1;
            return acc;
          }, {} as Record<string, number>)
        };
        
        res.json({
          success: true,
          data: stats
        });
      } catch (error) {
        res.status(500).json({
          success: false,
          error: (error as Error).message
        });
      }
    });
  }

  private async installFromMarketplace(pluginId: string): Promise<boolean> {
    // In a real implementation, this would fetch from a remote marketplace
    // For now, we'll simulate the installation
    console.log(`Installing plugin ${pluginId} from marketplace`);
    return true;
  }

  public getApp(): express.Application {
    return this.app;
  }

  public start(port: number = 8080): void {
    this.app.listen(port, () => {
      console.log(`Marketplace API running on port ${port}`);
    });
  }
}