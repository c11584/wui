import api from './client'
import type { Tunnel } from '../types'
import { ApiError, type ApiResponse } from './types'

export const tunnelApi = {
  list: async (params?: { name?: string; protocol?: string; enabled?: string }): Promise<Tunnel[]> => {
    const queryParams = new URLSearchParams()
    if (params?.name) queryParams.append('name', params.name)
    if (params?.protocol) queryParams.append('protocol', params.protocol)
    if (params?.enabled) queryParams.append('enabled', params.enabled)

    const response = await api.get<ApiResponse<Tunnel[]>>('/tunnels', { params: queryParams })
    return response.data.data || []
  },

  get: async (id: number): Promise<Tunnel> => {
    const response = await api.get<ApiResponse<Tunnel>>(`/tunnels/${id}`)
    if (!response.data.success || !response.data.data) {
      throw new ApiError(response.data.message || 'Failed to get tunnel', response.data.code)
    }
    return response.data.data
  },

  create: async (data: Partial<Tunnel>): Promise<Tunnel> => {
    const response = await api.post<ApiResponse<Tunnel>>('/tunnels', data)
    if (!response.data.success || !response.data.data) {
      throw new ApiError(response.data.message || 'Failed to create tunnel', response.data.code, response.data.limit)
    }
    return response.data.data
  },

  update: async (id: number, data: Partial<Tunnel>): Promise<Tunnel> => {
    const response = await api.put<ApiResponse<Tunnel>>(`/tunnels/${id}`, data)
    if (!response.data.success || !response.data.data) {
      throw new ApiError(response.data.message || 'Failed to update tunnel', response.data.code)
    }
    return response.data.data
  },

  delete: async (id: number): Promise<void> => {
    const response = await api.delete<ApiResponse>(`/tunnels/${id}`)
    if (!response.data.success) {
      throw new ApiError(response.data.message || 'Failed to delete tunnel', response.data.code)
    }
  },

  start: async (id: number): Promise<void> => {
    const response = await api.post<ApiResponse>(`/tunnels/${id}/start`)
    if (!response.data.success) {
      throw new ApiError(response.data.message || 'Failed to start tunnel', response.data.code)
    }
  },

  stop: async (id: number): Promise<void> => {
    const response = await api.post<ApiResponse>(`/tunnels/${id}/stop`)
    if (!response.data.success) {
      throw new ApiError(response.data.message || 'Failed to stop tunnel', response.data.code)
    }
  },

  restart: async (id: number): Promise<void> => {
    const response = await api.post<ApiResponse>(`/tunnels/${id}/restart`)
    if (!response.data.success) {
      throw new ApiError(response.data.message || 'Failed to restart tunnel', response.data.code)
    }
  },

  getStats: async (id: number): Promise<{ upload: number; download: number; connections: number }> => {
    const response = await api.get<ApiResponse<{ upload: number; download: number; connections: number }>>(`/tunnels/${id}/stats`)
    if (!response.data.success || !response.data.data) {
      throw new ApiError(response.data.message || 'Failed to get stats', response.data.code)
    }
    return response.data.data
  },

  getConfig: async (id: number): Promise<{ config: string }> => {
    const response = await api.get<ApiResponse<{ config: string }>>(`/tunnels/${id}/config`)
    if (!response.data.success || !response.data.data) {
      throw new ApiError(response.data.message || 'Failed to get config', response.data.code)
    }
    return response.data.data
  },
}
