import axios, { type AxiosResponse, type AxiosError } from "axios";

// Create axios instance with base configuration
export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api',
  withCredentials: true, // Use HTTP-only cookies for auth
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
    Accept: "application/json",
  },
});

// Request interceptor to add Authorization header
apiClient.interceptors.request.use(
  (config) => {
    try {
      if (typeof window !== 'undefined') {
        const token = localStorage.getItem('auth_token');
        if (token) {
          config.headers = config.headers ?? {};
          if (!config.headers['Authorization']) {
            (config.headers as any)['Authorization'] = `Bearer ${token}`;
          }
        }
      }
    } catch {}
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for token refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as any;

    // If 401 and not already retrying
    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      originalRequest.url !== '/auth/refresh'
    ) {
      originalRequest._retry = true;

      try {
        // Try to refresh the session
        await apiClient.post('/auth/refresh');
        
        // Retry original request
        return apiClient(originalRequest);
      } catch (refreshError) {
        if (typeof window !== 'undefined' && !window.location.href.includes("/login")) {
          window.location.href = '/login';
        }
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;
