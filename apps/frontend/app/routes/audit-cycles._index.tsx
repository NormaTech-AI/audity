import { useState, useEffect } from 'react';
import { Link } from 'react-router';
import { Plus, Calendar, Search, Trash2, Edit, Eye, BarChart3 } from 'lucide-react';
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
import { api } from '~/api';
import type { AuditCycle } from '~/types';

export default function AuditCyclesPage() {
  const [auditCycles, setAuditCycles] = useState<AuditCycle[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAuditCycles();
  }, []);

  const loadAuditCycles = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.auditCycles.list();
      setAuditCycles(response.data);
    } catch (err: any) {
      console.error('Failed to load audit cycles:', err);
      setError(err.response?.data?.error || 'Failed to load audit cycles');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this audit cycle?')) {
      return;
    }

    try {
      await api.auditCycles.delete(id);
      setAuditCycles(auditCycles.filter((ac) => ac.id !== id));
    } catch (err: any) {
      console.error('Failed to delete audit cycle:', err);
      setError(err.response?.data?.error || 'Failed to delete audit cycle');
    }
  };

  const filteredAuditCycles = auditCycles.filter((cycle) =>
    cycle.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (cycle.description && cycle.description.toLowerCase().includes(searchQuery.toLowerCase()))
  );

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

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Audit Cycles</h1>
          <p className="text-muted-foreground mt-2">
            Manage audit cycles and assign frameworks to clients
          </p>
        </div>
        <Link to="/audit-cycles/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Create Audit Cycle
          </Button>
        </Link>
      </div>

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search audit cycles..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </CardContent>
      </Card>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive" role="alert">
          <CardContent className="pt-6">
            <p className="text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Audit Cycles Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Audit Cycles</CardTitle>
          <CardDescription>
            {filteredAuditCycles.length} audit cycle{filteredAuditCycles.length !== 1 ? 's' : ''} found
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : filteredAuditCycles.length === 0 ? (
            <div className="text-center py-8">
              <Calendar className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No audit cycles found</h3>
              <p className="text-muted-foreground mt-2">
                {searchQuery
                  ? 'Try adjusting your search query'
                  : 'Get started by creating your first audit cycle'}
              </p>
              {!searchQuery && (
                <Link to="/audit-cycles/new">
                  <Button className="mt-4">
                    <Plus className="mr-2 h-4 w-4" />
                    Create Audit Cycle
                  </Button>
                </Link>
              )}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Start Date</TableHead>
                  <TableHead>End Date</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredAuditCycles.map((cycle) => (
                  <TableRow key={cycle.id}>
                    <TableCell className="font-medium">
                      <Link
                        to={`/audit-cycles/${cycle.id}`}
                        className="hover:underline"
                      >
                        {cycle.name}
                      </Link>
                    </TableCell>
                    <TableCell className="max-w-md truncate">
                      {cycle.description || '-'}
                    </TableCell>
                    <TableCell>
                      {new Date(cycle.start_date).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      {new Date(cycle.end_date).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Badge variant={getStatusBadgeVariant(cycle.status)}>
                        {cycle.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {new Date(cycle.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Link to={`/audit-cycles/${cycle.id}`}>
                          <Button variant="ghost" size="icon" title="View Details">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Link to={`/audit-cycles/${cycle.id}/edit`}>
                          <Button variant="ghost" size="icon" title="Edit">
                            <Edit className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(cycle.id)}
                          title="Delete"
                        >
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
    </div>
  );
}
