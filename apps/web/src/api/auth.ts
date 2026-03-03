import api from './client'
import type { User } from '../types'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  password: string
}

export interface UpdateUserRequest {
  email?: string
  password?: string
}

export interface UpdateUserAdminRequest {
  email?: string
  password?: string
  maxTunnels?: number
  maxTraffic?: number
  status?: string
}

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/login', data)
    return response.data
  },

  register: async (data: RegisterRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/register', data)
    return response.data
  },

  forgotPassword: async (data: ForgotPasswordRequest): Promise<void> => {
    await api.post('/auth/forgot-password', data)
  },

  resetPassword: async (data: ResetPasswordRequest): Promise<void> => {
    await api.post('/auth/reset-password', data)
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await api.get<User>('/user')
    return response.data
  },

  updateCurrentUser: async (data: UpdateUserRequest): Promise<void> => {
    await api.put('/user', data)
  },
}
