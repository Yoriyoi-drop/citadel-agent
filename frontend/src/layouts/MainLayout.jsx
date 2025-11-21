// frontend/src/layouts/MainLayout.jsx
import React, { useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { 
  HomeIcon, 
  QueueListIcon, 
  ChartBarIcon, 
  CogIcon, 
  UserIcon,
  ArrowLeftOnRectangleIcon,
  Bars3Icon,
  XMarkIcon
} from '@heroicons/react/24/outline';
import { useAuth } from '../context/AuthContext';

// Navigation items
const navigation = [
  { name: 'Dashboard', href: '/', icon: HomeIcon },
  { name: 'Workflows', href: '/workflows', icon: QueueListIcon },
  { name: 'Analytics', href: '/analytics', icon: ChartBarIcon },
  { name: 'Settings', href: '/settings', icon: CogIcon },
];

function MainLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Mobile sidebar */}
      <div className="md:hidden">
        <div className={`fixed inset-0 z-40 ${sidebarOpen ? 'block' : 'hidden'}`}>
          <div className="fixed inset-0 bg-gray-600 bg-opacity-75" onClick={() => setSidebarOpen(false)}></div>
          <div className="relative flex-1 flex flex-col max-w-xs w-full bg-white">
            <div className="absolute top-0 right-0 -mr-12 pt-2">
              <button
                type="button"
                className="ml-1 flex items-center justify-center h-10 w-10 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                onClick={() => setSidebarOpen(false)}
              >
                <span className="sr-only">Close sidebar</span>
                <XMarkIcon className="h-6 w-6 text-white" aria-hidden="true" />
              </button>
            </div>
            <div className="flex-1 h-0 pt-5 pb-4 overflow-y-auto">
              <div className="flex-shrink-0 flex items-center px-4">
                <div className="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center">
                  <span className="text-white font-bold text-sm">CA</span>
                </div>
                <span className="ml-2 text-xl font-bold text-gray-900">Citadel Agent</span>
              </div>
              <nav className="mt-5 px-2 space-y-1">
                {navigation.map((item) => {
                  const isActive = location.pathname === item.href;
                  return (
                    <Link
                      key={item.name}
                      to={item.href}
                      className={`${
                        isActive 
                          ? 'bg-blue-100 text-blue-600' 
                          : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                      } group flex items-center px-2 py-2 text-base font-medium rounded-md`}
                      onClick={() => setSidebarOpen(false)}
                    >
                      <item.icon
                        className={`${
                          isActive ? 'text-blue-600' : 'text-gray-400 group-hover:text-gray-500'
                        } mr-4 flex-shrink-0 h-6 w-6`}
                        aria-hidden="true"
                      />
                      {item.name}
                    </Link>
                  );
                })}
              </nav>
            </div>
            <div className="flex-shrink-0 flex border-t border-gray-200 p-4">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                    <UserIcon className="h-6 w-6 text-gray-600" />
                  </div>
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-gray-700">{user?.name || 'User'}</p>
                  <p className="text-xs font-medium text-gray-500">{user?.email || 'Account'}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Static sidebar for desktop */}
      <div className="hidden md:flex md:w-64 md:flex-col md:fixed md:inset-y-0">
        <div className="flex-1 flex flex-col min-h-0 border-r border-gray-200 bg-white">
          <div className="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
            <div className="flex items-center flex-shrink-0 px-4">
              <div className="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center">
                <span className="text-white font-bold text-sm">CA</span>
              </div>
              <span className="ml-2 text-xl font-bold text-gray-900">Citadel Agent</span>
            </div>
            <nav className="mt-5 flex-1 px-2 space-y-1">
              {navigation.map((item) => {
                const isActive = location.pathname === item.href;
                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`${
                      isActive 
                        ? 'bg-blue-100 text-blue-600' 
                        : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                    } group flex items-center px-2 py-2 text-sm font-medium rounded-md`}
                  >
                    <item.icon
                      className={`${
                        isActive ? 'text-blue-600' : 'text-gray-400 group-hover:text-gray-500'
                      } mr-3 flex-shrink-0 h-6 w-6`}
                      aria-hidden="true"
                    />
                    {item.name}
                  </Link>
                );
              })}
            </nav>
          </div>
          <div className="flex-shrink-0 flex border-t border-gray-200 p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                  <UserIcon className="h-6 w-6 text-gray-600" />
                </div>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-700">{user?.name || 'User'}</p>
                <p className="text-xs font-medium text-gray-500">{user?.email || 'Account'}</p>
              </div>
              <button
                type="button"
                className="ml-auto h-8 w-8 rounded-full bg-red-100 flex items-center justify-center text-red-600 hover:bg-red-200"
                onClick={handleLogout}
              >
                <ArrowLeftOnRectangleIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="md:pl-64 flex flex-col flex-1">
        {/* Top navigation */}
        <div className="md:hidden">
          <div className="bg-white shadow">
            <div className="flex items-center justify-between px-4 py-3 sm:px-6">
              <div className="flex items-center">
                <button
                  type="button"
                  className="p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
                  onClick={() => setSidebarOpen(true)}
                >
                  <span className="sr-only">Open sidebar</span>
                  <Bars3Icon className="h-6 w-6" aria-hidden="true" />
                </button>
                <div className="ml-4">
                  <h1 className="text-lg font-medium text-gray-900">
                    {location.pathname === '/' && 'Dashboard'}
                    {location.pathname === '/workflows' && 'Workflows'}
                    {location.pathname === '/workflows/new' && 'Create Workflow'}
                    {location.pathname.startsWith('/workflows/') && !location.pathname.includes('/edit') && 'Workflow Details'}
                    {location.pathname.includes('/edit') && 'Edit Workflow'}
                    {location.pathname === '/analytics' && 'Analytics'}
                    {location.pathname === '/settings' && 'Settings'}
                  </h1>
                </div>
              </div>
            </div>
          </div>
        </div>
        <main className="flex-1">
          <div className="py-6">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
              {/* Page content */}
              <Outlet />
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

export default MainLayout;