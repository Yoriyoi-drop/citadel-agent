// frontend/src/App.jsx
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { WorkflowProvider } from './context/WorkflowContext';

// Layouts
import MainLayout from './layouts/MainLayout';

// Pages
import LoginPage from './pages/auth/LoginPage';
import RegisterPage from './pages/auth/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import WorkflowBuilder from './pages/workflow/WorkflowBuilder';
import WorkflowList from './pages/workflow/WorkflowList';
import WorkflowDetail from './pages/workflow/WorkflowDetail';
import SettingsPage from './pages/SettingsPage';
import NotFoundPage from './pages/NotFoundPage';

function App() {
  return (
    <AuthProvider>
      <WorkflowProvider>
        <Router>
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            
            {/* Protected Routes */}
            <Route path="/" element={<MainLayout />}>
              <Route index element={<DashboardPage />} />
              <Route path="workflows" element={<WorkflowList />} />
              <Route path="workflows/new" element={<WorkflowBuilder />} />
              <Route path="workflows/:id" element={<WorkflowDetail />} />
              <Route path="workflows/:id/edit" element={<WorkflowBuilder />} />
              <Route path="settings" element={<SettingsPage />} />
              <Route path="*" element={<NotFoundPage />} />
            </Route>
          </Routes>
        </Router>
      </WorkflowProvider>
    </AuthProvider>
  );
}

export default App;