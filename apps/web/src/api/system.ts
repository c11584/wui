import api from './client'

export interface SystemInfo {
  version: string
  os: string
  arch: string
  mode: 'admin' | 'agent'
}

export const systemApi = {
  getInfo: async (): Promise<SystemInfo> => {
    const response = await api.get<{ success: boolean; data: SystemInfo }>('/system/info')
    return response.data.data
  },
}
