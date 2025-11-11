import { Link, useLocation } from 'react-router';
import {
  LayoutDashboard,
  Building2,
  Users,
  Shield,
  FileText,
  Settings,
  LogOut,
  ChevronLeft,
  ChevronRight,
  BookOpen,
  CircleCheckBig,
} from 'lucide-react';
import { cn } from '~/lib/utils';
import { Button } from '~/components/ui/button';
import { useAuth } from '~/contexts/AuthContext';
import { useState } from 'react';

interface NavItem {
  title: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
  permission?: string;
}

const navItems: NavItem[] = [
  {
    title: 'Dashboard',
    href: '/',
    icon: LayoutDashboard,
  },
  {
    title: 'Clients',
    href: '/clients',
    icon: Users,
    permission: 'clients:list',
  },
  {
    title: 'Users',
    href: '/users',
    icon: Users,
    permission: 'users:list',
  },
  {
    title: 'Audit',
    href: '/audit',
    icon: CircleCheckBig,
    permission: 'audit:list',
  },
  {
    title: 'Roles & Permissions',
    href: '/rbac',
    icon: Shield,
    permission: 'roles:list',
  },
  {
    title: 'Frameworks',
    href: '/frameworks',
    icon: BookOpen,
    permission: 'frameworks:list',
  },
  {
    title: 'Audit Cycles',
    href: '/audit-cycles',
    icon: CircleCheckBig,
  },
  {
    title: 'Assessments',
    href: '/assessments',
    icon: FileText,
    badge: 'Soon',
  },
  {
    title: 'Settings',
    href: '/settings',
    icon: Settings,
  },
];

export function Sidebar() {
  const location = useLocation();
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);

  return (
    <div
      className={cn(
        'flex flex-col h-screen bg-card border-r transition-all duration-300',
        collapsed ? 'w-16' : 'w-64'
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
              <Building2 className="w-4 h-4 text-primary-foreground" />
            </div>
            <span className="font-semibold">Audity</span>
          </div>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setCollapsed(!collapsed)}
          className={cn(collapsed && 'mx-auto')}
        >
          {collapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <ChevronLeft className="h-4 w-4" />
          )}
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-2 space-y-1 overflow-y-auto">
        {navItems
          .filter((item) => {
            // If user has visible_modules, filter based on that
            if(item.title == "Settings") return true
            if (user?.visible_modules && user.visible_modules.length > 0) {
              return user.visible_modules.includes(item.title);
            }
            // If no visible_modules, show all items (fallback)
            return false;
          })
          .map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.href;

            return (
              <Link
                key={item.href}
                to={item.href}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
                  collapsed && 'justify-center'
                )}
                title={collapsed ? item.title : undefined}
              >
                <Icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && (
                  <>
                    <span className="flex-1">{item.title}</span>
                    {item.badge && (
                      <span className="px-2 py-0.5 text-xs bg-muted rounded-full">
                        {item.badge}
                      </span>
                    )}
                  </>
                )}
              </Link>
            );
          })}
      </nav>

      {/* User Section */}
      <div className="p-2 border-t">
        {!collapsed && user && (
          <div className="px-3 py-2 mb-2">
            <p className="text-sm font-medium truncate">{user.name}</p>
            <p className="text-xs text-muted-foreground truncate">{user.email}</p>
            <p className="text-xs text-muted-foreground mt-1 capitalize">
              {user.designation}
            </p>
          </div>
        )}
        <Button
          variant="ghost"
          className={cn('w-full justify-start', collapsed && 'justify-center')}
          onClick={logout}
        >
          <LogOut className="h-5 w-5" />
          {!collapsed && <span className="ml-3">Logout</span>}
        </Button>
      </div>
    </div>
  );
}
