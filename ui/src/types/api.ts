export interface ApiResponse<T> {
  data: T
  status: 'success' | 'error'
  timestamp: string
  message?: string
}

export interface ApiError {
  message: string
  status: number
  details?: Record<string, unknown>
}

export interface PaginationParams {
  page?: number
  pageSize?: number
  filter?: string
  sortField?: string
  sortOrder?: 'asc' | 'desc'
}
