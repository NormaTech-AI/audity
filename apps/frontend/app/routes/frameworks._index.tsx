import { useState, useEffect } from 'react';
import { Link } from 'react-router';
import { Plus, FileText, Search, Trash2, Edit, Eye } from 'lucide-react';
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
import type { Framework } from '~/types';

export default function FrameworksPage() {
  const [frameworks, setFrameworks] = useState<Framework[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadFrameworks();
  }, []);

  const loadFrameworks = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.frameworks.list();
      setFrameworks(response.data);
    } catch (err: any) {
      console.error('Failed to load frameworks:', err);
      setError(err.response?.data?.error || 'Failed to load frameworks');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this framework?')) {
      return;
    }

    try {
      await api.frameworks.delete(id);
      setFrameworks(frameworks.filter((f) => f.id !== id));
    } catch (err: any) {
      console.error('Failed to delete framework:', err);
      setError(err.response?.data?.error || 'Failed to delete framework');
    }
  };

  const filteredFrameworks = frameworks.filter((framework) =>
    framework.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    framework.description.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Compliance Frameworks</h1>
          <p className="text-muted-foreground mt-2">
            Manage compliance frameworks and their checklists
          </p>
        </div>
        <Link to="/frameworks/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Add Framework
          </Button>
        </Link>
      </div>

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search frameworks..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </CardContent>
      </Card>

      {/* Error Message */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-destructive">{error}</p>
          </CardContent>
        </Card>
      )}

      {/* Frameworks Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Frameworks</CardTitle>
          <CardDescription>
            {filteredFrameworks.length} framework{filteredFrameworks.length !== 1 ? 's' : ''} found
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : filteredFrameworks.length === 0 ? (
            <div className="text-center py-8">
              <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No frameworks found</h3>
              <p className="text-muted-foreground mt-2">
                {searchQuery
                  ? 'Try adjusting your search query'
                  : 'Get started by creating your first framework'}
              </p>
              {!searchQuery && (
                <Link to="/frameworks/new">
                  <Button className="mt-4">
                    <Plus className="mr-2 h-4 w-4" />
                    Add Framework
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
                  <TableHead>Version</TableHead>
                  <TableHead>Questions</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredFrameworks.map((framework) => (
                  <TableRow key={framework.id}>
                    <TableCell className="font-medium">
                      <Link
                        to={`/frameworks/${framework.id}`}
                        className="hover:underline"
                      >
                        {framework.name}
                      </Link>
                    </TableCell>
                    <TableCell className="max-w-md truncate">
                      {framework.description}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{framework.version}</Badge>
                    </TableCell>
                    <TableCell>
                      {framework.question_count || 0}
                    </TableCell>
                    <TableCell>
                      {new Date(framework.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Link to={`/frameworks/${framework.id}`}>
                          <Button variant="ghost" size="icon">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Link to={`/frameworks/${framework.id}/edit`}>
                          <Button variant="ghost" size="icon">
                            <Edit className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(framework.id)}
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
