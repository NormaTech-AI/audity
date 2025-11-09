import { type AxiosResponse } from "axios";
import apiClient from "./client";
import type {
  User,
  LoginCredentials,
  AuthResponse,
  OAuthLoginResponse,
  Tenant,
  CreateTenantPayload,
  UpdateTenantPayload,
  Client,
  CreateClientPayload,
  UpdateClientPayload,
  Role,
  Permission,
  AssignRolePayload,
  DashboardData,
  PaginatedResponse,
  QueryParams,
} from "~/types";

// ============================================================================
// Authentication API
// ============================================================================

export const authApi = {
  // OAuth login - Get auth URL from backend
  initiateGoogleLogin: (): Promise<AxiosResponse<OAuthLoginResponse>> => 
    apiClient.get<OAuthLoginResponse>('/auth/login/google'),
  
  initiateMicrosoftLogin: (): Promise<AxiosResponse<OAuthLoginResponse>> =>
    apiClient.get<OAuthLoginResponse>('/auth/login/microsoft'),

  // Set token as cookie (for OAuth callback)
  setTokenCookie: (token: string): Promise<AxiosResponse<User>> =>
    apiClient.post<User>(`/auth/set-token?token=${token}`),

  // Validate current token
  validateToken: (): Promise<AxiosResponse<User>> =>
    apiClient.get<User>('/auth/validate'),

  // Refresh token
  refreshToken: (): Promise<AxiosResponse<AuthResponse>> =>
    apiClient.post<AuthResponse>('/auth/refresh'),

  // Logout
  logout: (): Promise<AxiosResponse<void>> =>
    apiClient.post<void>('/auth/logout'),

  // Get current user
  getMe: (): Promise<AxiosResponse<User>> =>
    apiClient.get<User>('/auth/validate'),
};

// ============================================================================
// Tenant API
// ============================================================================

export const tenantApi = {
  // List all tenants
  list: (params?: QueryParams): Promise<AxiosResponse<PaginatedResponse<Tenant>>> =>
    apiClient.get<PaginatedResponse<Tenant>>('/tenants', { params }),

  // Get tenant by ID
  getById: (id: string): Promise<AxiosResponse<Tenant>> =>
    apiClient.get<Tenant>(`/tenants/${id}`),

  // Create new tenant
  create: (payload: CreateTenantPayload): Promise<AxiosResponse<Tenant>> =>
    apiClient.post<Tenant>('/tenants', payload),

  // Update tenant
  update: (id: string, payload: UpdateTenantPayload): Promise<AxiosResponse<Tenant>> =>
    apiClient.put<Tenant>(`/tenants/${id}`, payload),

  // Delete tenant
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/tenants/${id}`),

  // Activate tenant
  activate: (id: string): Promise<AxiosResponse<Tenant>> =>
    apiClient.post<Tenant>(`/tenants/${id}/activate`),

  // Deactivate tenant
  deactivate: (id: string): Promise<AxiosResponse<Tenant>> =>
    apiClient.post<Tenant>(`/tenants/${id}/deactivate`),
};

// ============================================================================
// Client API
// ============================================================================

export const clientApi = {
  // List all clients
  list: (params?: QueryParams): Promise<AxiosResponse<PaginatedResponse<Client>>> =>
    apiClient.get<PaginatedResponse<Client>>('/clients', { params }),

  // Get client by ID
  getById: (id: string): Promise<AxiosResponse<Client>> =>
    apiClient.get<Client>(`/clients/${id}`),

  // Create new client
  create: (payload: CreateClientPayload): Promise<AxiosResponse<Client>> =>
    apiClient.post<Client>('/clients', payload),

  // Update client
  update: (id: string, payload: UpdateClientPayload): Promise<AxiosResponse<Client>> =>
    apiClient.put<Client>(`/clients/${id}`, payload),

  // Delete client
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/clients/${id}`),

  // Get client statistics
  getStats: (id: string): Promise<AxiosResponse<any>> =>
    apiClient.get(`/clients/${id}/stats`),
};

