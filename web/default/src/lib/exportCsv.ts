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
*/
import { toast } from 'sonner'

export interface ExportCsvOptions {
  params?: Record<string, unknown>
  filename?: string
  successMessage?: string
  errorMessage?: string
}

function triggerDownload(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)
}

function parseFilename(
  contentDisposition: string | null,
  fallback: string
): string {
  if (!contentDisposition) return fallback
  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match) return decodeURIComponent(utf8Match[1])
  const match = contentDisposition.match(/filename="?([^";]+)"?/i)
  if (match) return match[1]
  return fallback
}

/**
 * Trigger a backend CSV export download.
 * Uses native fetch (not the axios instance) so the raw CSV blob is not
 * intercepted by the JSON business-error handler.
 */
export async function exportCsv(
  url: string,
  options: ExportCsvOptions = {}
): Promise<void> {
  const { params, filename: fallbackName, successMessage, errorMessage } =
    options
  try {
    const query = new URLSearchParams()
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          query.append(key, String(value))
        }
      })
    }
    const sep = url.includes('?') ? '&' : '?'
    const full = query.toString() ? `${url}${sep}${query.toString()}` : url

    const res = await fetch(full, {
      method: 'GET',
      credentials: 'include',
      headers: { Accept: 'text/csv' },
    })

    const contentType = res.headers.get('Content-Type') || ''

    // Backend returns JSON (success:false) on error with HTTP 200.
    if (contentType.includes('application/json')) {
      try {
        const data = await res.json()
        toast.error(data?.message || errorMessage || '导出失败')
      } catch {
        toast.error(errorMessage || '导出失败')
      }
      return
    }

    if (!res.ok) {
      toast.error(errorMessage || '导出失败')
      return
    }

    const blob = await res.blob()
    const cd = res.headers.get('Content-Disposition')
    const name = parseFilename(cd, fallbackName || 'export.csv')
    triggerDownload(blob, name)
    if (successMessage) toast.success(successMessage)
  } catch {
    toast.error(errorMessage || '导出失败，请稍后重试')
  }
}
