import { useEffect, useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router';
import { useAuth } from '~/contexts/AuthContext';
import { DashboardLayout } from '~/components/layout/DashboardLayout';

export default function ProtectedLayout() {
  const { user, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  // Local state to trigger the redirect on the client
  const [redirectPath, setRedirectPath] = useState<string | null>(null);

  useEffect(() => {
    // This effect only runs on the client, after the initial render
    if (!isLoading && !user) {
      // If the user is not authenticated, set the redirect path
      setRedirectPath('/login');
    }
  }, [isLoading, user]);

  // 1. If the redirect path is set, navigate
  if (redirectPath !== null) {
    return navigate(redirectPath, { state: { from: location }, replace: true });
  }

  // 2. Always render the protected content initially to match server render
  // This prevents hydration mismatch - both server and client show the same content
  // Auth redirect will happen after the useEffect runs
  return (
    <DashboardLayout>
      <Outlet />
    </DashboardLayout>
  );
}
