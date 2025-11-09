import { useEffect, useState } from 'react';
import { Navigate, Outlet } from 'react-router';
import { useAuth } from '~/contexts/AuthContext';

export default function GuestLayout() {
  const { user, isLoading } = useAuth();
  // Local state to trigger the redirect on the client
  const [redirectPath, setRedirectPath] = useState<string | null>(null);

  useEffect(() => {
    // This effect only runs on the client, after the initial render
    if (!isLoading && user) {
      // If a user IS found, set the redirect path to the dashboard
      setRedirectPath('/');
    }
  }, [isLoading, user]);

  // 1. If the redirect path is set, navigate
  if (redirectPath) {
    return <Navigate to={redirectPath} replace />;
  }

  // 2. Always render the outlet initially to match server render
  // This prevents hydration mismatch - both server and client show the same content
  // Auth redirect will happen after the useEffect runs
  return <Outlet />;
}
