import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { Building2, Chrome } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { useAuth } from '~/contexts/AuthContext';
import { api } from '~/api';

export default function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login, isAuthenticated } = useAuth();

  // Handle OAuth callback
  useEffect(() => {
    const token = searchParams.get('token');
    if (token) {
      handleOAuthCallback(token);
    }
  }, [searchParams]);

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard');
    }
  }, [isAuthenticated, navigate]);

  const handleOAuthCallback = async (token: string) => {
    try {
      await login(token);
    } catch (error) {
      console.error('OAuth callback failed:', error);
      // Show error toast
    }
  };

  const handleGoogleLogin = async () => {
    try {
      const response = await api.auth.initiateGoogleLogin();
      window.location.href = response.data.auth_url;
    } catch (error) {
      console.error('Failed to initiate Google login:', error);
      // Show error toast
    }
  };

  const handleMicrosoftLogin = async () => {
    try {
      const response = await api.auth.initiateMicrosoftLogin();
      window.location.href = response.data.auth_url;
    } catch (error) {
      console.error('Failed to initiate Microsoft login:', error);
      // Show error toast
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900 p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-4 text-center">
          <div className="mx-auto w-12 h-12 bg-primary rounded-lg flex items-center justify-center">
            <Building2 className="w-6 h-6 text-primary-foreground" />
          </div>
          <div>
            <CardTitle className="text-2xl">Welcome to Audity</CardTitle>
            <CardDescription className="mt-2">
              Third-Party Risk Management Platform
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3">
            <Button
              onClick={handleGoogleLogin}
              variant="outline"
              className="w-full"
              size="lg"
            >
              <Chrome className="mr-2 h-5 w-5" />
              Continue with Google
            </Button>
            
            <Button
              onClick={handleMicrosoftLogin}
              variant="outline"
              className="w-full"
              size="lg"
            >
              <svg
                className="mr-2 h-5 w-5"
                viewBox="0 0 23 23"
                fill="currentColor"
              >
                <path d="M0 0h11v11H0z" fill="#f25022" />
                <path d="M12 0h11v11H12z" fill="#00a4ef" />
                <path d="M0 12h11v11H0z" fill="#7fba00" />
                <path d="M12 12h11v11H12z" fill="#ffb900" />
              </svg>
              Continue with Microsoft
            </Button>
          </div>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">
                Secure Authentication
              </span>
            </div>
          </div>

          <div className="text-center text-sm text-muted-foreground">
            <p>
              By continuing, you agree to our Terms of Service and Privacy Policy.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
