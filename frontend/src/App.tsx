// frontend/src/App.tsx
import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';

import Sidebar from './components/Sidebar/Sidebar';
import WorkflowCanvas from './components/WorkflowCanvas/Canvas';
import Inspector from './components/Inspector/Inspector';
import Dashboard from './components/Dashboard/Dashboard';
import { NodeData } from './types';

// Auth context to manage user authentication
interface AuthContextType {
  user: any;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
  isAuthenticated: boolean;
}

export const AuthContext = React.createContext<AuthContextType | null>(null);

const App: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // Check if user is already authenticated on app load
  useEffect(() => {
    // In a real implementation, check for token in localStorage or cookie
    const token = localStorage.getItem('citadel_token');
    if (token) {
      // Validate token with backend (simplified)
      setIsAuthenticated(true);
    }
    setLoading(false);
  }, []);

  const login = async (email: string, password: string): Promise<boolean> => {
    try {
      // In a real implementation, make API call to backend
      // const response = await fetch('/api/auth/login', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ email, password })
      // });
      // const data = await response.json();
      
      // For demo purposes
      const token = 'demo-token-' + Date.now();
      localStorage.setItem('citadel_token', token);
      setIsAuthenticated(true);
      setUser({ id: '1', email, name: email.split('@')[0] });
      return true;
    } catch (error) {
      console.error('Login error:', error);
      return false;
    }
  };

  const logout = () => {
    localStorage.removeItem('citadel_token');
    setIsAuthenticated(false);
    setUser(null);
  };

  // Protected route component
  const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    if (loading) {
      return <div className="flex justify-center items-center h-screen">Loading...</div>;
    }
    
    if (!isAuthenticated) {
      return <Navigate to="/login" replace />;
    }
    
    return <>{children}</>;
  };

  // Login page component
  const LoginPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      const success = await login(email, password);
      if (!success) {
        setError('Invalid email or password');
      }
    };

    if (isAuthenticated) {
      return <Navigate to="/dashboard" replace />;
    }

    return (
      <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to Citadel Agent
          </h2>
        </div>

        <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
          <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
            <form className="space-y-6" onSubmit={handleSubmit}>
              {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
                  {error}
                </div>
              )}
              
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                  Email address
                </label>
                <div className="mt-1">
                  <input
                    id="email"
                    name="email"
                    type="email"
                    autoComplete="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  />
                </div>
              </div>

              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                  Password
                </label>
                <div className="mt-1">
                  <input
                    id="password"
                    name="password"
                    type="password"
                    autoComplete="current-password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  />
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div className="text-sm">
                  <a href="#" className="font-medium text-blue-600 hover:text-blue-500">
                    Forgot your password?
                  </a>
                </div>
              </div>

              <div>
                <button
                  type="submit"
                  className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Sign in
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    );
  };

  // Main layout component for dashboard
  const MainLayout: React.FC = () => {
    const [selectedWorkflow, setSelectedWorkflow] = useState<string | null>(null);
    const [selectedNode, setSelectedNode] = useState<NodeData | null>(null);

    return (
      <div className="flex h-screen bg-gray-100">
        {/* Sidebar */}
        <Sidebar 
          selectedWorkflow={selectedWorkflow} 
          onWorkflowSelect={setSelectedWorkflow}
        />

        {/* Main content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Top navigation */}
          <header className="bg-white shadow-sm z-10">
            <div className="px-4 sm:px-6 lg:px-8">
              <div className="flex justify-between h-16 items-center">
                <div className="flex items-center">
                  <h1 className="text-xl font-semibold text-gray-900">
                    {selectedWorkflow ? selectedWorkflow : 'Dashboard'}
                  </h1>
                </div>
                <div className="flex items-center space-x-4">
                  <button
                    onClick={logout}
                    className="text-sm text-gray-700 hover:text-gray-900"
                  >
                    Logout
                  </button>
                </div>
              </div>
            </div>
          </header>

          {/* Workflow canvas or dashboard */}
          <div className="flex-1 flex overflow-hidden">
            {selectedWorkflow ? (
              <DndProvider backend={HTML5Backend}>
                <div className="flex-1 flex flex-col overflow-hidden">
                  <WorkflowCanvas 
                    onSave={(nodes, edges) => {
                      console.log('Saving workflow:', nodes, edges);
                    }}
                    onRun={() => {
                      console.log('Running workflow');
                    }}
                  />
                </div>
              </DndProvider>
            ) : (
              <Dashboard />
            )}

            {/* Inspector panel (when node is selected) */}
            {selectedNode && <Inspector node={selectedNode} />}
          </div>
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <AuthContext.Provider value={{ user, login, logout, isAuthenticated }}>
      <Router>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route 
            path="/dashboard" 
            element={
              <ProtectedRoute>
                <MainLayout />
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/workflow/:id" 
            element={
              <ProtectedRoute>
                <MainLayout />
              </ProtectedRoute>
            } 
          />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </Router>
    </AuthContext.Provider>
  );
};

export default App;