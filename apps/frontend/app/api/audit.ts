import { type AxiosResponse } from "axios";
import apiClient from "./client";

// ============================================================================
// Types
// ============================================================================

export interface Framework {
  id: string;
  name: string;
  description: string;
  version: string;
  question_count: number;
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
  number: string;
  text: string;
  type: "yes_no" | "text" | "file_upload" | "multiple_choice";
  help_text?: string;
  is_mandatory: boolean;
}

export interface Audit {
  id: string;
  client_id: string;
  framework_id: string;
  framework_name: string;
  status: "draft" | "active" | "in_review" | "completed" | "archived";
  due_date?: string;
  assigned_to?: string;
  progress: AuditProgress;
  created_at: string;
  updated_at: string;
}

export interface AuditProgress {
  total_questions: number;
  answered_count: number;
  approved_count: number;
  progress_percent: number;
}

export interface Question {
  id: string;
  audit_id: string;
  section: string;
  question_number: string;
  question_text: string;
  question_type: string;
  help_text?: string;
  is_mandatory: boolean;
  display_order: number;
  submission?: QuestionSubmission;
}

export interface QuestionSubmission {
  id: string;
  status: "not_started" | "in_progress" | "submitted" | "approved" | "rejected" | "referred";
  submitted_at?: string;
}

