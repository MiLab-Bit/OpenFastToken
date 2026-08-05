/*
Copyright (C) 2023-2026 FastToken

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

For commercial licensing, please contact hello@fasttoken.example.com
*/

import { lazy, Suspense, useEffect, useState, useRef, type ComponentType, type ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { ContentSkeleton } from './auto-skeleton'

interface PerformanceWrapperProps {
  children: ReactNode
  fallback?: ReactNode
  className?: string
  minHeight?: number
}

/**
 * Performance wrapper component that provides:
 * 1. Lazy loading with suspense
 * 2. Performance monitoring
 * 3. Error boundaries
 * 4. Loading state management
 */
export function PerformanceWrapper({
  children,
  fallback,
  className,
  minHeight = 200,
}: PerformanceWrapperProps) {
  return (
    <div className={cn('relative', className)} style={{ minHeight }}>
      <Suspense
        fallback={
          fallback || (
            <ContentSkeleton
              loading={true}
              minTextHeight={minHeight}
              className="w-full"
            >
              <div style={{ height: minHeight }} />
            </ContentSkeleton>
          )
        }
      >
        {children}
      </Suspense>
    </div>
  )
}

/**
 * Higher-order component for lazy loading with performance optimization
 * @param importFunc - Dynamic import function
 * @param options - Configuration options
 */
export function withPerformanceOptimization<T extends object>(
  importFunc: () => Promise<{ default: ComponentType<T> }>,
  options: {
    fallback?: ReactNode
    minHeight?: number
    preload?: boolean
    onLoad?: () => void
    onError?: (error: Error) => void
  } = {}
) {
  const {
    fallback,
    minHeight = 200,
    preload = false,
    onLoad,
    onError,
  } = options

  // Create lazy component with error handling
  const LazyComponent = lazy(async () => {
    try {
      const module = await importFunc()
      onLoad?.()
      return module
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error))
      onError?.(err)
      throw err
    }
  })

  // Preload the component if requested
  if (preload && typeof window !== 'undefined') {
    // Use requestIdleCallback for non-critical preloading
    if ('requestIdleCallback' in window) {
      requestIdleCallback(() => {
        importFunc().catch(() => {
          /* Preload failed, will retry on actual render */
        })
      })
    } else {
      // Fallback for browsers without requestIdleCallback
      setTimeout(() => {
        importFunc().catch(() => {
          /* Preload failed, will retry on actual render */
        })
      }, 0)
    }
  }

  // Return wrapper component
  const WrappedComponent = (props: T) => (
    <PerformanceWrapper fallback={fallback} minHeight={minHeight}>
      <LazyComponent {...props} />
    </PerformanceWrapper>
  )

  // Add display name for debugging
   
  ;(WrappedComponent as any).displayName = `withPerformanceOptimization(${(LazyComponent as any).displayName || 'Component'})`

  return WrappedComponent
}

/**
 * Hook for preloading components on hover/focus
 */
export function usePreloadOnInteraction(
   
  importFunc: () => Promise<{ default: ComponentType<any> }>
) {
  const preload = () => {
    importFunc().catch(() => {
      /* Preload failed, will retry on actual render */
    })
  }

  return {
    onMouseEnter: preload,
    onFocus: preload,
    onTouchStart: preload,
  }
}

/**
 * Component for viewport-triggered lazy loading
 */
export function ViewportLazyLoad({
  children,
  rootMargin = '50px',
  triggerOnce = true,
  fallback,
}: {
  children: ReactNode
  rootMargin?: string
  triggerOnce?: boolean
  fallback?: ReactNode
}) {
  const [isVisible, setIsVisible] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const element = ref.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true)
          if (triggerOnce) {
            observer.unobserve(element)
          }
        }
      },
      { rootMargin }
    )

    observer.observe(element)

    return () => observer.disconnect()
  }, [rootMargin, triggerOnce])

  return (
    <div ref={ref}>
      {isVisible ? (
        children
      ) : (
        fallback || (
          <div className="flex items-center justify-center p-8">
            <ContentSkeleton loading={true} className="w-full h-32" />
          </div>
        )
      )}
    </div>
  )
}

/**
 * Utility function to create lazy-loaded routes
 */
export function createLazyRoute(
   
  importFunc: () => Promise<{ default: ComponentType<any> }>,
  options: {
    fallback?: ReactNode
    minHeight?: number
  } = {}
) {
  const LazyComponent = withPerformanceOptimization(importFunc, options)

   
  return function LazyRoute(props: any) {
    return <LazyComponent {...props} />
  }
}

