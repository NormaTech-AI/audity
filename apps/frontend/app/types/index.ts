// ============================================================================
// Type Definitions for Audity TPRM Platform
// ============================================================================

// ============================================================================
// Authentication & User Types
// ============================================================================

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  designation: string;
  roles?: string[];
  client_id?: string;
  visible_modules?: string[];
  created_at: string;
  updated_at: string;
  last_login?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}

export interface OAuthProvider {
  name: string;
  display_name: string;
  icon: string;
  url: string;
}

export interface OAuthLoginResponse {
  auth_url: string;
  provider: string;
}

// ============================================================================
// Tenant Types
// ============================================================================

export interface Tenant {
  id: string;
  name: string;
  subdomain: string;
  database_name: string;
  status: TenantStatus;
  created_at: string;
  updated_at: string;
  settings?: TenantSettings;
}

export type TenantStatus = 'active' | 'inactive' | 'suspended' | 'pending';

export interface TenantSettings {
  max_users?: number;
  max_clients?: number;
  features?: string[];
  custom_branding?: boolean;
}

export interface CreateTenantPayload {
  name: string;
  subdomain: string;
  admin_email: string;
  admin_name: string;
}

export interface UpdateTenantPayload {
  name?: string;
  status?: TenantStatus;
  settings?: TenantSettings;
}

// ============================================================================
// Client Types
// ============================================================================

export interface Client {
  id: string;
  name: string;
  poc_email: string;
  email_domain?: string;
  status: ClientStatus;
  industry?: string;
  risk_tier?: RiskTier;
  contact_phone?: string;
  address?: string;
  created_at: string;
  updated_at: string;
  assigned_users?: string[];
}

export type ClientStatus = 'active' | 'inactive' | 'suspended' | 'onboarding' | 'offboarding';
export type RiskTier = 'low' | 'medium' | 'high' | 'critical';

export interface CreateClientPayload {
  name: string;
  poc_email: string;
  email_domain?: string;
  industry?: string;
  risk_tier?: RiskTier;
  contact_phone?: string;
  address?: string;
}

export interface UpdateClientPayload {
  name?: string;
  poc_email?: string;
  email_domain?: string;
  status?: ClientStatus;
  industry?: string;
  risk_tier?: RiskTier;
  contact_phone?: string;
  address?: string;
}

// ============================================================================
// RBAC Types
// ============================================================================

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: Permission[];
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  name: string;
  description?: string;
  resource: string;
  action: string;
  created_at: string;
}

export interface UserRole {
  user_id: string;
  role_id: string;
  assigned_at: string;
  assigned_by?: string;
}

export interface AssignRolePayload {
  user_id: string;
  role_id: string;
}

// ============================================================================
// Dashboard Types
// ============================================================================

export interface DashboardStats {
  total_tenants: number;
  active_tenants: number;
  total_clients: number;
  active_clients: number;
  total_users: number;
  active_users: number;
  high_risk_clients: number;
}

export interface DashboardData {
  stats: DashboardStats;
  recent_tenants?: Tenant[];
  recent_clients?: Client[];
  recent_activity?: ActivityLog[];
}

export interface ActivityLog {
  id: string;
  user_id: string;
  user_name: string;
  action: string;
  resource_type: string;
  resource_id: string;
  details?: string;
  timestamp: string;
}

// ============================================================================
// Framework Types
// ============================================================================

export interface Framework {
  id: string;
  name: string;
  description: string;
  version: string;
  question_count?: number;
  created_at: string;
  updated_at: string;
}

export interface FrameworkChecklist {
  sections: FrameworkSection[];
}

export interface FrameworkSection {
  title: string;
  description?: string;
  questions: FrameworkQuestion[];
}

export interface FrameworkQuestion {
  question_id?: string;
  section_title: string;
  control_id: string;
  question_text: string;
  help_text?: string;
  acceptable_evidence: string[];
}

export interface CreateFrameworkPayload {
  name: string;
  description: string;
  version: string;
  questions: FrameworkQuestion[];
}

export interface UpdateFrameworkPayload {
  name?: string;
  description?: string;
  version?: string;
  questions?: FrameworkQuestion[];
}

// ============================================================================
// Assessment Types (Future Scope)
// ============================================================================

export interface Assessment {
  id: string;
  client_id: string;
  name: string;
  type: AssessmentType;
  status: AssessmentStatus;
  due_date?: string;
  completed_date?: string;
  score?: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export type AssessmentType = 'initial' | 'annual' | 'ad_hoc' | 'continuous';
export type AssessmentStatus = 'draft' | 'in_progress' | 'completed' | 'approved' | 'rejected';

export interface Question {
  id: string;
  assessment_id: string;
  text: string;
  type: QuestionType;
  required: boolean;
  options?: string[];
  answer?: string;
  score?: number;
}

export type QuestionType = 'text' | 'multiple_choice' | 'yes_no' | 'rating' | 'file_upload';

// ============================================================================
// Document Types (Future Scope)
// ============================================================================

export interface Document {
  id: string;
  client_id: string;
  name: string;
  type: DocumentType;
  file_url: string;
  file_size: number;
  uploaded_by: string;
  uploaded_at: string;
  expires_at?: string;
  status: DocumentStatus;
}

export type DocumentType = 'contract' | 'certificate' | 'policy' | 'report' | 'other';
export type DocumentStatus = 'active' | 'expired' | 'pending_review' | 'archived';

// ============================================================================
// Risk Types (Future Scope)
// ============================================================================

export interface RiskAssessment {
  id: string;
  client_id: string;
  overall_score: number;
  risk_level: RiskTier;
  categories: RiskCategory[];
  assessed_at: string;
  assessed_by: string;
  next_review_date?: string;
}

export interface RiskCategory {
  name: string;
  score: number;
  weight: number;
  findings?: string[];
}

// ============================================================================
// Notification Types
// ============================================================================

export interface Notification {
  id: string;
  user_id: string;
  title: string;
  message: string;
  type: NotificationType;
  read: boolean;
  created_at: string;
  action_url?: string;
}

export type NotificationType = 'info' | 'warning' | 'error' | 'success';

// ============================================================================
// API Response Types
// ============================================================================

export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ApiError {
  error: string;
  message?: string;
  details?: Record<string, string[]>;
}

// ============================================================================
// Form Types
// ============================================================================

export interface FormField {
  name: string;
  label: string;
  type: string;
  required?: boolean;
  placeholder?: string;
  options?: { label: string; value: string }[];
  validation?: Record<string, any>;
}

// ============================================================================
// Table Types
// ============================================================================

export interface TableColumn<T> {
  key: keyof T | string;
  label: string;
  sortable?: boolean;
  render?: (value: any, row: T) => React.ReactNode;
}

export interface TableProps<T> {
  data: T[];
  columns: TableColumn<T>[];
  onRowClick?: (row: T) => void;
  loading?: boolean;
  emptyMessage?: string;
}

// ============================================================================
// Filter & Sort Types
// ============================================================================

export interface FilterOption {
  field: string;
  operator: 'eq' | 'ne' | 'gt' | 'lt' | 'gte' | 'lte' | 'contains' | 'in';
  value: any;
}

export interface SortOption {
  field: string;
  direction: 'asc' | 'desc';
}

export interface QueryParams {
  page?: number;
  page_size?: number;
  filters?: FilterOption[];
  sort?: SortOption;
  search?: string;
}
