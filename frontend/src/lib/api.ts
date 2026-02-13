import type { Agent, Workflow, Execution, ApiResponse } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function fetchApi<T>(
  path: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: "Unknown error" }));
      return { error: error.error || response.statusText };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : "Network error" };
  }
}

// Agent API
export const agentApi = {
  list: () => fetchApi<Agent[]>("/api/v1/agents"),

  get: (id: string) => fetchApi<Agent>(`/api/v1/agents/${id}`),

  create: (data: Partial<Agent>) =>
    fetchApi<Agent>("/api/v1/agents", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: string, data: Partial<Agent>) =>
    fetchApi<Agent>(`/api/v1/agents/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchApi<void>(`/api/v1/agents/${id}`, { method: "DELETE" }),
};

// Workflow API
export const workflowApi = {
  list: () => fetchApi<Workflow[]>("/api/v1/workflows"),

  get: (id: string) => fetchApi<Workflow>(`/api/v1/workflows/${id}`),

  create: (data: Partial<Workflow>) =>
    fetchApi<Workflow>("/api/v1/workflows", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: string, data: Partial<Workflow>) =>
    fetchApi<Workflow>(`/api/v1/workflows/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: string) =>
    fetchApi<void>(`/api/v1/workflows/${id}`, { method: "DELETE" }),

  execute: (id: string, inputData?: Record<string, unknown>) =>
    fetchApi<{ execution_id: string; status: string }>(
      `/api/v1/workflows/${id}/execute`,
      {
        method: "POST",
        body: JSON.stringify({ input_data: inputData }),
      }
    ),
};

// Execution API
export const executionApi = {
  list: (workflowId?: string, status?: string) => {
    const params = new URLSearchParams();
    if (workflowId) params.append("workflow_id", workflowId);
    if (status) params.append("status", status);
    const query = params.toString();
    return fetchApi<Execution[]>(`/api/v1/executions${query ? `?${query}` : ""}`);
  },

  get: (id: string) => fetchApi<Execution>(`/api/v1/executions/${id}`),

  replay: (id: string, modifiedSteps?: Array<{ step_id: string; new_output: string }>) =>
    fetchApi<{
      original_execution_id: string;
      new_execution_id: string;
      status: string;
    }>(`/api/v1/executions/${id}/replay`, {
      method: "POST",
      body: JSON.stringify({ modified_steps: modifiedSteps }),
    }),
};

// Health check
export const healthCheck = () => fetchApi<{ status: string }>("/health");
