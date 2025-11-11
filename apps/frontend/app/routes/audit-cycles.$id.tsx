import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router';
import { ArrowLeft, Users, FileText, BarChart3, Plus, Eye, Edit } from 'lucide-react';
import { Button } from '~/components/ui/button';
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
import { api } from '~/api';
import type { AuditCycle, AuditCycleClient, AuditCycleStats } from '~/types';

export default function AuditCycleDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [cycle, setCycle] = useState<AuditCycle | null>(null);
  const [stats, setStats] = useState<AuditCycleStats | null>(null);
  const [clients, setClients] = useState<AuditCycleClient[]>([]);
  const [loading, setLoading] = useState(true);
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
      
      const [cycleRes, statsRes, clientsRes] = await Promise.all([
        api.auditCycles.getById(id!),
        api.auditCycles.getStats(id!),
        api.auditCycles.getClients(id!),
      ]);

      setCycle(cycleRes.data);
      setStats(statsRes.data);
      setClients(clientsRes.data);
    } catch (err: any) {
      console.error('Failed to load audit cycle:', err);
      setError(err.response?.data?.error || 'Failed to load audit cycle');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'active':
        return 'default';
      case 'completed':
        return 'secondary';
      case 'archived':
        return 'outline';
      default:
        return 'default';
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  if (error || !cycle) {
    return (
      <div className="container mx-auto py-6">
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-destructive">{error || 'Audit cycle not found'}</p>
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
          onClick={() => navigate('/audit-cycles')}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <h1 className="text-3xl font-bold tracking-tight">{cycle.name}</h1>
          <p className="text-muted-foreground mt-2">
            {cycle.description || 'No description'}
          </p>
        </div>
        <Link to={`/audit-cycles/${id}/edit`}>
          <Button variant="outline">Edit Cycle</Button>
        </Link>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Clients</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_clients}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Frameworks</CardTitle>
              <FileText className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_frameworks}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Completed</CardTitle>
              <BarChart3 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.completed_frameworks}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Status</CardTitle>
            </CardHeader>
            <CardContent>
              <Badge variant={getStatusBadgeVariant(cycle.status)}>
                {cycle.status}
              </Badge>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Cycle Info */}
      <Card>
        <CardHeader>
          <CardTitle>Cycle Information</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-muted-foreground">Start Date</p>
            <p className="text-lg font-medium">
              {new Date(cycle.start_date).toLocaleDateString()}
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">End Date</p>
            <p className="text-lg font-medium">
              {new Date(cycle.end_date).toLocaleDateString()}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Clients Table */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Clients in Audit Cycle</CardTitle>
              <CardDescription>
                {clients.length} client{clients.length !== 1 ? 's' : ''} assigned
              </CardDescription>
            </div>
            <Link to={`/audit-cycles/${id}/add-client`}>
              <Button>Add Client</Button>
            </Link>
          </div>
        </CardHeader>
        <CardContent>
          {clients.length === 0 ? (
            <div className="text-center py-8">
              <Users className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No clients assigned</h3>
              <p className="text-muted-foreground mt-2">
                Add clients to this audit cycle to get started
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Client Name</TableHead>
                  <TableHead>POC Email</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Added On</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {clients.map((client) => (
                  <TableRow key={client.id}>
                    <TableCell className="font-medium">{client.client_name}</TableCell>
                    <TableCell>{client.poc_email}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{client.client_status}</Badge>
                    </TableCell>
                    <TableCell>
                      {new Date(client.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Link to={`/audit-cycles/${id}/clients/${client.id}`}>
                          <Button variant="ghost" size="icon" title="View Details">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Link to={`/audit-cycles/${id}/clients/${client.id}/edit`}>
                          <Button variant="ghost" size="icon" title="Edit Client">
                            <Edit className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Link to={`/audit-cycles/${id}/clients/${client.id}/assign-framework`}>
                          <Button variant="outline" size="sm">
                            <Plus className="mr-2 h-4 w-4" />
                            Assign Framework
                          </Button>
                        </Link>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
