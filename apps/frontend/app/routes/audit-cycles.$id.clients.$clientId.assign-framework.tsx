import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router';
import { ArrowLeft, Plus, FileText } from 'lucide-react';
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
import type { AuditCycle, AuditCycleClient, Framework, AuditCycleFramework, User } from '~/types';

export default function AssignFrameworkPage() {
  const { id, clientId } = useParams();
  const navigate = useNavigate();
  const [cycle, setCycle] = useState<AuditCycle | null>(null);
  const [client, setClient] = useState<AuditCycleClient | null>(null);
  const [frameworks, setFrameworks] = useState<Framework[]>([]);
  const [assignedFrameworks, setAssignedFrameworks] = useState<AuditCycleFramework[]>([]);
  const [availableFrameworks, setAvailableFrameworks] = useState<Framework[]>([]);
  const [selectedFrameworks, setSelectedFrameworks] = useState<Set<string>>(new Set());
  const [users, setUsers] = useState<User[]>([]);
  const [auditors, setAuditors] = useState<User[]>([]);
  const [selectedAuditor, setSelectedAuditor] = useState<string>('');
  const [dueDate, setDueDate] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
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
      
      const [cycleRes, clientsRes, frameworksRes, assignedFrameworksRes, usersRes] = await Promise.all([
        api.auditCycles.getById(id!),
        api.auditCycles.getClients(id!),
        api.frameworks.list(),
        api.auditCycles.getFrameworks(id!),
        api.users.listAuditors(), // Fetch all users to show as potential auditors
      ]);

      setCycle(cycleRes.data);
      setFrameworks(frameworksRes.data);
      setUsers(usersRes.data.data || []);
      
      // Filter users who are auditors (have auditor role)
      // For now, show all users - can be filtered by role later
      setAuditors(usersRes.data.data || []);
      
      // Find the specific client
      const foundClient = clientsRes.data.find(c => c.id === clientId);
      if (foundClient) {
        setClient(foundClient);
        
        // Filter assigned frameworks for this specific client
        const clientAssignedFrameworks = assignedFrameworksRes.data.filter(
          f => f.audit_cycle_client_id === clientId
        );
        setAssignedFrameworks(clientAssignedFrameworks);
        
        // Filter out already assigned frameworks
        const assignedFrameworkIds = new Set(clientAssignedFrameworks.map(f => f.framework_id));
        const available = frameworksRes.data.filter(
          framework => !assignedFrameworkIds.has(framework.id)
        );
        setAvailableFrameworks(available);
      } else {
        setError('Client not found in this audit cycle');
      }
    } catch (err: any) {
      console.error('Failed to load data:', err);
      setError(err.response?.data?.error || 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleFramework = (frameworkId: string) => {
    const newSelected = new Set(selectedFrameworks);
    if (newSelected.has(frameworkId)) {
      newSelected.delete(frameworkId);
    } else {
      newSelected.add(frameworkId);
    }
    setSelectedFrameworks(newSelected);
  };

  const handleAssignFrameworks = async () => {
    if (selectedFrameworks.size === 0) {
      alert('Please select at least one framework');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);

      // Assign each selected framework
      const promises = Array.from(selectedFrameworks).map(frameworkId => {
        const framework = frameworks.find(f => f.id === frameworkId);
        if (!framework) return Promise.resolve();
        
        return api.auditCycles.assignFramework(clientId!, {
          framework_id: frameworkId,
          framework_name: framework.name,
          due_date: dueDate || undefined,
          auditor_id: selectedAuditor || undefined,
        });
      });

      await Promise.all(promises);
      
      // Navigate back to the audit cycle detail page
      navigate(`/audit-cycles/${id}`);
    } catch (err: any) {
      console.error('Failed to assign frameworks:', err);
      setError(err.response?.data?.error || 'Failed to assign frameworks');
    } finally {
      setSubmitting(false);
    }
  };

  const filteredFrameworks = availableFrameworks.filter(framework =>
    framework.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (framework.description && framework.description.toLowerCase().includes(searchQuery.toLowerCase()))
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
        <Card className="border-destructive">
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
          <h1 className="text-3xl font-bold tracking-tight">Assign Frameworks</h1>
          <p className="text-muted-foreground mt-2">
            {cycle?.name} → {client?.client_name}
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

      {/* Assignment Details */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <Label htmlFor="auditor">Assign Auditor (Optional)</Label>
              <select
                id="auditor"
                value={selectedAuditor}
                onChange={(e) => setSelectedAuditor(e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="">No auditor assigned</option>
                {auditors.map((auditor) => (
                  <option key={auditor.id} value={auditor.id}>
                    {auditor.name} ({auditor.email})
                  </option>
                ))}
              </select>
              <p className="text-sm text-muted-foreground">
                Select an auditor to review these frameworks
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="due_date">Due Date (Optional)</Label>
              <Input
                id="due_date"
                type="date"
                value={dueDate}
                onChange={(e) => setDueDate(e.target.value)}
                placeholder="Set a due date for all selected frameworks"
              />
              <p className="text-sm text-muted-foreground">
                This due date will be applied to all selected frameworks
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-2">
            <Label htmlFor="search">Search Frameworks</Label>
            <Input
              id="search"
              placeholder="Search by name or description..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Available Frameworks */}
      <Card>
        <CardHeader>
          <CardTitle>Available Frameworks</CardTitle>
          <CardDescription>
            {selectedFrameworks.size} framework{selectedFrameworks.size !== 1 ? 's' : ''} selected
          </CardDescription>
        </CardHeader>
        <CardContent>
          {filteredFrameworks.length === 0 ? (
            <div className="text-center py-8">
              <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No frameworks found</h3>
              <p className="text-muted-foreground mt-2">
                {searchQuery
                  ? 'Try adjusting your search query'
                  : 'No frameworks available'}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <input
                      type="checkbox"
                      checked={selectedFrameworks.size === filteredFrameworks.length && filteredFrameworks.length > 0}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedFrameworks(new Set(filteredFrameworks.map(f => f.id)));
                        } else {
                          setSelectedFrameworks(new Set());
                        }
                      }}
                      className="h-4 w-4"
                    />
                  </TableHead>
                  <TableHead>Framework Name</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Questions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredFrameworks.map((framework) => (
                  <TableRow
                    key={framework.id}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleToggleFramework(framework.id)}
                  >
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={selectedFrameworks.has(framework.id)}
                        onChange={() => handleToggleFramework(framework.id)}
                        onClick={(e) => e.stopPropagation()}
                        className="h-4 w-4"
                      />
                    </TableCell>
                    <TableCell className="font-medium">{framework.name}</TableCell>
                    <TableCell className="max-w-md truncate">
                      {framework.description || '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{framework.version}</Badge>
                    </TableCell>
                    <TableCell>{framework.question_count || 0}</TableCell>
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
          onClick={handleAssignFrameworks}
          disabled={submitting || selectedFrameworks.size === 0}
        >
          {submitting ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Assigning...
            </>
          ) : (
            <>
              <Plus className="mr-2 h-4 w-4" />
              Assign {selectedFrameworks.size} Framework{selectedFrameworks.size !== 1 ? 's' : ''}
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
