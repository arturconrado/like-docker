import type {
  CreateWorkloadPayload,
  DashboardSummaryResponse,
  DemoDefinition,
  DemoRunResponse,
  DemoValidation,
  EventItem,
  HealthResponse,
  HostCapabilities,
  Workload,
} from '../types'

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL?.toString().trim() || 'http://localhost:8080'

class APIError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'APIError'
    this.status = status
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers ?? {}),
    },
    ...options,
  })

  if (!response.ok) {
    let message = `Falha na requisição (${response.status})`
    try {
      const payload = (await response.json()) as { error?: string }
      if (payload.error) {
        message = payload.error
      }
    } catch {
      // noop
    }
    throw new APIError(message, response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

export function getHealth() {
  return request<HealthResponse>('/health')
}

export function listWorkloads() {
  return request<Workload[]>('/api/workloads')
}

export function getWorkload(id: string) {
  return request<Workload>(`/api/workloads/${id}`)
}

export function createWorkload(payload: CreateWorkloadPayload) {
  return request<Workload>('/api/workloads', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function stopWorkload(id: string) {
  return request<Workload>(`/api/workloads/${id}/stop`, {
    method: 'POST',
  })
}

export function deleteWorkload(id: string) {
  return request<void>(`/api/workloads/${id}`, {
    method: 'DELETE',
  })
}

export function getEvents() {
  return request<EventItem[]>('/api/events')
}

export function getSummary() {
  return request<DashboardSummaryResponse>('/api/summary')
}

export function getCapabilities() {
  return request<HostCapabilities>('/api/capabilities')
}

export function seedDemo() {
  return request<{ message: string; workloads: Workload[] }>('/api/demo/seed', {
    method: 'POST',
  })
}

export function getDemos() {
  return request<DemoDefinition[]>('/api/demos')
}

export function getDemo(id: string) {
  return request<DemoDefinition>(`/api/demos/${id}`)
}

export function runDemo(id: string) {
  return request<DemoRunResponse>(`/api/demos/${id}/run`, {
    method: 'POST',
  })
}

export function validateDemo(id: string) {
  return request<DemoValidation>(`/api/demos/${id}/validate`)
}
