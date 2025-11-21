// frontend/src/services/api.ts
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:5001/api/v1';

interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

class ApiService {
  private baseUrl: string;
  private token: string | null;

  constructor() {
    this.baseUrl = API_BASE_URL;
    this.token = localStorage.getItem('citadel_token');
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as any)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'API request failed');
      }

      return { data };
    } catch (error) {
      return { 
        error: error instanceof Error ? error.message : 'Unknown error occurred' 
      };
    }
  }

  // Authentication
  async login(email: string, password: string): Promise<ApiResponse<{token: string; user: any}>> {
    return this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async register(userData: any): Promise<ApiResponse<any>> {
    return this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async logout(): Promise<void> {
    this.token = null;
    localStorage.removeItem('citadel_token');
  }

  // Workflows
  async getWorkflows(): Promise<ApiResponse<any[]>> {
    return this.request('/workflows');
  }

  async getWorkflow(id: string): Promise<ApiResponse<any>> {
    return this.request(`/workflows/${id}`);
  }

  async createWorkflow(workflow: any): Promise<ApiResponse<any>> {
    return this.request('/workflows', {
      method: 'POST',
      body: JSON.stringify(workflow),
    });
  }

  async updateWorkflow(id: string, workflow: any): Promise<ApiResponse<any>> {
    return this.request(`/workflows/${id}`, {
      method: 'PUT',
      body: JSON.stringify(workflow),
    });
  }

  async deleteWorkflow(id: string): Promise<ApiResponse<any>> {
    return this.request(`/workflows/${id}`, {
      method: 'DELETE',
    });
  }

  async runWorkflow(id: string): Promise<ApiResponse<any>> {
    return this.request(`/workflows/${id}/run`, {
      method: 'POST',
    });
  }

  // Executions
  async getExecutions(workflowId?: string): Promise<ApiResponse<any[]>> {
    const endpoint = workflowId ? `/executions?workflowId=${workflowId}` : '/executions';
    return this.request(endpoint);
  }

  async getExecution(id: string): Promise<ApiResponse<any>> {
    return this.request(`/executions/${id}`);
  }

  // Nodes
  async getNodes(): Promise<ApiResponse<any[]>> {
    return this.request('/nodes');
  }

  async getNodeTypes(): Promise<ApiResponse<any[]>> {
    return this.request('/nodes/types');
  }

  // Plugins
  async getPlugins(): Promise<ApiResponse<any[]>> {
    return this.request('/plugins');
  }

  async installPlugin(pluginId: string): Promise<ApiResponse<any>> {
    return this.request(`/plugins/${pluginId}/install`, {
      method: 'POST',
    });
  }

  async uninstallPlugin(pluginId: string): Promise<ApiResponse<any>> {
    return this.request(`/plugins/${pluginId}`, {
      method: 'DELETE',
    });
  }

  // Users
  async getCurrentUser(): Promise<ApiResponse<any>> {
    return this.request('/users/me');
  }

  async updateUser(userData: any): Promise<ApiResponse<any>> {
    return this.request('/users/me', {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
  }
}

export default new ApiService();