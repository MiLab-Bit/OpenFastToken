/*
Copyright (C) 2023-2026 OpenFastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@example.com
*/
import axios, { AxiosError, AxiosRequestConfig } from 'axios'
import i18next from 'i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'

// ============================================================================
// Type Definitions
// ============================================================================

interface CustomRequestConfig extends AxiosRequestConfig {
  disableDeduplicate?: boolean
  skipBusinessError?: boolean
  skipErrorHandler?: boolean
}

interface ApiErrorResponse {
  message?: string
  code?: string
  data?: unknown
}

interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

// ============================================================================
// Axios Instance Configuration
// ============================================================================

// Base URL: empty string for same-origin API requests
const baseURL = ''

// Create axios instance with default config
export const api = axios.create({
  baseURL,
  withCredentials: true, // Include cookies in cross-origin requests
  headers: {
    'Cache-Control': 'no-store', // Prevent caching
  },
})

// ============================================================================
// Request Deduplication
// ============================================================================

// Deduplicate concurrent GET requests to the same URL
// Prevents multiple identical requests from being sent simultaneously
const inFlightGet = new Map<string, Promise<unknown>>()
const originalGet = api.get.bind(api)

api.get = ((url: string, config?: CustomRequestConfig) => {
  if (config?.disableDeduplicate) return originalGet(url, config)

  const params = config?.params
    ? JSON.stringify(config.params)
    : '{}'
  const key = `${url}?${params}`

  // Return existing in-flight request if available
  if (inFlightGet.has(key)) return inFlightGet.get(key)!

  // Create new request and clean up after completion
  const req = originalGet(url, config).finally(() => inFlightGet.delete(key))
  inFlightGet.set(key, req)
  return req
}) as typeof api.get

// ============================================================================
// Response Interceptor
// ============================================================================

// Enhanced error handling with better user feedback
api.interceptors.response.use(
  (response) => {
    const config = response.config as CustomRequestConfig
    const skipBusiness = config.skipBusinessError

    // Unified business response format: { success, message, data }
    if (
      !skipBusiness &&
      response &&
      response.data &&
      typeof response.data.success === 'boolean'
    ) {
      if (!response.data.success) {
        // Enhanced business error handling
        const errorCode = response.data.code || 'UNKNOWN_ERROR'
        const errorMessage = getLocalizedErrorMessage(errorCode) || response.data.message || 'Request failed'
        
        // Show more specific error toasts
        toast.error(errorMessage, {
          description: `Error code: ${errorCode}`,
          duration: 5000,
        })
        
        // Log error for debugging (only in development)
        if (import.meta.env.DEV) {
          console.warn(`Business Error [${errorCode}]:`, response.data)
        }
      }
    }
    return response
  },
  (error: AxiosError<ApiErrorResponse>) => {
    const config = error.config as CustomRequestConfig
    const skip = config?.skipErrorHandler
    if (!skip) {
      handleHttpError(error)
    }
    return Promise.reject(error)
  }
)

// Helper function for HTTP error handling
function handleHttpError(error: AxiosError<ApiErrorResponse>): void {
  const status = error?.response?.status
  const isNetworkError = !error.response && error.message?.includes('Network Error')
  
  if (isNetworkError) {
    toast.error(i18next.t('Network error. Please check your connection.'), {
      duration: 6000,
    })
    return
  }
  
  switch (status) {
    case 401:
      toast.error(i18next.t('Session expired. Please sign in again.'), {
        duration: 6000,
      })
      try {
        useAuthStore.getState().auth.reset()
        // Redirect to sign-in page
        window.location.href = '/sign-in'
      } catch {
        /* empty */
      }
      break
    case 403:
      toast.error(i18next.t('You do not have permission to perform this action.'), {
        duration: 6000,
      })
      break
    case 404:
      toast.error(i18next.t('The requested resource was not found.'), {
        duration: 5000,
      })
      break
    case 429:
      toast.error(i18next.t('Too many requests. Please try again later.'), {
        duration: 8000,
      })
      break
    case 500:
      toast.error(i18next.t('Internal server error. Please try again later.'), {
        duration: 6000,
      })
      break
    default:
      const msg = error?.response?.data?.message || error?.message || i18next.t('An unexpected error occurred.')
      toast.error(msg, {
        duration: 5000,
      })
  }
  
  // Log error for debugging (only in development)
  if (import.meta.env.DEV) {
    console.error(`HTTP Error [${status}]:`, error)
  }
}

