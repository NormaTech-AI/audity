import { createContext, useContext, type ReactNode } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '~/api';
import type { User } from '~/types';

export interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (token: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();

  // Fetch the user session using TanStack Query
  const { data: user, isLoading } = useQuery({
    queryKey: ['auth', 'user'],
    queryFn: async () => {
      // Only run on client-side
      if (typeof window === 'undefined') {
        return null;
      }
      const response = await api.auth.validateToken();
      return response.data;
    },
    retry: false, // Don't retry on failure; a 401 just means the user is not logged in
    refetchOnWindowFocus: true,
    enabled: typeof window !== 'undefined', // Only enable query on client-side
  });

  // Logout function
  const logout = () => {
    api.auth.logout(); // Call the logout API endpoint
    queryClient.setQueryData(['auth', 'user'], null); // Immediately clear the user data in the cache
    // Clear token from localStorage
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
    }
  };

  // Login function
  const login = async (token: string) => {
    try {
      // Store token in localStorage for Authorization header
      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_token', token);
      }
      
      // The token is also handled by the backend via HTTP-only cookie
      // Validate the session and update cache
      if (typeof window === 'undefined') {
        return; // Don't run login on server-side
      }
      const response = await api.auth.validateToken();
      queryClient.setQueryData(['auth', 'user'], response.data);
    } catch (error) {
      console.error('Login failed:', error);
      // Clear token on error
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
      }
      throw error;
    }
  };

  const isAuthenticated = !!user;

  const value: AuthContextType = {
    user: user || null,
    isLoading,
    login,
    logout,
    isAuthenticated,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
