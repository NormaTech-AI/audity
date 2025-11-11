import { index, layout, prefix, route, type RouteConfig } from "@react-router/dev/routes";

export default [
  // Guest routes (unauthenticated)
  layout("layouts/guest-layout.tsx", [
    route("login", "routes/login.tsx"),
    route("auth/callback", "routes/auth.callback.tsx"),
  ]),

  // Protected routes (authenticated)
  layout("layouts/protected-layout.tsx", [
    index("routes/dashboard.tsx"),
    
    // Tenant routes
    // route("tenants", "routes/tenants._index.tsx"),

    // TODO: Add more routes as pages are created
    // route("tenants/new", "routes/tenants.new.tsx"),
    route("clients", "routes/clients._index.tsx"),
    route("clients/new", "routes/clients.new.tsx"),
    route("clients/:id", "routes/clients.$id.tsx"),
    route("users", "routes/users._index.tsx"),
    route("rbac", "routes/rbac._index.tsx"),
    
    // Framework routes
    route("frameworks", "routes/frameworks._index.tsx"),
    route("frameworks/new", "routes/frameworks.new.tsx"),
    route("frameworks/:id/edit", "routes/frameworks.$id.edit.tsx"),
    route("frameworks/:id", "routes/frameworks.$id.tsx"),

    // Audit Cycle routes
    route("audit-cycles", "routes/audit-cycles._index.tsx"),
    route("audit-cycles/new", "routes/audit-cycles.new.tsx"),
    route("audit-cycles/:id/edit", "routes/audit-cycles.$id.edit.tsx"),
    route("audit-cycles/:id/add-client", "routes/audit-cycles.$id.add-client.tsx"),
    route("audit-cycles/:id/clients/:clientId/edit", "routes/audit-cycles.$id.clients.$clientId.edit.tsx"),
    route("audit-cycles/:id/clients/:clientId/assign-framework", "routes/audit-cycles.$id.clients.$clientId.assign-framework.tsx"),
    route("audit-cycles/:id/clients/:clientId", "routes/audit-cycles.$id.clients.$clientId.tsx"),
    route("audit-cycles/:id", "routes/audit-cycles.$id.tsx"),
    // route("settings", "routes/settings.tsx"),

    // Client Audit routes
    route("audit", "routes/audit.tsx"),
    route("audit/:auditId", "routes/audit.$auditId.tsx"),
]),
] satisfies RouteConfig;
