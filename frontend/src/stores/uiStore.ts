import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { UIState, Notification } from '@/types/node';

interface UIStore extends UIState {
  // Mobile menu state
  mobileMenuOpen: boolean;

  // Actions
  setSidebarCollapsed: (collapsed: boolean) => void;
  setMobileMenuOpen: (open: boolean) => void;
  toggleMobileMenu: () => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  setLanguage: (language: string) => void;
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => void;
  removeNotification: (id: string) => void;
  markNotificationAsRead: (id: string) => void;
  clearAllNotifications: () => void;
  toggleSidebar: () => void;
}

export const useUIStore = create<UIStore>()(
  persist(
    (set, get) => ({
      // Initial state
      sidebarCollapsed: false,
      mobileMenuOpen: false,
      theme: 'system',
      language: 'en',
      notifications: [],

      // Actions
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      setMobileMenuOpen: (open) => set({ mobileMenuOpen: open }),

      toggleMobileMenu: () => set((state) => ({
        mobileMenuOpen: !state.mobileMenuOpen
      })),

      setTheme: (theme) => set({ theme }),

      setLanguage: (language) => set({ language }),

      addNotification: (notification) => set((state) => ({
        notifications: [
          {
            ...notification,
            id: `notification_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            timestamp: new Date(),
            read: false
          },
          ...state.notifications
        ].slice(0, 50) // Keep only last 50 notifications
      })),

      removeNotification: (id) => set((state) => ({
        notifications: state.notifications.filter(n => n.id !== id)
      })),

      markNotificationAsRead: (id) => set((state) => ({
        notifications: state.notifications.map(n =>
          n.id === id ? { ...n, read: true } : n
        )
      })),

      clearAllNotifications: () => set({ notifications: [] }),

      toggleSidebar: () => set((state) => ({
        sidebarCollapsed: !state.sidebarCollapsed
      }))
    }),
    {
      name: 'ui-storage',
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
        language: state.language
      })
    }
  )
);