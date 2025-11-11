import { useQuery } from '@tanstack/react-query';
import { Building2, Users, Shield, TrendingUp, Calendar, CheckCircle2, Clock, AlertCircle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { api } from '~/api';
import { useAuth } from '~/contexts/AuthContext';
import type { ClientDashboardData } from '~/types';

export default function DashboardPage() {
  const { user, isAuthenticated } = useAuth();
  
  // Check if user is a client user (has client_id)
  const isClientUser = !!user?.client_id;

  // Tenant dashboard data
  const { data: dashboardData, isLoading } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const response = await api.dashboard.getTenantDashboardData();
      return response.data;
    },
    enabled: isAuthenticated && !isClientUser,
  });
  
  // Client-specific dashboard data
  const { data: clientDashboardData, isLoading: isClientLoading } = useQuery({
    queryKey: ['client-dashboard', user?.client_id],
    queryFn: async () => {
      if (!user?.client_id) return null;
      const response = await api.dashboard.getClientSpecificDashboard(user.client_id);
      return response.data;
    },
    enabled: isAuthenticated && isClientUser && !!user?.client_id,
  });
  
  // If client user, show client dashboard
  if (isClientUser) {
    return <ClientDashboard data={clientDashboardData} isLoading={isClientLoading} user={user} />;
  }

  const stats = [
    // {
    //   title: 'Total Tenants',
    //   value: dashboardData?.stats?.total_tenants || 0,
    //   icon: Building2,
    //   description: `${dashboardData?.stats?.active_tenants || 0} active`,
    //   trend: '+12%',
    // },
    {
      title: 'Total Clients',
      value: dashboardData?.stats?.total_clients || 0,
      icon: Users,
      description: `${dashboardData?.stats?.active_clients || 0} active`,
      trend: '+8%',
    },
    {
      title: 'Total Users',
      value: dashboardData?.stats?.total_users || 0,
      icon: Shield,
      description: `${dashboardData?.stats?.active_users || 0} active`,
      trend: '+5%',
    },
    {
      title: 'High Risk Clients',
      value: dashboardData?.stats?.high_risk_clients || 0,
      icon: TrendingUp,
      description: 'Requires attention',
      trend: '-3%',
      trendPositive: false,
    },
  ];

  return (
    <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground mt-2">
            Welcome back, {user?.name}! Here's what's happening with your TPRM platform.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <Card key={stat.title}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">
                    {stat.title}
                  </CardTitle>
                  <Icon className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {isLoading ? '...' : stat.value}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {stat.description}
                  </p>
                  {stat.trend && (
                    <p
                      className={`text-xs mt-1 ${
                        stat.trendPositive === false
                          ? 'text-red-600'
                          : 'text-green-600'
                      }`}
                    >
                      {stat.trend} from last month
                    </p>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Recent Activity */}
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Recent Tenants</CardTitle>
              <CardDescription>Latest tenant registrations</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <p className="text-sm text-muted-foreground">Loading...</p>
              ) : dashboardData?.recent_tenants?.length ? (
                <div className="space-y-4">
                  {dashboardData.recent_tenants.map((tenant) => (
                    <div
                      key={tenant.id}
                      className="flex items-center justify-between"
                    >
                      <div>
                        <p className="text-sm font-medium">{tenant.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {tenant.subdomain}
                        </p>
                      </div>
                      <span
                        className={`text-xs px-2 py-1 rounded-full ${
                          tenant.status === 'active'
                            ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                            : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
                        }`}
                      >
                        {tenant.status}
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No recent tenants</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recent Clients</CardTitle>
              <CardDescription>Latest client onboardings</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <p className="text-sm text-muted-foreground">Loading...</p>
              ) : dashboardData?.recent_clients?.length ? (
                <div className="space-y-4">
                  {dashboardData.recent_clients.map((client) => (
                    <div
                      key={client.id}
                      className="flex items-center justify-between"
                    >
                      <div>
                        <p className="text-sm font-medium">{client.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {client.industry || 'No industry'}
                        </p>
                      </div>
                      <span
                        className={`text-xs px-2 py-1 rounded-full ${
                          client.risk_tier === 'high' || client.risk_tier === 'critical'
                            ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                            : client.risk_tier === 'medium'
                            ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
                            : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                        }`}
                      >
                        {client.risk_tier || 'low'}
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No recent clients</p>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common tasks and shortcuts</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-2 md:grid-cols-3">
              <button className="p-4 text-left border rounded-lg hover:bg-accent transition-colors">
                <Building2 className="h-5 w-5 mb-2 text-primary" />
                <p className="font-medium text-sm">Create Tenant</p>
                <p className="text-xs text-muted-foreground">Add new organization</p>
              </button>
              <button className="p-4 text-left border rounded-lg hover:bg-accent transition-colors">
                <Users className="h-5 w-5 mb-2 text-primary" />
                <p className="font-medium text-sm">Add Client</p>
                <p className="text-xs text-muted-foreground">Onboard third-party</p>
              </button>
              <button className="p-4 text-left border rounded-lg hover:bg-accent transition-colors">
                <Shield className="h-5 w-5 mb-2 text-primary" />
                <p className="font-medium text-sm">Manage Roles</p>
                <p className="text-xs text-muted-foreground">Configure permissions</p>
              </button>
            </div>
          </CardContent>
        </Card>
      </div>
  );
}

// Client Dashboard Component
function ClientDashboard({ data, isLoading, user }: { data: ClientDashboardData | null | undefined, isLoading: boolean, user: any }) {
  const companyName = data?.client_name || 'Company';
  
  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">{companyName} Dashboard</h1>
        <p className="text-muted-foreground mt-2">
          Welcome back, {user?.name}! Track your audit cycles and compliance progress.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Audit Cycles</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {isLoading ? '...' : data?.stats?.active_audit_cycles || 0}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Currently enrolled cycles
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Framework Assignments</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {isLoading ? '...' : data?.stats?.total_framework_assignments || 0}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Total frameworks to complete
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Audit Cycles */}
      <Card>
        <CardHeader>
          <CardTitle>Audit Cycle Enrollments</CardTitle>
          <CardDescription>Your current and upcoming audit cycles</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Loading...</p>
          ) : data?.audit_cycles?.length ? (
            <div className="space-y-6">
              {data.audit_cycles.map((cycle) => (
                <div key={cycle.audit_cycle_id} className="border rounded-lg p-4">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h3 className="font-semibold">{cycle.audit_cycle_name}</h3>
                      {cycle.audit_cycle_description && (
                        <p className="text-sm text-muted-foreground mt-1">
                          {cycle.audit_cycle_description}
                        </p>
                      )}
                      <div className="flex gap-4 mt-2 text-xs text-muted-foreground">
                        <span>Start: {new Date(cycle.start_date).toLocaleDateString()}</span>
                        <span>End: {new Date(cycle.end_date).toLocaleDateString()}</span>
                      </div>
                    </div>
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      cycle.cycle_status === 'active'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                        : 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
                    }`}>
                      {cycle.cycle_status}
                    </span>
                  </div>
                  
                  {/* Frameworks in this cycle */}
                  {cycle.frameworks?.length > 0 && (
                    <div className="mt-4 space-y-2">
                      <p className="text-sm font-medium">Assigned Frameworks:</p>
                      <div className="grid gap-2">
                        {cycle.frameworks.map((framework, idx) => (
                          <div key={idx} className="flex items-center justify-between p-2 bg-muted rounded">
                            <div className="flex items-center gap-2">
                              <Shield className="h-4 w-4" />
                              <span className="text-sm">{framework.framework_name}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              {framework.due_date && (
                                <span className="text-xs text-muted-foreground">
                                  Due: {new Date(framework.due_date).toLocaleDateString()}
                                </span>
                              )}
                              <span className={`text-xs px-2 py-1 rounded-full ${
                                framework.framework_status === 'completed'
                                  ? 'bg-green-100 text-green-800'
                                  : framework.framework_status === 'in_progress'
                                  ? 'bg-blue-100 text-blue-800'
                                  : framework.framework_status === 'overdue'
                                  ? 'bg-red-100 text-red-800'
                                  : 'bg-gray-100 text-gray-800'
                              }`}>
                                {framework.framework_status}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No audit cycles assigned</p>
          )}
        </CardContent>
      </Card>

      {/* Framework Analytics */}
      <Card>
        <CardHeader>
          <CardTitle>Framework Progress Analytics</CardTitle>
          <CardDescription>Questions answered vs total questions for each framework</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Loading...</p>
          ) : data?.framework_analytics?.length ? (
            <div className="space-y-4">
              {data.framework_analytics.map((analytics) => {
                const progress = analytics.total_questions > 0 
                  ? (analytics.answered_questions / analytics.total_questions) * 100 
                  : 0;
                
                return (
                  <div key={analytics.audit_id} className="border rounded-lg p-4">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <h3 className="font-semibold">{analytics.framework_name}</h3>
                        <div className="flex gap-4 mt-1 text-xs text-muted-foreground">
                          <span>Due: {new Date(analytics.due_date).toLocaleDateString()}</span>
                          <span className={`px-2 py-0.5 rounded-full ${
                            analytics.status === 'completed'
                              ? 'bg-green-100 text-green-800'
                              : analytics.status === 'in_progress'
                              ? 'bg-blue-100 text-blue-800'
                              : analytics.status === 'overdue'
                              ? 'bg-red-100 text-red-800'
                              : 'bg-gray-100 text-gray-800'
                          }`}>
                            {analytics.status}
                          </span>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-2xl font-bold">{Math.round(progress)}%</div>
                        <p className="text-xs text-muted-foreground">Complete</p>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm">
                        <span>Questions Answered</span>
                        <span className="font-medium">
                          {analytics.answered_questions} / {analytics.total_questions}
                        </span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                          className="bg-primary h-2 rounded-full transition-all"
                          style={{ width: `${progress}%` }}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No framework analytics available</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
