// Example HTTP Request Node
import { NodeExecutionContext, NodeExecutionResult, NodeInterface } from './interfaces/node';

export class HttpRequestNode implements NodeInterface {
  async execute(context: NodeExecutionContext): Promise<NodeExecutionResult> {
    try {
      const { input, credentials, settings } = context;
      
      // Extract input parameters
      const {
        url,
        method = 'GET',
        headers = {},
        body,
        timeout = 30000
      } = input;

      if (!url) {
        return {
          success: false,
          error: 'URL is required for HTTP request node',
        };
      }

      // Implement HTTP request logic here
      // This is a simplified example - in real implementation would use fetch/axios
      console.log(`Executing HTTP request: ${method} ${url}`);
      
      // Simulated response
      const response = {
        statusCode: 200,
        headers: { 'content-type': 'application/json' },
        data: { message: `Successfully processed ${method} request to ${url}` },
        url,
        method,
      };

      return {
        success: true,
        data: response,
        metadata: {
          executionTime: Date.now(),
          timestamp: new Date(),
        },
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error occurred during HTTP request',
      };
    }
  }
}

export const httpRequestDefinition = {
  id: 'http-request',
  type: 'action',
  metadata: {
    name: 'HTTP Request',
    description: 'Makes an HTTP request to a specified URL',
    version: '1.0.0',
    category: 'Integration',
    tags: ['http', 'api', 'request', 'integration'],
  },
  inputs: [
    {
      id: 'url',
      name: 'URL',
      type: 'string',
      required: true,
      description: 'The URL to make the request to',
    },
    {
      id: 'method',
      name: 'HTTP Method',
      type: 'string',
      required: false,
      default: 'GET',
      description: 'HTTP method (GET, POST, PUT, DELETE, etc.)',
    },
    {
      id: 'headers',
      name: 'Headers',
      type: 'json',
      required: false,
      description: 'Request headers as JSON object',
    },
    {
      id: 'body',
      name: 'Request Body',
      type: 'json',
      required: false,
      description: 'Request body for POST/PUT requests',
    },
    {
      id: 'timeout',
      name: 'Timeout (ms)',
      type: 'number',
      required: false,
      default: 30000,
      description: 'Request timeout in milliseconds',
    },
  ],
  outputs: [
    {
      id: 'response',
      name: 'Response',
      type: 'json',
      description: 'HTTP response object',
    },
  ],
  settings: [
    {
      id: 'retryCount',
      name: 'Retry Count',
      type: 'number',
      required: false,
      default: 0,
      description: 'Number of times to retry on failure',
    },
    {
      id: 'followRedirects',
      name: 'Follow Redirects',
      type: 'boolean',
      required: false,
      default: true,
      description: 'Whether to follow HTTP redirects',
    },
  ],
};