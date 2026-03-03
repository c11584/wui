import axios from 'axios'
import { ApiError } from './types'

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.request.use((config) => {
  const authStorage = localStorage.getItem('auth-storage')
  if (authStorage) {
    try {
      const { state } = JSON.parse(authStorage)
      if (state?.token) {
        config.headers.Authorization = `Bearer ${state.token}`
      }
    } catch (e) {
      console.error('Failed to parse auth storage:', e)
    }
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth-storage')
      window.location.href = '/login'
    }
    
    const responseData = error.response?.data
    if (responseData) {
      const apiError = new ApiError(
        responseData.message || error.message,
        responseData.code,
        responseData.limit
      )
      return Promise.reject(apiError)
    }
    
    return Promise.reject(error)
  }
)

export default api
