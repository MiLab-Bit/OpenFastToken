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

'use client'

import type { ErrorInfo, ReactNode } from 'react'
import { Component } from 'react'
import { Button } from '@/components/ui/button'
import { RefreshCw } from 'lucide-react'

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    }
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    // Update state so the next render will show the fallback UI
    return {
      hasError: true,
      error,
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Log error to console and potentially to an error reporting service
    console.error('Error caught by Error Boundary:', error, errorInfo)
    
    this.setState({
      error,
      errorInfo,
    })
  }

  handleRetry = (): void => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    })
  }

  render(): ReactNode {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Default fallback UI
      return (
        <div className="flex min-h-screen flex-col items-center justify-center bg-background p-4">
          <div className="w-full max-w-md space-y-4 rounded-lg border bg-card p-6 shadow-lg">
            <div className="space-y-2 text-center">
              <h2 className="text-2xl font-semibold tracking-tight text-destructive">
                Oops! Something went wrong
              </h2>
              <p className="text-sm text-muted-foreground">
                {this.state.error?.message || 'An unexpected error occurred'}
              </p>
            </div>

            {this.state.errorInfo && (
              <details className="mt-4 rounded-md bg-muted p-3 text-xs">
                <summary className="cursor-pointer font-medium">
                  Technical Details
                </summary>
                <pre className="mt-2 overflow-auto whitespace-pre-wrap break-words text-muted-foreground">
                  {this.state.error?.stack}
                  {'\n\n'}
                  {this.state.errorInfo.componentStack}
                </pre>
              </details>
            )}

            <div className="flex gap-2 pt-4">
              <Button
                onClick={this.handleRetry}
                className="flex-1 gap-2"
                variant="default"
              >
                <RefreshCw className="h-4 w-4" />
                Try Again
              </Button>
              <Button
                onClick={() => window.location.reload()}
                className="flex-1"
                variant="outline"
              >
                Reload Page
              </Button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
