// frontend/src/context/AuthContext.jsx
import React, { createContext, useContext, useState, useEffect } from 'react';
import axios from 'axios';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check if user is already logged in on initial load
  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    if (storedToken && storedUser) {
      setToken(storedToken);
      setUser(JSON.parse(storedUser));
      // Set default axios header
      axios.defaults.headers.common['Authorization'] = `Bearer ${storedToken}`;
    }
    
    setIsLoading(false);
  }, []);

  const login = async (email, password) => {
    try {
      const response = await axios.post('/api/v1/auth/login', {
        email,
        password
      });

      const { user: userData, token: tokenData } = response.data;

      setToken(tokenData.access_token);
      setUser(userData.user);
      
      // Store in localStorage
      localStorage.setItem('token', tokenData.access_token);
      localStorage.setItem('user', JSON.stringify(userData.user));
      
      // Set axios header
      axios.defaults.headers.common['Authorization'] = `Bearer ${tokenData.access_token}`;
      
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Login failed:', error);
      return { 
        success: false, 
        error: error.response?.data?.error || 'Login failed' 
      };
    }
  };

  const register = async (name, email, password) => {
    try {
      const response = await axios.post('/api/v1/auth/register', {
        name,
        email,
        password
      });

      const { user: userData, token: tokenData } = response.data;

      setToken(tokenData.access_token);
      setUser(userData.user);
      
      // Store in localStorage
      localStorage.setItem('token', tokenData.access_token);
      localStorage.setItem('user', JSON.stringify(userData.user));
      
      // Set axios header
      axios.defaults.headers.common['Authorization'] = `Bearer ${tokenData.access_token}`;
      
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Registration failed:', error);
      return { 
        success: false, 
        error: error.response?.data?.error || 'Registration failed' 
      };
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    
    // Clear localStorage
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    
    // Remove axios header
    delete axios.defaults.headers.common['Authorization'];
  };

  const changePassword = async (oldPassword, newPassword) => {
    try {
      const response = await axios.put('/api/v1/auth/password', {
        old_password: oldPassword,
        new_password: newPassword
      }, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      return { success: true, data: response.data };
    } catch (error) {
      console.error('Change password failed:', error);
      return { 
        success: false, 
        error: error.response?.data?.error || 'Change password failed' 
      };
    }
  };

  const value = {
    user,
    token,
    isLoading,
    login,
    register,
    logout,
    changePassword,
    isAuthenticated: !!token
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};