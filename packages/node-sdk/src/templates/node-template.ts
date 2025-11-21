// templates/node-template.ts
import { NodeDefinition, NodeInterface } from '../interfaces/node';

export class NodeTemplate {
  static createNodeTemplate(nodeType: string, name: string): string {
    return `import { NodeExecutionContext, NodeExecutionResult, NodeInterface } from '@citadel/node-sdk';

export class ${this.pascalCase(name)}Node implements NodeInterface {
  async execute(context: NodeExecutionContext): Promise<NodeExecutionResult> {
    try {
      // Extract input data
      const { input, credentials, settings } = context;
      
      // Perform node-specific operations here
      console.log('${name} node executing with input:', input);
      
      // Example operation - replace with actual logic
      const result = await this.process(input, credentials, settings);
      
      return {
        success: true,
        data: result,
        metadata: {
          executionTime: Date.now(),
          timestamp: new Date(),
        },
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error occurred',
      };
    }
  }

  private async process(input: any, credentials?: any, settings?: any): Promise<any> {
    // Implement the actual logic here
    return {
      processed: true,
      input,
      timestamp: new Date().toISOString(),
    };
  }
}

// Export the node definition
export const ${this.camelCase(name)}Definition: NodeDefinition = {
  id: '${this.kebabCase(name)}',
  type: '${nodeType}',
  metadata: {
    name: '${name}',
    description: 'A description for this ${name} node',
    version: '1.0.0',
    category: 'Utility',
    tags: ['${name}', 'utility'],
  },
  inputs: [
    {
      id: 'data',
      name: 'Input Data',
      type: 'json',
      required: true,
      description: 'The data to process',
    },
  ],
  outputs: [
    {
      id: 'result',
      name: 'Processed Result',
      type: 'json',
      description: 'The processed result',
    },
  ],
  settings: [
    {
      id: 'timeout',
      name: 'Timeout (ms)',
      type: 'number',
      required: false,
      default: 30000,
      description: 'Maximum execution time in milliseconds',
    },
  ],
};
`;
  }

  private static pascalCase(str: string): string {
    return str
      .replace(/(?:^\w|[A-Z]|\b\w)/g, (word, index) => {
        return index === 0 ? word.toUpperCase() : word.toUpperCase();
      })
      .replace(/\s+/g, '');
  }

  private static camelCase(str: string): string {
    return str
      .replace(/(?:^\w|[A-Z]|\b\w)/g, (word, index) => {
        return index === 0 ? word.toLowerCase() : word.toUpperCase();
      })
      .replace(/\s+/g, '');
  }

  private static kebabCase(str: string): string {
    return str.toLowerCase().replace(/\s+/g, '-');
  }
}