// Helper function to get localized error messages
function getLocalizedErrorMessage(errorCode: string): string | null {
  const errorMessages: Record<string, string> = {
    'INVALID_CREDENTIALS': i18next.t('Invalid username or password.'),
    'USER_NOT_FOUND': i18next.t('User not found.'),
    'TOKEN_EXPIRED': i18next.t('Your session has expired.'),
    'INSUFFICIENT_QUOTA': i18next.t('Insufficient quota available.'),
    'RATE_LIMIT_EXCEEDED': i18next.t('Rate limit exceeded. Please try again later.'),
    'VALIDATION_ERROR': i18next.t('Please check your input and try again.'),
  }
  
  return errorMessages[errorCode] || null
}

// ============================================================================
// Common Headers Utility
// ============================================================================

/**
 * Get user ID from localStorage
 */
function getUserId(): string | null {
  try {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem('uid')
    }
  } catch {
    /* empty */
  }
  return null
}

/**
 * Get common request headers (for both axios and SSE requests)
 */
export function getCommonHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  const uid = getUserId()
  if (uid) {
    headers['FastToken-User'] = uid
  }

  return headers
}

// ============================================================================
// Request Interceptor
// ============================================================================

// Attach user ID header for all requests
api.interceptors.request.use((config) => {
  const uid = getUserId()
  if (uid) {
    // Custom header for user identification
    config.headers['FastToken-User'] = uid
  }
  return config
})

// ============================================================================
// Common API Functions
// ============================================================================

// ----------------------------------------------------------------------------
// User APIs
// ----------------------------------------------------------------------------

// Get current user info
export async function getSelf() {
  const res = await api.get<ApiResponse>('/api/user/self', {
    // Avoid global 401 toast during guards/preloads
    skipErrorHandler: true,
  } as CustomRequestConfig)
  return res.data
}

// Get user available models
export async function getUserModels(): Promise<ApiResponse<string[]>> {
  const res = await api.get<ApiResponse<string[]>>('/api/user/models')
  return res.data
}

// Get user groups with descriptions and ratios
export async function getUserGroups(): Promise<ApiResponse<Record<string, { desc: string; ratio: number | string }>>> {
  const res = await api.get<ApiResponse<Record<string, { desc: string; ratio: number | string }>>>('/api/user/self/groups')
  return res.data
}

// ----------------------------------------------------------------------------
// System APIs
// ----------------------------------------------------------------------------

// Get system status
export async function getStatus(): Promise<Record<string, unknown>> {
  const res = await api.get<ApiResponse<Record<string, unknown>>>('/api/status')
  return res.data?.data as Record<string, unknown>
}

// Get system notice
export async function getNotice(): Promise<{
  success: boolean
  message?: string
  data?: string
}> {
  const res = await api.get('/api/notice')
  return res.data
}

// ============================================================================
// Export
// ============================================================================

export default api

// ============================================================================
// 2FA (Two-Factor Authentication)
// ============================================================================

export interface TwoFASetupResponse {
  secret?: string
  qr_code?: string
  [key: string]: unknown
}

export async function get2FAStatus(): Promise<ApiResponse<{ enabled: boolean; setup: boolean }>> {
  try {
    const res = await api.get<ApiResponse<{ enabled: boolean; setup: boolean }>>('/api/user/self/2fa/status')
    return res.data
   
  } catch (error: any) {
    if (error?.response?.status === 404) {
      // 2FA endpoint not implemented on backend
      return { success: true, data: { enabled: false, setup: false } }
    }
    throw error
  }
}

export async function setup2FA(): Promise<ApiResponse<TwoFASetupResponse>> {
  const res = await api.post<ApiResponse<TwoFASetupResponse>>('/api/user/self/2fa/setup')
  return res.data
}

export async function enable2FA(code: string): Promise<ApiResponse<null>> {
  const res = await api.post<ApiResponse<null>>('/api/user/self/2fa/enable', { code })
  return res.data
}

// ============================================================================
// 2FA - Disable & Regenerate Backup Codes
// ============================================================================

export async function disable2FA(code: string): Promise<ApiResponse<null>> {
  const res = await api.post<ApiResponse<null>>('/api/user/self/2fa/disable', { code })
  return res.data
}

export async function regenerate2FABackupCodes(code: string): Promise<ApiResponse<{ backup_codes: string[] }>> {
  const res = await api.post<ApiResponse<{ backup_codes: string[] }>>('/api/user/self/2fa/regenerate-backup', { code })
  return res.data
}