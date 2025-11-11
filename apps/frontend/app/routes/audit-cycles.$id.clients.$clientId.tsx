import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router';
import { ArrowLeft, Plus, Edit, Trash2, Calendar, User, CheckCircle2, Clock, AlertCircle } from 'lucide-react';
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
import type { AuditCycle, AuditCycleClient, AuditCycleFramework } from '~/types';

export default function ClientDetailPage() {
  const { id, clientId } = useParams();
  const navigate = useNavigate();
  const [cycle, setCycle] = useState<AuditCycle | null>(null);
  const [client, setClient] = useState<AuditCycleClient | null>(null);
  const [frameworks, setFrameworks] = useState<AuditCycleFramework[]>([]);
  const [loading, setLoading] = useState(true);
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

      const [cycleRes, clientsRes, frameworksRes] = await Promise.all([
        api.auditCycles.getById(id!),
        api.auditCycles.getClients(id!),
        api.auditCycles.getFrameworks(id!),
      ]);

      setCycle(cycleRes.data);

      // Find the specific client
      const foundClient = clientsRes.data.find(c => c.id === clientId);
      if (foundClient) {
        setClient(foundClient);
      } else {
        setError('Client not found');
      }

      // Filter frameworks for this specific client
      const clientFrameworks = frameworksRes.data.filter(
        f => f.audit_cycle_client_id === clientId
      );
      setFrameworks(clientFrameworks);
    } catch (err: any) {
      console.error('Failed to load data:', err);
      setError(err.response?.data?.error || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { variant: 'default' | 'secondary' | 'destructive' | 'outline', icon: any }> = {
      pending: { variant: 'outline', icon: Clock },
      in_progress: { variant: 'default', icon: AlertCircle },
      completed: { variant: 'secondary', icon: CheckCircle2 },
      overdue: { variant: 'destructive', icon: AlertCircle },
    };

    const config = variants[status] || variants.pending;
    const Icon = config.icon;

    return (
      <Badge variant={config.variant}>
        <Icon className="mr-1 h-3 w-3" />
        {status.replace('_', ' ')}
      </Badge>
    );
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
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate(`/audit-cycles/${id}`)}
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold">{client.client_name}</h1>
            <p className="text-muted-foreground">
              {cycle.name} → Client Details
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Link to={`/audit-cycles/${id}/clients/${clientId}/edit`}>
            <Button variant="outline">
              <Edit className="mr-2 h-4 w-4" />
              Edit Client
            </Button>
          </Link>
          <Link to={`/audit-cycles/${id}/clients/${clientId}/assign-framework`}>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Assign Framework
            </Button>
          </Link>
        </div>
      </div>

      {/* Client Information */}
      <Card>
        <CardHeader>
          <CardTitle>Client Information</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Client Name</p>
              <p className="text-lg">{client.client_name}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">POC Email</p>
              <p className="text-lg">{client.poc_email}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <Badge variant="outline" className="mt-1">{client.client_status}</Badge>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Added On</p>
              <p className="text-lg">{new Date(client.created_at).toLocaleDateString()}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Assigned Frameworks */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Assigned Frameworks</CardTitle>
              <CardDescription>
                Frameworks assigned to this client in this audit cycle
              </CardDescription>
            </div>
            <Badge variant="secondary">{frameworks.length} Total</Badge>
          </div>
        </CardHeader>
        <CardContent>
          {frameworks.length === 0 ? (
            <div className="text-center py-12">
              <div className="mx-auto h-12 w-12 text-muted-foreground mb-4">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25zM6.75 12h.008v.008H6.75V12zm0 3h.008v.008H6.75V15zm0 3h.008v.008H6.75V18z"
                  />
                </svg>
              </div>
              <h3 className="text-lg font-semibold">No frameworks assigned</h3>
              <p className="text-muted-foreground mt-2 mb-4">
                Assign frameworks to this client to start the audit process
              </p>
              <Link to={`/audit-cycles/${id}/clients/${clientId}/assign-framework`}>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Assign Framework
                </Button>
              </Link>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Framework</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Auditor</TableHead>
                  <TableHead>Due Date</TableHead>
                  <TableHead>Assigned On</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {frameworks.map((framework) => (
                  <TableRow key={framework.id}>
                    <TableCell className="font-medium">
                      {framework.framework_name}
                    </TableCell>
                    <TableCell>
                      {getStatusBadge(framework.status)}
                    </TableCell>
                    <TableCell>
                      {framework.auditor_id ? (
                        <div className="flex items-center gap-2">
                          <User className="h-4 w-4 text-muted-foreground" />
                          <span className="text-sm">Assigned</span>
                        </div>
                      ) : (
                        <span className="text-sm text-muted-foreground">Not assigned</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {framework.due_date ? (
                        <div className="flex items-center gap-2">
                          <Calendar className="h-4 w-4 text-muted-foreground" />
                          {new Date(framework.due_date).toLocaleDateString()}
                        </div>
                      ) : (
                        <span className="text-sm text-muted-foreground">No due date</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {new Date(framework.assigned_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button variant="ghost" size="icon" title="Edit Framework">
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" title="Remove Framework">
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Framework Statistics */}
      {frameworks.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Total Frameworks</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{frameworks.length}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Pending</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {frameworks.filter(f => f.status === 'pending').length}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">In Progress</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {frameworks.filter(f => f.status === 'in_progress').length}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Completed</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {frameworks.filter(f => f.status === 'completed').length}
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
