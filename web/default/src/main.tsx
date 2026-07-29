/*
Copyright (C) 2023-2026 鏅轰紒鎯?
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
import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import { AxiosError } from 'axios'
import {
  MutationCache,
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import i18next from 'i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { getStatus } from '@/lib/api'
import { installBuildMetadata } from '@/lib/build-metadata'
import '@/lib/dayjs'
import { applyFaviconToDom } from '@/lib/dom-utils'
import { initializeFrontendCache } from '@/lib/frontend-cache'
import { handleServerError } from '@/lib/handle-server-error'
import { DirectionProvider } from './context/direction-provider'
import { FontProvider } from './context/font-provider'
import { ThemeProvider } from './context/theme-provider'
import { SkinProvider } from './context/skin-provider'
import './l10n/config'
// Generated Routes
import { routeTree } from './routeTree.gen'
// Styles
import './styles/index.css'
import './styles/skins.css'
import { ErrorBoundary } from './components/error-boundary'

// Ensure VChart theme is initialized before any chart mounts (prevents white default theme flash)
// VChart theme is driven by our ThemeProvider (html.light/html.dark) via per-chart `theme` prop.
initializeFrontendCache()
installBuildMetadata()

// Enhanced QueryClient configuration with better performance and caching
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Optimized retry logic
      retry: (failureCount, error) => {
        // Don't retry for client errors (4xx)
        if (error instanceof AxiosError) {
          const status = error.response?.status ?? 0
          if (status >= 400 && status < 500 && status !== 408) {
            return false // Don't retry client errors
          }
        }
        
        // Retry logic based on environment
        if (import.meta.env.DEV) {
          return failureCount < 2 // Fewer retries in development
        }
        return failureCount < 3 // More retries in production
      },
      
      // Optimized refetch behavior
      refetchOnWindowFocus: import.meta.env.PROD ? 'always' : false,
      refetchOnReconnect: 'always',
      refetchOnMount: true,
      
      // Better cache management
      staleTime: 30 * 1000, // 30 seconds - data is fresh for 30s
      gcTime: 5 * 60 * 1000, // 5 minutes - garbage collection time
      
      // Performance optimizations
      structuralSharing: true, // Enable structural sharing for better performance
      placeholderData: (previousData: unknown) => previousData, // Keep previous data while loading
    },
    mutations: {
      // Enhanced error handling for mutations
      onError: (error) => {
        handleServerError(error)
        
        // Show specific error messages for common HTTP status codes
        if (error instanceof AxiosError) {
          const status = error.response?.status
          switch (status) {
            case 304:
              toast.error(i18next.t('Content not modified!'))
              break
            case 400:
              toast.error(i18next.t('Invalid request. Please check your input.'))
              break
            case 413:
              toast.error(i18next.t('File too large. Please upload a smaller file.'))
              break
            case 429:
              toast.error(i18next.t('Too many requests. Please try again later.'))
              break
            default:
              // Use default error handling
              break
          }
        }
      },
      // Optimistic updates helper
      onMutate: (_variables) => {
        // Return context for rollback
        return { timestamp: Date.now() }
      },
    },
  },
  // Enhanced query cache with better error handling
  queryCache: new QueryCache({
    onError: (error) => {
      if (error instanceof AxiosError) {
        const status = error.response?.status
        
        // Handle authentication errors
        if (status === 401) {
          toast.error(i18next.t('Session expired. Please sign in again.'))
          useAuthStore.getState().auth.reset()
          window.location.href = '/sign-in'
        } else if (status === 403) {
          toast.error(i18next.t('You do not have permission to access this resource.'))
        } else if (status === 500) {
          toast.error(i18next.t('Server error. Please try again later.'))
        }
        
        // Log error for debugging (only in development)
        if (import.meta.env.DEV) {
          console.error('Query Error:', error)
        }
      }
    },
    onSuccess: (data) => {
      // Optional: Log successful queries in development
      if (import.meta.env.DEV && data) {
        console.log('Query Success:', data)
      }
    },
  }),
  // Mutation cache for better mutation state management
  mutationCache: new MutationCache({
    onError: (error) => {
      if (import.meta.env.DEV) {
        console.error('Mutation Error:', error)
      }
    },
  }),
})

// Create a new router instance
const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
})

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

// Render the app
const rootElement = document.getElementById('root')!
// Set document.title and favicon from cached status, then refresh from network
;(function initSystemBranding() {
  try {
    if (typeof window === 'undefined' || typeof document === 'undefined') return
    const apply = (name: string) => {
      document.title = name
      const metaTitle = document.querySelector(
        'meta[name="title"]'
      ) as HTMLMetaElement | null
      if (metaTitle) metaTitle.setAttribute('content', name)
    }
    // Cache-first
    try {
      const saved = localStorage.getItem('status')
      if (saved) {
        const s = JSON.parse(saved)
        if (s?.system_name) apply(s.system_name)
        if (s?.logo) applyFaviconToDom(s.logo)
      }
    } catch {
      /* empty */
    }
    // Background refresh
    getStatus()
      .then((s) => {
        if (s?.system_name) {
          apply(s.system_name as string)
          try {
            localStorage.setItem('status', JSON.stringify(s))
          } catch {
            /* empty */
          }
        }
        if (s?.logo) applyFaviconToDom(s.logo as string)
      })
      .catch(() => {
        /* empty */
      })
  } catch {
    /* empty */
  }
})()
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <StrictMode>
      <ErrorBoundary>
        <QueryClientProvider client={queryClient}>
          <ThemeProvider>
            <SkinProvider>
              <FontProvider>
              <DirectionProvider>
                <RouterProvider router={router} />
              </DirectionProvider>
            </FontProvider>
            </SkinProvider>
          </ThemeProvider>
        </QueryClientProvider>
      </ErrorBoundary>
    </StrictMode>
  )
}
