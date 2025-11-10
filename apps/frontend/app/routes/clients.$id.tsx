import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router';
import { ArrowLeft, Save, Trash2, Power, PowerOff, Database, HardDrive } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { Label } from '~/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { api } from '~/api';
import type { Client, ClientStatus } from '~/types';

interface UpdateClientForm {
  name: string;
  poc_email: string;
  email_domain: string;
  status: ClientStatus;
}

export default function ClientDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState<UpdateClientForm>({
    name: '',
    poc_email: '',
    email_domain: '',
    status: 'active',
  });
  const [errors, setErrors] = useState<Partial<UpdateClientForm>>({});

  // Fetch client details
  const { data: client, isLoading, error } = useQuery({
    queryKey: ['clients', id],
    queryFn: async () => {
      const response = await api.clients.getById(id!);
      return response.data;
    },
    enabled: !!id,
  });

  // Initialize form when client data loads
  useEffect(() => {
    if (client) {
      setFormData({
        name: client.name,
        poc_email: client.poc_email,
        email_domain: client.email_domain || '',
        status: client.status,
      });
    }
  }, [client]);

  // Update client mutation
  const updateMutation = useMutation({
    mutationFn: async (data: UpdateClientForm) => {
      const response = await api.clients.update(id!, data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clients', id] });
      queryClient.invalidateQueries({ queryKey: ['clients'] });
      setIsEditing(false);
    },
    onError: (error: any) => {
      console.error('Failed to update client:', error);
    },
  });

  // Delete client mutation
  const deleteMutation = useMutation({
    mutationFn: async () => {
      await api.clients.delete(id!);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clients'] });
      navigate('/clients');
    },
    onError: (error: any) => {
      console.error('Failed to delete client:', error);
    },
  });

  const validateForm = (): boolean => {
    const newErrors: Partial<UpdateClientForm> = {};

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
      updateMutation.mutate(formData);
    }
  };

  const handleChange = (field: keyof UpdateClientForm, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleCancel = () => {
    if (client) {
      setFormData({
        name: client.name,
        poc_email: client.poc_email,
        email_domain: client.email_domain || '',
        status: client.status,
      });
    }
    setErrors({});
    setIsEditing(false);
  };

  const handleDelete = () => {
    if (window.confirm('Are you sure you want to delete this client? This action cannot be undone.')) {
      deleteMutation.mutate();
    }
  };

  const toggleStatus = () => {
    const newStatus = client?.status === 'active' ? 'inactive' : 'active';
    updateMutation.mutate({
      ...formData,
      status: newStatus,
    });
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/clients')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
        </div>
        <Card className="animate-pulse">
          <CardHeader>
            <div className="h-8 bg-muted rounded w-1/3" />
            <div className="h-4 bg-muted rounded w-1/4 mt-2" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="h-4 bg-muted rounded w-full" />
              <div className="h-4 bg-muted rounded w-full" />
              <div className="h-4 bg-muted rounded w-2/3" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error || !client) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/clients')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <p className="text-lg font-medium text-red-600">Client not found</p>
            <p className="text-sm text-muted-foreground mt-1">
              The client you're looking for doesn't exist or has been deleted.
            </p>
            <Button onClick={() => navigate('/clients')} className="mt-4">
              Go to Clients
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/clients')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{client.name}</h1>
            <p className="text-muted-foreground mt-2">
              Client ID: {client.id}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          {!isEditing && (
            <>
              <Button
                variant="outline"
                onClick={toggleStatus}
                disabled={updateMutation.isPending}
              >
                {client.status === 'active' ? (
                  <>
                    <PowerOff className="h-4 w-4 mr-2" />
                    Deactivate
                  </>
                ) : (
                  <>
                    <Power className="h-4 w-4 mr-2" />
                    Activate
                  </>
                )}
              </Button>
              <Button onClick={() => setIsEditing(true)}>
                Edit Client
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Status Badge */}
      <div className="flex items-center gap-2">
        <span
          className={`px-3 py-1 rounded-full text-sm font-medium ${
            client.status === 'active'
              ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
              : client.status === 'inactive'
              ? 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
              : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
          }`}
        >
          {client.status.charAt(0).toUpperCase() + client.status.slice(1)}
        </span>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Client Details */}
        <Card>
          <CardHeader>
            <CardTitle>Client Information</CardTitle>
            <CardDescription>
              Basic information about this client
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isEditing ? (
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">
                    Client Name <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => handleChange('name', e.target.value)}
                    className={errors.name ? 'border-red-500' : ''}
                  />
                  {errors.name && (
                    <p className="text-sm text-red-500">{errors.name}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="poc_email">
                    POC Email <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="poc_email"
                    type="email"
                    value={formData.poc_email}
                    onChange={(e) => handleChange('poc_email', e.target.value)}
                    className={errors.poc_email ? 'border-red-500' : ''}
                  />
                  {errors.poc_email && (
                    <p className="text-sm text-red-500">{errors.poc_email}</p>
                  )}
                </div>

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
                </div>

                <div className="space-y-2">
                  <Label htmlFor="status">Status</Label>
                  <select
                    id="status"
                    value={formData.status}
                    onChange={(e) => handleChange('status', e.target.value as any)}
                    className="w-full px-3 py-2 border rounded-md"
                  >
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                    <option value="suspended">Suspended</option>
                  </select>
                </div>

                {updateMutation.isError && (
                  <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-3">
                    <p className="text-sm text-red-800 dark:text-red-200">
                      Failed to update client. Please try again.
                    </p>
                  </div>
                )}

                <div className="flex gap-2 pt-2">
                  <Button
                    type="submit"
                    disabled={updateMutation.isPending}
                  >
                    {updateMutation.isPending ? (
                      <>
                        <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                        Saving...
                      </>
                    ) : (
                      <>
                        <Save className="mr-2 h-4 w-4" />
                        Save Changes
                      </>
                    )}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleCancel}
                    disabled={updateMutation.isPending}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            ) : (
              <div className="space-y-4">
                <div>
                  <Label className="text-muted-foreground">Client Name</Label>
                  <p className="text-lg font-medium">{client.name}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">POC Email</Label>
                  <p className="text-lg">{client.poc_email}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Email Domain</Label>
                  <p className="text-lg">{client.email_domain || 'Not specified'}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Created</Label>
                  <p className="text-lg">
                    {new Date(client.created_at).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Last Updated</Label>
                  <p className="text-lg">
                    {new Date(client.updated_at).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Infrastructure */}
        <Card>
          <CardHeader>
            <CardTitle>Infrastructure</CardTitle>
            <CardDescription>
              Provisioned resources for this client
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-start gap-3 p-3 border rounded-lg">
              <Database className="h-5 w-5 text-blue-600 mt-0.5" />
              <div className="flex-1">
                <p className="font-medium">PostgreSQL Database</p>
                <p className="text-sm text-muted-foreground">
                  Dedicated isolated database
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  Database: client_{client.id.split('-')[0]}
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3 p-3 border rounded-lg">
              <HardDrive className="h-5 w-5 text-green-600 mt-0.5" />
              <div className="flex-1">
                <p className="font-medium">MinIO Storage Bucket</p>
                <p className="text-sm text-muted-foreground">
                  Object storage for files and documents
                </p>
                <p className="text-xs text-muted-foreground mt-1">
                  Bucket: client-{client.id.split('-')[0]}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Danger Zone */}
      {!isEditing && (
        <Card className="border-red-200 dark:border-red-800">
          <CardHeader>
            <CardTitle className="text-red-600">Danger Zone</CardTitle>
            <CardDescription>
              Irreversible actions for this client
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Delete this client</p>
                <p className="text-sm text-muted-foreground">
                  This will permanently delete the client and all associated data
                </p>
              </div>
              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? (
                  <>
                    <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    Deleting...
                  </>
                ) : (
                  <>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete Client
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
