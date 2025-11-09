import { useQuery } from '@tanstack/react-query';
import { Building2, Users, Shield, TrendingUp } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { api } from '~/api';
import { useAuth } from '~/contexts/AuthContext';

export default function DashboardPage() {
  const { user, isAuthenticated } = useAuth();

  const { data: dashboardData, isLoading } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const response = await api.dashboard.getTenantDashboardData();
      return response.data;
    },
    enabled: isAuthenticated, // Only fetch when user is authenticated
  });

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
