// frontend/src/utils/api.js
import axios from 'axios';

// Create axios instance with defaults
const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:5001/api/v1',
  timeout: 30000, // 30 seconds
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle common errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Clear auth data and redirect to login
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// API functions for workflows
export const workflowAPI = {
  // Get all workflows
  getWorkflows: (page = 0, limit = 20) => {
    return api.get(`/workflows?page=${page}&limit=${limit}`);
  },

  // Get workflow by ID
  getWorkflow: (id) => {
    return api.get(`/workflows/${id}`);
  },

  // Create workflow
  createWorkflow: (data) => {
    return api.post('/workflows', data);
  },

  // Update workflow
  updateWorkflow: (id, data) => {
    return api.put(`/workflows/${id}`, data);
  },

  // Delete workflow
  deleteWorkflow: (id) => {
    return api.delete(`/workflows/${id}`);
  },

  // Execute workflow
  executeWorkflow: (id, params = {}) => {
    return api.post(`/workflows/${id}/run`, params);
  },

  // Update workflow status
  updateWorkflowStatus: (id, status) => {
    return api.put(`/workflows/${id}/status`, { status });
  }
};

// API functions for executions
export const executionAPI = {
  // Get execution by ID
  getExecution: (id) => {
    return api.get(`/executions/${id}`);
  },

  // Get execution logs
  getExecutionLogs: (id, page = 0, limit = 20) => {
    return api.get(`/executions/${id}/logs?page=${page}&limit=${limit}`);
  },

  // Retry execution
  retryExecution: (id) => {
    return api.post(`/executions/${id}/retry`);
  },

  // Cancel execution
  cancelExecution: (id) => {
    return api.post(`/executions/${id}/cancel`);
  }
};

// API functions for auth
export const authAPI = {
  // Login
  login: (email, password) => {
    return api.post('/auth/login', { email, password });
  },

  // Register
  register: (name, email, password) => {
    return api.post('/auth/register', { name, email, password });
  },

  // Get current user
  getMe: () => {
    return api.get('/auth/me');
  },

  // Update profile
  updateProfile: (profile) => {
    return api.put('/auth/profile', profile);
  },

  // Update preferences
  updatePreferences: (preferences) => {
    return api.put('/auth/preferences', preferences);
  },

  // Change password
  changePassword: (oldPassword, newPassword) => {
    return api.put('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword
    });
  },

  // Create API key
  createAPIKey: (name, permissions, expiresIn, teamId) => {
    return api.post('/auth/api-keys', {
      name,
      permissions,
      expires_in_days: expiresIn,
      team_id: teamId
    });
  },

  // Get API keys
  getAPIKeys: () => {
    return api.get('/auth/api-keys');
  },

  // Revoke API key
  revokeAPIKey: (id) => {
    return api.delete(`/auth/api-keys/${id}`);
  }
};

// API functions for users
export const userAPI = {
  // Search users
  searchUsers: (query, page = 0, limit = 20) => {
    return api.get(`/users/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`);
  }
};

// API functions for teams
export const teamAPI = {
  // Get current user's teams
  getMyTeams: () => {
    return api.get('/teams/my');
  },

  // Get team by ID
  getTeam: (id) => {
    return api.get(`/teams/${id}`);
  },

  // Create team
  createTeam: (data) => {
    return api.post('/teams', data);
  },

  // Update team
  updateTeam: (id, data) => {
    return api.put(`/teams/${id}`, data);
  },

  // Delete team
  deleteTeam: (id) => {
    return api.delete(`/teams/${id}`);
  },

  // Add member to team
  addTeamMember: (teamId, userId, role) => {
    return api.post(`/teams/${teamId}/members`, {
      user_id: userId,
      role
    });
  },

  // Remove member from team
  removeTeamMember: (teamId, userId) => {
    return api.delete(`/teams/${teamId}/members/${userId}`);
  }
};

// General API utility functions
export const apiUtils = {
  // Handle API errors
  handleAPIError: (error) => {
    if (error.response) {
      // Server responded with error status
      return error.response.data?.error || error.response.statusText;
    } else if (error.request) {
      // Request was made but no response received
      return 'Network error: Please check your connection';
    } else {
      // Something else happened
      return error.message;
    }
  },

  // Format API response for UI
  formatResponse: (response) => {
    return {
      data: response.data,
      pagination: response.data.pagination || null,
      success: true
    };
  }
};

export default api;