export interface Submission {
  id: string;
  question_id: string;
  submitted_by: string;
  answer?: string;
  answer_value?: string;
  status: string;
  version: number;
  reviewed_by?: string;
  reviewed_at?: string;
  rejection_notes?: string;
  submitted_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Evidence {
  id: string;
  submission_id: string;
  file_name: string;
  file_type: string;
  file_size: number;
  storage_path: string;
  uploaded_by: string;
  description?: string;
  download_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Comment {
  id: string;
  submission_id: string;
  user_id: string;
  user_name: string;
  comment_text: string;
  is_internal: boolean;
  created_at: string;
  updated_at: string;
}

export interface ActivityLog {
  id: string;
  user_id: string;
  user_email: string;
  action: string;
  entity_type: string;
  entity_id: string;
  details?: Record<string, any>;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface Report {
  id: string;
  audit_id: string;
  unsigned_file_path?: string;
  signed_file_path?: string;
  generated_by: string;
  generated_at: string;
  signed_by?: string;
  signed_at?: string;
  status: "pending" | "generated" | "signed" | "delivered";
  metadata?: Record<string, any>;
  download_url?: string;
  created_at: string;
  updated_at: string;
}

// ============================================================================
// Framework API
// ============================================================================

export const frameworkApi = {
  // List all frameworks
  list: (): Promise<AxiosResponse<Framework[]>> =>
    apiClient.get<Framework[]>('/frameworks'),

  // Get framework by ID
  getById: (id: string): Promise<AxiosResponse<Framework>> =>
    apiClient.get<Framework>(`/frameworks/${id}`),

  // Get framework checklist
  getChecklist: (id: string): Promise<AxiosResponse<FrameworkChecklist>> =>
    apiClient.get<FrameworkChecklist>(`/frameworks/${id}/checklist`),

  // Create framework (admin only)
  create: (payload: {
    name: string;
    description: string;
    version: string;
    checklist_json: any;
  }): Promise<AxiosResponse<Framework>> =>
    apiClient.post<Framework>('/frameworks', payload),

  // Update framework (admin only)
  update: (id: string, payload: Partial<Framework>): Promise<AxiosResponse<Framework>> =>
    apiClient.put<Framework>(`/frameworks/${id}`, payload),

  // Delete framework (admin only)
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/frameworks/${id}`),
};

// ============================================================================
// Audit API
// ============================================================================

export const auditApi = {
  // List audits for a client
  list: (clientId: string): Promise<AxiosResponse<Audit[]>> =>
    apiClient.get<Audit[]>(`/clients/${clientId}/audits`),

  // Get audit with questions
  getById: (clientId: string, auditId: string): Promise<AxiosResponse<{
    audit: Audit;
    questions: Question[];
  }>> =>
    apiClient.get(`/clients/${clientId}/audits/${auditId}`),

  // Update audit
  update: (clientId: string, auditId: string, payload: {
    status?: string;
    assigned_to?: string;
    due_date?: string;
  }): Promise<AxiosResponse<Audit>> =>
    apiClient.patch<Audit>(`/clients/${clientId}/audits/${auditId}`, payload),
};

// ============================================================================
// Submission API
// ============================================================================

export const submissionApi = {
  // Create or update submission
  createOrUpdate: (clientId: string, payload: {
    question_id: string;
    answer?: string;
    answer_value?: string;
  }): Promise<AxiosResponse<Submission>> =>
    apiClient.post<Submission>(`/clients/${clientId}/submissions`, payload),

  // Submit for review
  submitForReview: (clientId: string, submissionId: string): Promise<AxiosResponse<Submission>> =>
    apiClient.post<Submission>(`/clients/${clientId}/submissions/${submissionId}/submit`),

  // Review submission
  review: (clientId: string, submissionId: string, payload: {
    action: "approve" | "reject" | "refer";
    rejection_notes?: string;
  }): Promise<AxiosResponse<Submission>> =>
    apiClient.post<Submission>(`/clients/${clientId}/submissions/${submissionId}/review`, payload),

  // List submissions by status
  list: (clientId: string, status?: string): Promise<AxiosResponse<Submission[]>> =>
    apiClient.get<Submission[]>(`/clients/${clientId}/submissions`, {
      params: { status },
    }),

  // Get submission
  getById: (clientId: string, submissionId: string): Promise<AxiosResponse<Submission>> =>
    apiClient.get<Submission>(`/clients/${clientId}/submissions/${submissionId}`),
};

// ============================================================================
// Evidence API
// ============================================================================

export const evidenceApi = {
  // Upload evidence file
  upload: (clientId: string, formData: FormData): Promise<AxiosResponse<Evidence>> =>
    apiClient.post<Evidence>(`/clients/${clientId}/evidence/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  // Get presigned upload URL
  getUploadURL: (clientId: string, params: {
    submission_id: string;
    file_name: string;
  }): Promise<AxiosResponse<{
    evidence_id: string;
    upload_url: string;
    expires_in: number;
  }>> =>
    apiClient.get(`/clients/${clientId}/evidence/upload-url`, { params }),

  // List evidence by submission
  listBySubmission: (clientId: string, submissionId: string, includeURLs = false): Promise<AxiosResponse<Evidence[]>> =>
    apiClient.get<Evidence[]>(`/clients/${clientId}/evidence/submissions/${submissionId}`, {
      params: { include_urls: includeURLs },
    }),

  // Get evidence with download URL
  getById: (clientId: string, evidenceId: string): Promise<AxiosResponse<Evidence>> =>
    apiClient.get<Evidence>(`/clients/${clientId}/evidence/${evidenceId}`),

  // Download evidence
  download: (clientId: string, evidenceId: string): Promise<AxiosResponse<Blob>> =>
    apiClient.get(`/clients/${clientId}/evidence/${evidenceId}/download`, {
      responseType: 'blob',
    }),

  // Delete evidence
  delete: (clientId: string, evidenceId: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/clients/${clientId}/evidence/${evidenceId}`),
};

// ============================================================================
// Comment API
// ============================================================================

export const commentApi = {
  // Create comment
  create: (clientId: string, payload: {
    submission_id: string;
    comment_text: string;
    is_internal: boolean;
  }): Promise<AxiosResponse<Comment>> =>
    apiClient.post<Comment>(`/clients/${clientId}/comments`, payload),

  // List comments by submission
  listBySubmission: (clientId: string, submissionId: string, filter?: "all" | "internal" | "external"): Promise<AxiosResponse<Comment[]>> =>
    apiClient.get<Comment[]>(`/clients/${clientId}/comments/submissions/${submissionId}`, {
      params: { filter },
    }),

  // Get comment
  getById: (clientId: string, commentId: string): Promise<AxiosResponse<Comment>> =>
    apiClient.get<Comment>(`/clients/${clientId}/comments/${commentId}`),

  // Update comment
  update: (clientId: string, commentId: string, payload: {
    comment_text: string;
  }): Promise<AxiosResponse<Comment>> =>
    apiClient.put<Comment>(`/clients/${clientId}/comments/${commentId}`, payload),

  // Delete comment
  delete: (clientId: string, commentId: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/clients/${clientId}/comments/${commentId}`),
};

// ============================================================================
// Activity Log API
// ============================================================================

export const activityApi = {
  // Create activity log
  create: (clientId: string, payload: {
    action: string;
    entity_type: string;
    entity_id: string;
    details?: Record<string, any>;
  }): Promise<AxiosResponse<ActivityLog>> =>
    apiClient.post<ActivityLog>(`/clients/${clientId}/activity`, payload),

  // List activity logs
  list: (clientId: string, params?: {
    limit?: number;
    offset?: number;
  }): Promise<AxiosResponse<{
    data: ActivityLog[];
    limit: number;
    offset: number;
    count: number;
  }>> =>
    apiClient.get(`/clients/${clientId}/activity`, { params }),

  // Get recent activity
  getRecent: (clientId: string, limit = 20): Promise<AxiosResponse<ActivityLog[]>> =>
    apiClient.get<ActivityLog[]>(`/clients/${clientId}/activity/recent`, {
      params: { limit },
    }),

  // List activity by user
  listByUser: (clientId: string, userId: string, params?: {
    limit?: number;
    offset?: number;
  }): Promise<AxiosResponse<{
    data: ActivityLog[];
    limit: number;
    offset: number;
    count: number;
  }>> =>
    apiClient.get(`/clients/${clientId}/activity/users/${userId}`, { params }),

  // List activity by entity
  listByEntity: (clientId: string, params: {
    entity_type: string;
    entity_id: string;
  }): Promise<AxiosResponse<ActivityLog[]>> =>
    apiClient.get<ActivityLog[]>(`/clients/${clientId}/activity/entities`, { params }),
};

// ============================================================================
// Report API
// ============================================================================

export const reportApi = {
  // Generate report
  generate: (clientId: string, auditId: string): Promise<AxiosResponse<Report>> =>
    apiClient.post<Report>(`/clients/${clientId}/reports/audits/${auditId}/generate`),

  // Get report by ID
  getById: (clientId: string, reportId: string, includeURL = false): Promise<AxiosResponse<Report>> =>
    apiClient.get<Report>(`/clients/${clientId}/reports/${reportId}`, {
      params: { include_url: includeURL },
    }),

  // Get report by audit ID
  getByAudit: (clientId: string, auditId: string, includeURL = false): Promise<AxiosResponse<Report>> =>
    apiClient.get<Report>(`/clients/${clientId}/reports/audits/${auditId}`, {
      params: { include_url: includeURL },
    }),

  // List reports by status
  list: (clientId: string, status?: "pending" | "generated" | "signed" | "delivered"): Promise<AxiosResponse<Report[]>> =>
    apiClient.get<Report[]>(`/clients/${clientId}/reports`, {
      params: { status },
    }),

  // Sign report
  sign: (clientId: string, reportId: string): Promise<AxiosResponse<Report>> =>
    apiClient.post<Report>(`/clients/${clientId}/reports/${reportId}/sign`),

  // Mark as delivered
  markDelivered: (clientId: string, reportId: string): Promise<AxiosResponse<Report>> =>
    apiClient.post<Report>(`/clients/${clientId}/reports/${reportId}/deliver`),

  // Download report
  download: (clientId: string, reportId: string, version: "signed" | "unsigned" = "signed"): Promise<AxiosResponse<Blob>> =>
    apiClient.get(`/clients/${clientId}/reports/${reportId}/download`, {
      params: { version },
      responseType: 'blob',
    }),
};

// ============================================================================
// Export combined audit API
// ============================================================================

export const auditModule = {
  frameworks: frameworkApi,
  audits: auditApi,
  submissions: submissionApi,
  evidence: evidenceApi,
  comments: commentApi,
  activity: activityApi,
  reports: reportApi,
};

export default auditModule;
