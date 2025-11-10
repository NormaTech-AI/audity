import { type AxiosResponse } from "axios";
import apiClient from "./client";
import type {
  Framework,
  FrameworkQuestion,
  CreateFrameworkPayload,
  UpdateFrameworkPayload,
} from "~/types";

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

  // Get framework questions
  getChecklist: (id: string): Promise<AxiosResponse<FrameworkQuestion[]>> =>
    apiClient.get<FrameworkQuestion[]>(`/frameworks/${id}/checklist`),

  // Create framework (admin only)
  create: (payload: CreateFrameworkPayload): Promise<AxiosResponse<Framework>> =>
    apiClient.post<Framework>('/frameworks', payload),

  // Update framework (admin only)
  update: (id: string, payload: UpdateFrameworkPayload): Promise<AxiosResponse<Framework>> =>
    apiClient.put<Framework>(`/frameworks/${id}`, payload),

  // Delete framework (admin only)
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete<void>(`/frameworks/${id}`),
};

export default frameworkApi;