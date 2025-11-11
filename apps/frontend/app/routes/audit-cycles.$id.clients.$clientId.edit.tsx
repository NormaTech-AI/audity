import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router';
import { ArrowLeft, Save } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '~/components/ui/card';
import { Label } from '~/components/ui/label';
import { api } from '~/api';
import type { AuditCycle, AuditCycleClient } from '~/types';

export default function EditClientPage() {
  const { id, clientId } = useParams();
  const navigate = useNavigate();
  const [cycle, setCycle] = useState<AuditCycle | null>(null);
  const [client, setClient] = useState<AuditCycleClient | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id && clientId) {
      loadData();
    }
  }, [id, clientId]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [cycleRes, clientsRes] = await Promise.all([
        api.auditCycles.getById(id!),
        api.auditCycles.getClients(id!),
      ]);

      setCycle(cycleRes.data);

      // Find the specific client
      const foundClient = clientsRes.data.find(c => c.id === clientId);
      if (foundClient) {
        setClient(foundClient);
      } else {
        setError('Client not found');
      }
    } catch (err: any) {
      console.error('Failed to load data:', err);
      setError(err.response?.data?.error || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!client) return;

    try {
      setSubmitting(true);
      setError(null);

      // Note: You'll need to implement the update API endpoint
      // For now, this is a placeholder
      console.log('Update client:', client);
      
      // Navigate back to client detail page
      navigate(`/audit-cycles/${id}/clients/${clientId}`);
    } catch (err: any) {
      console.error('Failed to update client:', err);
      setError(err.response?.data?.error || 'Failed to update client');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex items-center justify-center">
          <div className="text-lg">Loading...</div>
        </div>
      </div>
    );
  }

  if (error || !cycle || !client) {
    return (
      <div className="container mx-auto py-8">
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-destructive">{error || 'Client not found'}</p>
            <Button
              variant="outline"
              onClick={() => navigate(`/audit-cycles/${id}`)}
              className="mt-4"
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Audit Cycle
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate(`/audit-cycles/${id}/clients/${clientId}`)}
        >
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">Edit Client</h1>
          <p className="text-muted-foreground">
            {cycle.name} → {client.client_name}
          </p>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Edit Form */}
      <form onSubmit={handleSubmit}>
        <Card>
          <CardHeader>
            <CardTitle>Client Information</CardTitle>
            <CardDescription>
              Update the client information for this audit cycle
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="client_name">Client Name</Label>
              <Input
                id="client_name"
                value={client.client_name}
                onChange={(e) => setClient({ ...client, client_name: e.target.value })}
                required
                disabled
              />
              <p className="text-sm text-muted-foreground">
                Client name cannot be changed (linked to master client record)
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="poc_email">POC Email</Label>
              <Input
                id="poc_email"
                type="email"
                value={client.poc_email}
                onChange={(e) => setClient({ ...client, poc_email: e.target.value })}
                required
                disabled
              />
              <p className="text-sm text-muted-foreground">
                POC email cannot be changed (linked to master client record)
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="client_status">Status</Label>
              <select
                id="client_status"
                value={client.client_status}
                onChange={(e) => setClient({ ...client, client_status: e.target.value })}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="pending">Pending</option>
              </select>
            </div>

            <div className="flex justify-end gap-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate(`/audit-cycles/${id}/clients/${clientId}`)}
                disabled={submitting}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                <Save className="mr-2 h-4 w-4" />
                {submitting ? 'Saving...' : 'Save Changes'}
              </Button>
            </div>
          </CardContent>
        </Card>
      </form>

      {/* Information Note */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Note</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Most client information is managed at the organization level and cannot be changed here.
            Only the status within this audit cycle can be modified. To update client details like
            name or POC email, please edit the client record in the Clients section.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
