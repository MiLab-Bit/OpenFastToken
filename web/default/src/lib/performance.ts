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

import { unsafeCast } from '@/lib/cast'

/**
 * Performance monitoring utilities for FastToken frontend
 * Provides tools for measuring and monitoring application performance
 */

// Store observer instances for cleanup
const observers: PerformanceObserver[] = [];

// Store event listener cleanup function
let navigationTimingCleanup: (() => void) | null = null;

/**
 * Initialize performance monitoring
 * Sets up Web Vitals monitoring and custom performance observers
 */
export function initPerformanceMonitoring(): void {
  if (typeof window === 'undefined') {
    return; // Only run in browser
  }

  // Monitor Core Web Vitals
  measureWebVitals();

  // Monitor resource loading
  monitorResourceTiming();

  // Monitor long tasks
  monitorLongTasks();

  // Monitor navigation timing
  monitorNavigationTiming();

  // Log performance marks
  logPerformanceMarks();
}

/**
 * Cleanup performance monitoring
 * Disconnects all observers and removes event listeners
 */
export function cleanupPerformanceMonitoring(): void {
  // Disconnect all PerformanceObservers
  observers.forEach(observer => {
    try {
      observer.disconnect();
    } catch {
      // Ignore errors during cleanup
    }
  });
  observers.length = 0;

  // Remove event listener if cleanup function exists
  if (navigationTimingCleanup) {
    navigationTimingCleanup();
    navigationTimingCleanup = null;
  }
}

/**
 * Measure Core Web Vitals using web-vitals library (if available)
 */
function measureWebVitals(): void {
  try {
    // Dynamically import web-vitals to avoid bundling if not needed
    import('web-vitals').then(({ onCLS, onFCP, onFID, onLCP, onTTFB }) => {
       
      onCLS(sendToAnalytics as any)
       
      onFCP(sendToAnalytics as any)
       
      onFID(sendToAnalytics as any)
       
      onLCP(sendToAnalytics as any)
       
      onTTFB(sendToAnalytics as any)
    }).catch(() => {
      // web-vitals not available, use native Performance API
      measureWebVitalsNative();
    });
  } catch {
    measureWebVitalsNative();
  }
}

/**
 * Native Web Vitals measurement using Performance API
 */
function measureWebVitalsNative(): void {
  // Measure LCP
  const lcpObserver = new PerformanceObserver((entryList) => {
    for (const entry of entryList.getEntries()) {
      const lcp = entry.startTime;
      sendToAnalytics('LCP', lcp);
    }
  });
  lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });
  observers.push(lcpObserver);

  // Measure FCP
  const fcpObserver = new PerformanceObserver((entryList) => {
    for (const entry of entryList.getEntries()) {
      const fcp = entry.startTime;
      sendToAnalytics('FCP', fcp);
    }
  });
  fcpObserver.observe({ type: 'paint', buffered: true });
  observers.push(fcpObserver);

  // Measure CLS
  const clsObserver = new PerformanceObserver((entryList) => {
    for (const entry of entryList.getEntries()) {
      // Use proper type checking instead of `as any`
      if ('hadRecentInputChange' in entry && !entry.hadRecentInputChange) {
        const layoutShiftEntry = unsafeCast<LayoutShift>(entry);
        const cls = layoutShiftEntry.value;
        sendToAnalytics('CLS', cls);
      }
    }
  });
  clsObserver.observe({ type: 'layout-shift', buffered: true });
  observers.push(clsObserver);
}

/**
 * Monitor resource loading performance
 */
function monitorResourceTiming(): void {
  const resourceObserver = new PerformanceObserver((entryList) => {
    for (const entry of entryList.getEntries()) {
      const resource = entry as PerformanceResourceTiming;
      
      // Log slow resources (>1s)
      if (resource.duration > 1000) {
        if (import.meta.env.DEV) {
          console.warn(`Slow resource: ${resource.name} took ${resource.duration.toFixed(2)}ms`);
        }
        
        // Send to analytics if needed
        if (import.meta.env.PROD) {
          sendResourceTimingToAnalytics(resource);
        }
      }
    }
  });
  resourceObserver.observe({ type: 'resource', buffered: true });
  observers.push(resourceObserver);
}

/**
 * Monitor long tasks (>50ms)
 */
function monitorLongTasks(): void {
  const longTaskObserver = new PerformanceObserver((entryList) => {
    for (const entry of entryList.getEntries()) {
      const task = entry as PerformanceEntry;
      if (import.meta.env.DEV) {
        console.warn(`Long task detected: ${task.duration.toFixed(2)}ms`);
      }
      
      // Send to analytics
      if (import.meta.env.PROD) {
        sendToAnalytics('LONG_TASK', task.duration);
      }
    }
  });
  longTaskObserver.observe({ type: 'longtask', buffered: true });
  observers.push(longTaskObserver);
}

