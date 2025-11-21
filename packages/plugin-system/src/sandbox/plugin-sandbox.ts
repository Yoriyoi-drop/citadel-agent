// packages/plugin-system/src/sandbox/plugin-sandbox.ts
import * as vm from 'vm';
import * as fs from 'fs';
import * as path from 'path';
import { PluginMetadata, SecurityPolicy } from '../types/plugin';

export interface SandboxOptions {
  timeout: number;
  memoryLimit: number;
  allowedModules: string[];
  blockedModules: string[];
  allowedDomains: string[];
  securityPolicy?: SecurityPolicy;
}

export class PluginSandbox {
  private defaultOptions: SandboxOptions = {
    timeout: 5000, // 5 seconds
    memoryLimit: 128 * 1024 * 1024, // 128 MB
    allowedModules: ['querystring', 'url', 'path', 'crypto', 'buffer', 'stream', 'util'],
    blockedModules: ['fs', 'child_process', 'cluster', 'dgram', 'dns', 'net', 'tls', 'repl', 'worker_threads'],
    allowedDomains: [],
  };

  async executePluginCode(
    code: string, 
    context: any, 
    options?: Partial<SandboxOptions>
  ): Promise<any> {
    const opts = { ...this.defaultOptions, ...options };

    // Create a context with limited access
    const sandboxContext = {
      // Safe global objects
      console: {
        log: (...args: any[]) => console.log('[Plugin Log]', ...args),
        error: (...args: any[]) => console.error('[Plugin Error]', ...args),
        warn: (...args: any[]) => console.warn('[Plugin Warning]', ...args),
      },
      setTimeout,
      clearTimeout,
      setInterval,
      clearInterval,
      Buffer,
      URL,
      URLSearchParams,
      TextEncoder,
      TextDecoder,
      
      // Limited global objects
      Date,
      Math,
      
      // Plugin's input context
      ...context,
      
      // Allow limited access to JSON for data manipulation
      JSON: {
        parse: JSON.parse,
        stringify: JSON.stringify,
      },
      
      // Custom plugin APIs (safely implemented)
      pluginAPI: this.createSafePluginAPI(),
    };

    // Create the VM context
    const vmContext = vm.createContext(sandboxContext);

    try {
      // Execute the code in the sandbox
      const result = vm.runInContext(code, vmContext, {
        timeout: opts.timeout,
        displayErrors: true,
      });

      return result;
    } catch (error) {
      throw new Error(`Plugin execution error: ${(error as Error).message}`);
    }
  }

  async executeNodePlugin(
    nodeCode: string,
    inputs: Record<string, any>,
    pluginMetadata: PluginMetadata,
    securityPolicy?: SecurityPolicy
  ): Promise<any> {
    // Prepare execution context with inputs
    const context = {
      inputs,
      config: pluginMetadata,
      utils: this.createNodeUtils(),
    };

    // Apply security policy
    const sandboxOptions = this.applySecurityPolicy(securityPolicy);

    return await this.executePluginCode(nodeCode, context, sandboxOptions);
  }

  private createSafePluginAPI(): any {
    return {
      // Safe data transformation utilities
      transform: {
        json: (data: any) => {
          try {
            return JSON.parse(JSON.stringify(data));
          } catch {
            return data;
          }
        },
        array: {
          map: (arr: any[], fn: Function) => arr.map(fn),
          filter: (arr: any[], fn: Function) => arr.filter(fn),
          reduce: (arr: any[], fn: Function, initial?: any) => arr.reduce(fn, initial),
        },
        string: {
          capitalize: (str: string) => str.charAt(0).toUpperCase() + str.slice(1),
          truncate: (str: string, length: number) => str.length > length ? str.substring(0, length) + '...' : str,
        },
      },
      
      // Safe utility functions
      utils: {
        sleep: (ms: number) => new Promise(resolve => setTimeout(resolve, ms)),
        uuid: () => {
          return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
          });
        },
        hash: (data: string) => {
          const crypto = require('crypto');
          return crypto.createHash('sha256').update(data).digest('hex');
        },
      },
    };
  }

  private createNodeUtils(): any {
    return {
      validate: (data: any, schema: any) => this.validateData(data, schema),
      format: (data: any, format: string) => this.formatData(data, format),
      transform: (data: any, operations: any[]) => this.transformData(data, operations),
    };
  }

  private validateData(data: any, schema: any): { valid: boolean; errors?: string[] } {
    // Basic validation implementation
    // In a real system, you'd use a proper validation library
    try {
      // Simple type checking
      if (schema.type && typeof data !== schema.type) {
        return { valid: false, errors: [`Expected type ${schema.type}, got ${typeof data}`] };
      }

      return { valid: true };
    } catch (error) {
      return { valid: false, errors: [(error as Error).message] };
    }
  }

  private formatData(data: any, format: string): any {
    switch (format) {
      case 'json':
        return JSON.stringify(data, null, 2);
      case 'csv':
        // Simplified CSV formatting
        if (Array.isArray(data) && data.length > 0) {
          const headers = Object.keys(data[0]).join(',');
          const rows = data.map(obj => Object.values(obj).join(','));
          return [headers, ...rows].join('\n');
        }
        return data;
      default:
        return data;
    }
  }

  private transformData(data: any, operations: any[]): any {
    // Apply a series of transformations to the data
    let result = data;
    for (const op of operations) {
      switch (op.type) {
        case 'rename':
          if (typeof result === 'object' && result[op.from] !== undefined) {
            result[op.to] = result[op.from];
            delete result[op.from];
          }
          break;
        case 'filter':
          if (Array.isArray(result)) {
            result = result.filter(op.fn);
          }
          break;
        case 'map':
          if (Array.isArray(result)) {
            result = result.map(op.fn);
          }
          break;
        default:
          console.warn(`Unknown transformation type: ${op.type}`);
      }
    }
    return result;
  }

  private applySecurityPolicy(policy?: SecurityPolicy): Partial<SandboxOptions> {
    if (!policy) return {};

    return {
      timeout: policy.maxExecutionTime,
      memoryLimit: policy.memoryLimit,
      allowedDomains: policy.allowedDomains,
    };
  }
}