import { type AxiosResponse } from "axios";
import apiClient from "./client";
import type {
  ClientAudit,
  ClientAuditDetail,
  ClientSubmissionPayload,
} from "~/types";

export const clientAuditApi = {
  // List all audits for authenticated client user
  listAudits: (): Promise<AxiosResponse<ClientAudit[]>> =>
    apiClient.get<ClientAudit[]>('/client-audit'),

  // Get audit detail with questions (role-based filtering)
  getAuditDetail: (auditId: string): Promise<AxiosResponse<ClientAuditDetail>> =>
    apiClient.get<ClientAuditDetail>(`/client-audit/${auditId}`),

  // Save submission (create or update draft)
  saveSubmission: (payload: ClientSubmissionPayload): Promise<AxiosResponse<any>> =>
    apiClient.post('/client-audit/submissions', payload),

  // Submit answer for review
  submitAnswer: (submissionId: string): Promise<AxiosResponse<any>> =>
    apiClient.post(`/client-audit/submissions/${submissionId}/submit`),
};
