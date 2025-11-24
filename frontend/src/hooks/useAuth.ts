import { useCallback, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';

export function useAuth() {
  const {
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    login,
    logout,
    setUser,
    setToken,
    setLoading,
    setError: setAuthError,
    clearError
  } = useAuthStore();

  // Register new user
  const register = useCallback(async (email: string, password: string, name: string) => {
    setLoading(true);
    setAuthError(null);
    
    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password, name }),
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Registration failed');
      }
      
      const data = await response.json();
      setUser(data.user);
      setToken(data.token);
      
      return data;
    } catch (error) {
      setAuthError(error instanceof Error ? error.message : 'Registration failed');
      throw error;
    } finally {
      setLoading(false);
    }
  }, [setUser, setToken, setLoading, setAuthError]);

  // Logout user
  const logoutUser = useCallback(async () => {
    try {
      if (token) {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        });
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      logout();
    }
  }, [token, logout]);

  // Refresh token
  const refreshToken = useCallback(async () => {
    if (!token) return false;
    
    try {
      const response = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Token refresh failed');
      }
      
      const data = await response.json();
      setToken(data.token);
      
      return true;
    } catch (error) {
      console.error('Token refresh error:', error);
      logout();
      return false;
    }
  }, [token, setToken, logout]);

  // Check authentication status
  const checkAuth = useCallback(async () => {
    if (!token) return false;
    
    try {
      const response = await fetch('/api/auth/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Authentication check failed');
      }
      
      const data = await response.json();
      setUser(data.user);
      
      return true;
    } catch (error) {
      console.error('Auth check error:', error);
      logout();
      return false;
    }
  }, [token, setUser, logout]);

  // Update user profile
  const updateProfile = useCallback(async (updates: Partial<{ name: string; email: string }>) => {
    if (!token || !user) throw new Error('Not authenticated');
    
    try {
      const response = await fetch('/api/auth/profile', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(updates),
      });
      
      if (!response.ok) {
        throw new Error('Profile update failed');
      }
      
      const data = await response.json();
      setUser(data.user);
      
      return data.user;
    } catch (error) {
      setAuthError(error instanceof Error ? error.message : 'Profile update failed');
      throw error;
    }
  }, [token, user, setUser, setAuthError]);

  // Change password
  const changePassword = useCallback(async (currentPassword: string, newPassword: string) => {
    if (!token) throw new Error('Not authenticated');
    
    try {
      const response = await fetch('/api/auth/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ currentPassword, newPassword }),
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Password change failed');
      }
      
      return true;
    } catch (error) {
      setAuthError(error instanceof Error ? error.message : 'Password change failed');
      throw error;
    }
  }, [token, setAuthError]);

  // Auto-refresh token
  useEffect(() => {
    if (!token) return;
    
    const refreshInterval = setInterval(() => {
      refreshToken();
    }, 15 * 60 * 1000); // Refresh every 15 minutes
    
    return () => clearInterval(refreshInterval);
  }, [token, refreshToken]);

  // Check auth on mount
  useEffect(() => {
    if (token) {
      checkAuth();
    }
  }, [token, checkAuth]);

  return {
    // State
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    
    // Actions
    login,
    register,
    logout: logoutUser,
    updateProfile,
    changePassword,
    refreshToken,
    checkAuth,
    
    // Utilities
    clearError
  };
}