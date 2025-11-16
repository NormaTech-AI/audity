import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router';
import { ArrowLeft, Plus, Users } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '~/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '~/components/ui/table';
import { Badge } from '~/components/ui/badge';
import { Label } from '~/components/ui/label';
import { api } from '~/api';
import type { AuditCycle, Client, AuditCycleClient } from '~/types';

export default function AddClientToAuditCyclePage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [cycle, setCycle] = useState<AuditCycle | null>(null);
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [existingClients, setExistingClients] = useState<AuditCycleClient[]>([]);
  const [availableClients, setAvailableClients] = useState<Client[]>([]);
  const [selectedClients, setSelectedClients] = useState<Set<string>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      loadData();
    }
  }, [id]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [cycleRes, clientsRes, existingClientsRes] = await Promise.all([
        api.auditCycles.getById(id!),
        api.clients.list(),
        api.auditCycles.getClients(id!),
      ]);

      setCycle(cycleRes.data);
      setAllClients(clientsRes.data || []);
      setExistingClients(existingClientsRes.data);

      // Filter out clients that are already in the cycle
      const existingClientIds = new Set(existingClientsRes.data.map(c => c.client_id));
      console.log(existingClientIds)
      const available = (clientsRes.data || []).filter(
        client => !existingClientIds.has(client.id)
      );
      console.log(allClients)
      setAvailableClients(available);
    } catch (err: any) {
      console.error('Failed to load data:', err);
      setError(err.response?.data?.error || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleClient = (clientId: string) => {
    const newSelected = new Set(selectedClients);
    if (newSelected.has(clientId)) {
      newSelected.delete(clientId);
    } else {
      newSelected.add(clientId);
    }
    setSelectedClients(newSelected);
  };

  const handleAddClients = async () => {
    if (selectedClients.size === 0) {
      setError('Please select at least one client');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);

      // Add each selected client
      const promises = Array.from(selectedClients).map(clientId =>
        api.auditCycles.addClient(id!, clientId)
      );

      await Promise.all(promises);
      
      // Navigate back to the audit cycle detail page
      navigate(`/audit-cycles/${id}`);
    } catch (err: any) {
      console.error('Failed to add clients:', err);
      setError(err.response?.data?.error || 'Failed to add clients');
    } finally {
      setSubmitting(false);
    }
  };

  const filteredClients = availableClients.filter(client =>
    client.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    client.poc_email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  if (error && !cycle) {
    return (
      <div className="container mx-auto py-6">
        <Card className="border-destructive" role="alert">
          <CardContent className="pt-6">
            <p className="text-destructive">{error}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate(`/audit-cycles/${id}`)}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <h1 className="text-3xl font-bold tracking-tight">Add Clients to Audit Cycle</h1>
          <p className="text-muted-foreground mt-2">
            {cycle?.name || 'Loading...'}
          </p>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive" role="alert">
          <CardContent className="pt-6">
            <p className="text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-2">
            <Label htmlFor="search">Search Clients</Label>
            <Input
              id="search"
              placeholder="Search by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Available Clients */}
      <Card>
        <CardHeader>
          <CardTitle>Available Clients</CardTitle>
          <CardDescription>
            {selectedClients.size} client{selectedClients.size !== 1 ? 's' : ''} selected
          </CardDescription>
        </CardHeader>
        <CardContent>
          {filteredClients.length === 0 ? (
            <div className="text-center py-8">
              <Users className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No available clients</h3>
              <p className="text-muted-foreground mt-2">
                {searchQuery
                  ? 'Try adjusting your search query'
                  : 'All clients are already added to this audit cycle'}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <input
                      type="checkbox"
                      checked={selectedClients.size === filteredClients.length && filteredClients.length > 0}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedClients(new Set(filteredClients.map(c => c.id)));
                        } else {
                          setSelectedClients(new Set());
                        }
                      }}
                      className="h-4 w-4"
                    />
                  </TableHead>
                  <TableHead>Client Name</TableHead>
                  <TableHead>POC Email</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredClients.map((client) => (
                  <TableRow
                    key={client.id}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleToggleClient(client.id)}
                  >
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={selectedClients.has(client.id)}
                        onChange={() => handleToggleClient(client.id)}
                        onClick={(e) => e.stopPropagation()}
                        className="h-4 w-4"
                      />
                    </TableCell>
                    <TableCell className="font-medium">{client.name}</TableCell>
                    <TableCell>{client.poc_email}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{client.status}</Badge>
                    </TableCell>
                    <TableCell>
                      {new Date(client.created_at).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Actions */}
      <div className="flex justify-end gap-4">
        <Button
          variant="outline"
          onClick={() => navigate(`/audit-cycles/${id}`)}
          disabled={submitting}
        >
          Cancel
        </Button>
        <Button
          onClick={handleAddClients}
          disabled={submitting || selectedClients.size === 0}
        >
          {submitting ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Adding...
            </>
          ) : (
            <>
              <Plus className="mr-2 h-4 w-4" />
              Add {selectedClients.size} Client{selectedClients.size !== 1 ? 's' : ''}
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
