// frontend/src/services/workflowService.js
import axios from 'axios';

class WorkflowService {
  constructor() {
    this.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:5001/api/v1';
    this.headers = {
      'Content-Type': 'application/json',
    };
  }

  // Get all workflows for user
  async getWorkflows() {
    try {
      const response = await axios.get(`${this.baseURL}/workflows`, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch workflows: ${error.message}`);
    }
  }

  // Get specific workflow by ID
  async getWorkflow(id) {
    try {
      const response = await axios.get(`${this.baseURL}/workflows/${id}`, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch workflow: ${error.message}`);
    }
  }

  // Create new workflow
  async createWorkflow(workflowData) {
    try {
      const response = await axios.post(`${this.baseURL}/workflows`, workflowData, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create workflow: ${error.message}`);
    }
  }

  // Update existing workflow
  async updateWorkflow(id, workflowData) {
    try {
      const response = await axios.put(`${this.baseURL}/workflows/${id}`, workflowData, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to update workflow: ${error.message}`);
    }
  }

  // Delete workflow
  async deleteWorkflow(id) {
    try {
      const response = await axios.delete(`${this.baseURL}/workflows/${id}`, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to delete workflow: ${error.message}`);
    }
  }

  // Execute workflow
  async executeWorkflow(id, executionData = {}) {
    try {
      const response = await axios.post(`${this.baseURL}/workflows/${id}/execute`, executionData, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to execute workflow: ${error.message}`);
    }
  }

  // Get workflow executions
  async getWorkflowExecutions(workflowId) {
    try {
      const response = await axios.get(`${this.baseURL}/workflows/${workflowId}/executions`, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch executions: ${error.message}`);
    }
  }

  // Get execution details
  async getExecutionDetails(executionId) {
    try {
      const response = await axios.get(`${this.baseURL}/executions/${executionId}`, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch execution details: ${error.message}`);
    }
  }

  // Get node types available
  async getNodeTypes() {
    try {
      const response = await axios.get(`${this.baseURL}/nodes/types`, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to fetch node types: ${error.message}`);
    }
  }

  // Test node configuration
  async testNode(nodeData) {
    try {
      const response = await axios.post(`${this.baseURL}/nodes/test`, nodeData, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to test node: ${error.message}`);
    }
  }

  // Validate workflow
  async validateWorkflow(workflowData) {
    try {
      const response = await axios.post(`${this.baseURL}/workflows/validate`, workflowData, {
        headers: this.headers,
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to validate workflow: ${error.message}`);
    }
  }
}

export default new WorkflowService();