import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router';
import { useQueryClient } from '@tanstack/react-query';
import { api } from '~/api';

export default function AuthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  useEffect(() => {
    const token = searchParams.get('token');
    
    if (token) {
      handleCallback(token);
    } else {
      // No token, redirect to login
      navigate('/login');
    }
  }, [searchParams]);

  const handleCallback = async (token: string) => {
    try {
      // Store token in localStorage for Authorization header
      localStorage.setItem('auth_token', token);
      
      // Send the token to backend to set it as HTTP-only cookie
      const response = await api.auth.setTokenCookie(token);
      
      // Update React Query cache with user data
      queryClient.setQueryData(['auth', 'user'], response.data);
      
      // Navigate to dashboard
      navigate('/');
    } catch (error) {
      console.error('OAuth callback failed:', error);
      // Clear token on error
      localStorage.removeItem('auth_token');
      navigate('/login');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
        <p className="text-muted-foreground">Completing sign in...</p>
      </div>
    </div>
  );
}
