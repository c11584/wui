export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
  code?: string
  limit?: number
}

export class ApiError extends Error {
  code?: string
  limit?: number

  constructor(message: string, code?: string, limit?: number) {
    super(message)
    this.code = code
    this.limit = limit
  }
}
