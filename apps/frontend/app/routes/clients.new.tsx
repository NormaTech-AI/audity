import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router';
import { ArrowLeft, Save } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { Label } from '~/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { api } from '~/api';

interface CreateClientForm {
  name: string;
  poc_email: string;
  email_domain: string;
}

export default function CreateClientPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  
  const [formData, setFormData] = useState<CreateClientForm>({
    name: '',
    poc_email: '',
    email_domain: '',
  });

  const [errors, setErrors] = useState<Partial<CreateClientForm>>({});

  const createMutation = useMutation({
    mutationFn: async (data: CreateClientForm) => {
      const response = await api.clients.create(data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clients'] });
      navigate('/clients');
    },
    onError: (error: any) => {
      console.error('Failed to create client:', error);
      // Handle error display
    },
  });

  const validateForm = (): boolean => {
    const newErrors: Partial<CreateClientForm> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Client name is required';
    }

    if (!formData.poc_email.trim()) {
      newErrors.poc_email = 'POC email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.poc_email)) {
      newErrors.poc_email = 'Invalid email format';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (validateForm()) {
      createMutation.mutate(formData);
    }
  };

  const handleChange = (field: keyof CreateClientForm, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error for this field when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate('/clients')}
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create New Client</h1>
          <p className="text-muted-foreground mt-2">
            Add a new client organization to the system
          </p>
        </div>
      </div>

      {/* Form */}
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Client Information</CardTitle>
          <CardDescription>
            Enter the basic information for the new client. A dedicated database and storage bucket will be automatically provisioned.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Client Name */}
            <div className="space-y-2">
              <Label htmlFor="name">
                Client Name <span className="text-red-500">*</span>
              </Label>
              <Input
                id="name"
                placeholder="e.g., Acme Corporation"
                value={formData.name}
                onChange={(e) => handleChange('name', e.target.value)}
                className={errors.name ? 'border-red-500' : ''}
              />
              {errors.name && (
                <p className="text-sm text-red-500">{errors.name}</p>
              )}
            </div>

            {/* POC Email */}
            <div className="space-y-2">
              <Label htmlFor="poc_email">
                Point of Contact Email <span className="text-red-500">*</span>
              </Label>
              <Input
                id="poc_email"
                type="email"
                placeholder="e.g., contact@acme.com"
                value={formData.poc_email}
                onChange={(e) => handleChange('poc_email', e.target.value)}
                className={errors.poc_email ? 'border-red-500' : ''}
              />
              {errors.poc_email && (
                <p className="text-sm text-red-500">{errors.poc_email}</p>
              )}
              <p className="text-sm text-muted-foreground">
                This email will be used as the primary contact for this client
              </p>
            </div>

            {/* Email Domain */}
            <div className="space-y-2">
              <Label htmlFor="email_domain">
                Email Domain
              </Label>
              <Input
                id="email_domain"
                placeholder="e.g., acme.com"
                value={formData.email_domain}
                onChange={(e) => handleChange('email_domain', e.target.value)}
                className={errors.email_domain ? 'border-red-500' : ''}
              />
              {errors.email_domain && (
                <p className="text-sm text-red-500">{errors.email_domain}</p>
              )}
              <p className="text-sm text-muted-foreground">
                The primary email domain for this client organization (without @)
              </p>
            </div>

            {/* Info Box */}
            <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
              <h4 className="font-medium text-blue-900 dark:text-blue-100 mb-2">
                What happens when you create a client?
              </h4>
              <ul className="text-sm text-blue-800 dark:text-blue-200 space-y-1 list-disc list-inside">
                <li>A dedicated PostgreSQL database will be provisioned</li>
                <li>A dedicated MinIO storage bucket will be created</li>
                <li>Database credentials will be securely stored</li>
                <li>The client will be set to "active" status by default</li>
              </ul>
            </div>

            {/* Error Message */}
            {createMutation.isError && (
              <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4">
                <p className="text-sm text-red-800 dark:text-red-200">
                  Failed to create client. Please try again or contact support if the issue persists.
                </p>
              </div>
            )}

            {/* Actions */}
            <div className="flex gap-3 pt-4">
              <Button
                type="submit"
                disabled={createMutation.isPending}
              >
                {createMutation.isPending ? (
                  <>
                    <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    Creating...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 h-4 w-4" />
                    Create Client
                  </>
                )}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate('/clients')}
                disabled={createMutation.isPending}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
