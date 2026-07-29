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
import { useEffect, useState } from 'react'
import { getHomePageContent } from '../api'
import type { HomePageContentResult } from '../types'

const STORAGE_KEY = 'home_page_content'

/**
 * Hook to load and manage custom home page content
 * Supports both Markdown/HTML content and iframe URLs
 */
export function useHomePageContent(): HomePageContentResult {
  const [content, setContent] = useState<string>('')
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    let mounted = true

    // One-time cache invalidation: prior versions of this hook trusted
    // localStorage and could keep rendering a stale custom home page after
    // a deploy. Clear the legacy cache key on first run so users see the
    // latest default landing sections (Hero/CTA/Footer) immediately.
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch {
      // localStorage may be unavailable (private mode, etc.) - safe to ignore.
    }

    const loadContent = async () => {
      // Don't trust localStorage cache: show default Hero immediately, then
      // let the API response decide. This prevents stale custom-home HTML
      // from blocking updates to the default landing sections (e.g. Hero,
      // CTA, Footer) across deploys.
      try {
        const timeoutPromise = new Promise<never>((_, reject) =>
          setTimeout(() => reject(new Error('timeout')), 3000)
        )
        const response = await Promise.race([
          getHomePageContent(),
          timeoutPromise,
        ])
        const { success, data } = response

        if (!mounted) return

        if (success && data) {
          setContent(data)
          localStorage.setItem(STORAGE_KEY, data)
        } else {
          // Clear content if API returns empty
          setContent('')
          localStorage.removeItem(STORAGE_KEY)
        }
      } catch (error) {
        if (!mounted) return
        // API failed (404 / network / timeout): clear any stale cache so the
        // user sees the latest default landing page instead of an old custom
        // home page from a previous deploy.
        localStorage.removeItem(STORAGE_KEY)
        setContent('')
         
        console.error('Failed to load home page content:', error)
      } finally {
        if (mounted) {
          setIsLoaded(true)
        }
      }
    }

    loadContent()

    return () => {
      mounted = false
    }
  }, [])

  let isUrl = false
  try {
    const url = new URL(content)
    isUrl = url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    // not a URL
  }

  return { content, isLoaded, isUrl }
}