/**
 * Monitor navigation timing
 */
function monitorNavigationTiming(): void {
  const handleLoad = () => {
    // Use setTimeout to ensure all timing data is available
    const timeoutId = setTimeout(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      
      const metrics = {
        dns: navigation.domainLookupEnd - navigation.domainLookupStart,
        tcp: navigation.connectEnd - navigation.connectStart,
        ttfb: navigation.responseStart - navigation.requestStart,
        download: navigation.responseEnd - navigation.responseStart,
        domInteractive: navigation.domInteractive - navigation.startTime,
        domComplete: navigation.domComplete - navigation.startTime,
        loadEvent: navigation.loadEventEnd - navigation.loadEventStart,
      };
      
      if (import.meta.env.DEV) {
        console.log('Navigation Timing:', metrics);
      }
      
      // Send to analytics
      if (import.meta.env.PROD) {
        sendNavigationTimingToAnalytics(metrics);
      }
    }, 0);

    // Cleanup timeout if component unmounts
    return () => clearTimeout(timeoutId);
  };

  window.addEventListener('load', handleLoad);
  
  // Store cleanup function
  navigationTimingCleanup = () => {
    window.removeEventListener('load', handleLoad);
  };
}

/**
 * Log performance marks and measures
 */
function logPerformanceMarks(): void {
  const marksObserver = new PerformanceObserver((entryList) => {
    for (const entry of entryList.getEntries()) {
      if (import.meta.env.DEV) {
        console.log(`Performance ${entry.entryType}: ${entry.name}`, entry);
      }
    }
  });
  marksObserver.observe({ entryTypes: ['mark', 'measure'] });
  observers.push(marksObserver);
}

/**
 * Send metrics to analytics service
 */
function sendToAnalytics(metric: string, value: number): void {
  // In development, just log to console
  if (import.meta.env.DEV) {
    console.log(`Performance Metric - ${metric}: ${value}`);
    return;
  }

  // In production, send to analytics service
  // Example: send to Google Analytics, Sentry, or custom endpoint
  const analyticsEndpoint = import.meta.env.VITE_ANALYTICS_ENDPOINT;
  if (analyticsEndpoint) {
    fetch(analyticsEndpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        metric,
        value,
        timestamp: Date.now(),
        url: window.location.pathname,
      }),
    }).catch(() => {
      // Silently fail - don't block the application
    });
  }
}

/**
 * Send resource timing to analytics
 */
function sendResourceTimingToAnalytics(resource: PerformanceResourceTiming): void {
  sendToAnalytics('RESOURCE_TIMING', resource.duration);
}

/**
 * Send navigation timing to analytics
 */
function sendNavigationTimingToAnalytics(metrics: Record<string, number>): void {
  Object.entries(metrics).forEach(([key, value]) => {
    sendToAnalytics(`NAVIGATION_${key.toUpperCase()}`, value);
  });
}

/**
 * Create a performance marker
 */
export function markPerformance(name: string): void {
  performance.mark(name);
}

/**
 * Measure performance between two marks
 */
export function measurePerformance(name: string, startMark: string, endMark?: string): void {
  if (endMark) {
    performance.measure(name, startMark, endMark);
  } else {
    performance.measure(name, startMark);
  }
}

/**
 * Get all performance entries
 */
export function getPerformanceEntries(): PerformanceEntryList {
  return performance.getEntries();
}

/**
 * Clear performance entries
 */
export function clearPerformanceEntries(): void {
  performance.clearMarks();
  performance.clearMeasures();
}

/**
 * Measure component render time
 */
export function measureComponentRender<T>(componentName: string, renderFn: () => T): T {
  const startMark = `${componentName}-render-start`;
  const endMark = `${componentName}-render-end`;
  
  performance.mark(startMark);
  const result = renderFn();
  performance.mark(endMark);
  
  performance.measure(componentName, startMark, endMark);
  
  return result;
}

/**
 * Debounce performance measurement to avoid excessive measurements
 */
export function debouncePerformanceMeasurement(
  measurementFn: () => void,
  delay: number = 100
): () => void {
  let timeoutId: ReturnType<typeof setTimeout>;
  
  return () => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(measurementFn, delay);
  };
}

// Auto-initialize in development
if (import.meta.env.DEV && typeof window !== 'undefined') {
  // Wait for DOM to be ready
  if (document.readyState === 'loading') {
    const domContentLoadedHandler = () => {
      initPerformanceMonitoring();
      document.removeEventListener('DOMContentLoaded', domContentLoadedHandler);
    };
    document.addEventListener('DOMContentLoaded', domContentLoadedHandler);
  } else {
    initPerformanceMonitoring();
  }
}
