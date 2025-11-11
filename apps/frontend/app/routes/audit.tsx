import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router';
import { Calendar, CheckCircle2, Clock, AlertCircle, FileText } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { Badge } from '~/components/ui/badge';
import { api } from '~/api';
import { useAuth } from '~/contexts/AuthContext';
import type { ClientAudit } from '~/types';

export default function ClientAuditListPage() {
  const { user, isAuthenticated } = useAuth();

  const { data: audits, isLoading } = useQuery({
    queryKey: ['client-audits'],
    queryFn: async () => {
      const response = await api.clientAudit.listAudits();
      return response.data;
    },
    enabled: isAuthenticated,
  });

  const getStatusBadge = (status: string) => {
    const statusConfig: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      not_started: { label: 'Not Started', variant: 'outline' },
      in_progress: { label: 'In Progress', variant: 'default' },
      under_review: { label: 'Under Review', variant: 'secondary' },
      completed: { label: 'Completed', variant: 'secondary' },
      overdue: { label: 'Overdue', variant: 'destructive' },
    };

    const config = statusConfig[status] || { label: status, variant: 'outline' };
    return <Badge variant={config.variant}>{config.label}</Badge>;
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle2 className="h-5 w-5 text-green-600" />;
      case 'in_progress':
        return <Clock className="h-5 w-5 text-blue-600" />;
      case 'overdue':
        return <AlertCircle className="h-5 w-5 text-red-600" />;
      default:
        return <FileText className="h-5 w-5 text-gray-600" />;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Audit Frameworks</h1>
        <p className="text-muted-foreground mt-2">
          View and complete your assigned compliance frameworks
        </p>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading audits...</p>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && (!audits || audits.length === 0) && (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">No Audits Assigned</h3>
              <p className="text-muted-foreground">
                You don't have any compliance frameworks assigned yet.
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Audits Grid */}
      {!isLoading && audits && audits.length > 0 && (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {audits.map((audit: ClientAudit) => (
            <Link key={audit.id} to={`/audit/${audit.id}`} className="block">
              <Card className="hover:shadow-lg transition-shadow h-full">
                <CardHeader>
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(audit.status)}
                      <CardTitle className="text-lg">{audit.framework_name}</CardTitle>
                    </div>
                    {getStatusBadge(audit.status)}
                  </div>
                  <CardDescription className="flex items-center gap-2">
                    <Calendar className="h-4 w-4" />
                    Due: {new Date(audit.due_date).toLocaleDateString()}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {/* Progress Bar */}
                    <div>
                      <div className="flex justify-between text-sm mb-2">
                        <span className="text-muted-foreground">Progress</span>
                        <span className="font-medium">{Math.round(audit.progress_percent)}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div
                          className={`h-2 rounded-full transition-all ${
                            audit.progress_percent === 100
                              ? 'bg-green-600'
                              : audit.progress_percent > 50
                              ? 'bg-blue-600'
                              : 'bg-yellow-600'
                          }`}
                          style={{ width: `${audit.progress_percent}%` }}
                        />
                      </div>
                    </div>

                    {/* Stats */}
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <p className="text-muted-foreground">Total Questions</p>
                        <p className="font-semibold text-lg">{audit.total_questions}</p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Answered</p>
                        <p className="font-semibold text-lg">{audit.answered_count}</p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
