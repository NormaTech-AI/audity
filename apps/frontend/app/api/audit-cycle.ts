import { type AxiosResponse } from "axios";
import apiClient from "./client";
import type {
  AuditCycle,
  AuditCycleClient,
  AuditCycleFramework,
  AuditCycleStats,
  CreateAuditCyclePayload,
  UpdateAuditCyclePayload,
  AssignFrameworkPayload,
} from "~/types";

// ============================================================================
// Audit Cycle API
// ============================================================================

export const auditCycleApi = {
  // List all audit cycles
  list: (): Promise<AxiosResponse<AuditCycle[]>> =>
    apiClient.get<AuditCycle[]>('/audit-cycles'),

  // Get audit cycle by ID
  getById: (id: string): Promise<AxiosResponse<AuditCycle>> =>
    apiClient.get<AuditCycle>(`/audit-cycles/${id}`),

  // Get audit cycle statistics
  getStats: (id: string): Promise<AxiosResponse<AuditCycleStats>> =>
    apiClient.get<AuditCycleStats>(`/audit-cycles/${id}/stats`),

  // Create audit cycle
  create: (payload: CreateAuditCyclePayload): Promise<AxiosResponse<AuditCycle>> =>
    apiClient.post<AuditCycle>('/audit-cycles', payload),

  // Update audit cycle
  update: (id: string, payload: UpdateAuditCyclePayload): Promise<AxiosResponse<AuditCycle>> =>
    apiClient.put<AuditCycle>(`/audit-cycles/${id}`, payload),

  // Delete audit cycle
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/audit-cycles/${id}`),

  // Get clients in audit cycle
  getClients: (id: string): Promise<AxiosResponse<AuditCycleClient[]>> =>
    apiClient.get<AuditCycleClient[]>(`/audit-cycles/${id}/clients`),

  // Add client to audit cycle
  addClient: (id: string, clientId: string): Promise<AxiosResponse<AuditCycleClient>> =>
    apiClient.post<AuditCycleClient>(`/audit-cycles/${id}/clients`, { client_id: clientId }),

  // Remove client from audit cycle
  removeClient: (id: string, clientId: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/audit-cycles/${id}/clients/${clientId}`),

  // Get frameworks in audit cycle
  getFrameworks: (id: string): Promise<AxiosResponse<AuditCycleFramework[]>> =>
    apiClient.get<AuditCycleFramework[]>(`/audit-cycles/${id}/frameworks`),

  // Assign framework to client in audit cycle
  assignFramework: (cycleClientId: string, payload: AssignFrameworkPayload): Promise<AxiosResponse<AuditCycleFramework>> =>
    apiClient.post<AuditCycleFramework>(`/audit-cycles/clients/${cycleClientId}/frameworks`, payload),
};

export default auditCycleApi;