// ============================================================================
// User API
// ============================================================================

export const userApi = {
  // List all users
  list: (params?: QueryParams): Promise<AxiosResponse<PaginatedResponse<User>>> =>
    apiClient.get<PaginatedResponse<User>>('/users', { params }),

  // Get user by ID
  getById: (id: string): Promise<AxiosResponse<User>> =>
    apiClient.get<User>(`/users/${id}`),

  // Create new user
  create: (payload: Partial<User>): Promise<AxiosResponse<User>> =>
    apiClient.post<User>('/users', payload),

  // Update user
  update: (id: string, payload: Partial<User>): Promise<AxiosResponse<User>> =>
    apiClient.put<User>(`/users/${id}`, payload),

  // Delete user
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/users/${id}`),

  // Get user's roles
  getRoles: (id: string): Promise<AxiosResponse<Role[]>> =>
    apiClient.get<Role[]>(`/users/${id}/roles`),

  // Assign role to user
  assignRole: (payload: AssignRolePayload): Promise<AxiosResponse<void>> =>
    apiClient.post<void>('/users/roles', payload),

  // Remove role from user
  removeRole: (userId: string, roleId: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/users/${userId}/roles/${roleId}`),
};

// ============================================================================
// Role & Permission API
// ============================================================================

export const rbacApi = {
  // List all roles
  listRoles: (): Promise<AxiosResponse<Role[]>> =>
    apiClient.get<Role[]>('/roles'),

  // Get role by ID
  getRole: (id: string): Promise<AxiosResponse<Role>> =>
    apiClient.get<Role>(`/roles/${id}`),

  // Create new role
  createRole: (payload: Partial<Role>): Promise<AxiosResponse<Role>> =>
    apiClient.post<Role>('/roles', payload),

  // Update role
  updateRole: (id: string, payload: Partial<Role>): Promise<AxiosResponse<Role>> =>
    apiClient.put<Role>(`/roles/${id}`, payload),

  // Delete role
  deleteRole: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/roles/${id}`),

  // List all permissions
  listPermissions: (): Promise<AxiosResponse<Permission[]>> =>
    apiClient.get<Permission[]>('/permissions'),

  // Assign permission to role
  assignPermission: (roleId: string, permissionId: string): Promise<AxiosResponse<void>> =>
    apiClient.post<void>(`/roles/${roleId}/permissions/${permissionId}`),

  // Remove permission from role
  removePermission: (roleId: string, permissionId: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/roles/${roleId}/permissions/${permissionId}`),
};

// ============================================================================
// Dashboard API
// ============================================================================

export const dashboardApi = {
  // Get dashboard data
  getTenantDashboardData: (): Promise<AxiosResponse<DashboardData>> =>
    apiClient.get<DashboardData>('/tenant/dashboard'),

  // Get statistics
  getTenantDashboardStats: (): Promise<AxiosResponse<any>> =>
    apiClient.get('/tenant/dashboard/stats'),

  getClientDashboardData: (): Promise<AxiosResponse<DashboardData>> =>
    apiClient.get<DashboardData>('/client/dashboard'),

  getClientDashboardStats: (): Promise<AxiosResponse<any>> =>
    apiClient.get('/client/dashboard/stats'),
};

// ============================================================================
// Audit Module (Import all audit-related APIs)
// ============================================================================

import { auditModule } from "./audit";

// ============================================================================
// Export all APIs
// ============================================================================

export const api = {
  auth: authApi,
  tenants: tenantApi,
  clients: clientApi,
  users: userApi,
  rbac: rbacApi,
  dashboard: dashboardApi,
  // Audit module
  frameworks: auditModule.frameworks,
  audits: auditModule.audits,
  submissions: auditModule.submissions,
  evidence: auditModule.evidence,
  comments: auditModule.comments,
  activity: auditModule.activity,
  reports: auditModule.reports,
};

export default api;

// Re-export audit types for convenience
export type * from "./audit";
