/*
 * Unit tests for lib/exportCsv.ts — CSV export helper.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { toast } from 'sonner'
import { exportCsv, type ExportCsvOptions } from './exportCsv'

function makeResponse(opts: {
  contentType: string
  contentDisposition?: string | null
  ok?: boolean
  jsonBody?: unknown
  blob?: Blob
}) {
  const {
    contentType,
    contentDisposition = null,
    ok = true,
    jsonBody = {},
    blob = new Blob(['a,b'], { type: 'text/csv' }),
  } = opts
  return {
    ok,
    headers: {
      get: (key: string) => {
        const k = key.toLowerCase()
        if (k === 'content-type') return contentType
        if (k === 'content-disposition') return contentDisposition
        return null
      },
    },
    json: vi.fn().mockResolvedValue(jsonBody),
    blob: vi.fn().mockResolvedValue(blob),
  }
}

describe('exportCsv', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    if (!window.URL.createObjectURL) {
      // @ts-expect-error test stub
      window.URL.createObjectURL = vi.fn(() => 'blob:url')
      // @ts-expect-error test stub
      window.URL.revokeObjectURL = vi.fn()
    }
  })
  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('downloads a CSV on success and parses the filename', async () => {
    const res = makeResponse({
      contentType: 'text/csv',
      contentDisposition: 'attachment; filename="out.csv"',
    })
    vi.mocked(fetch).mockResolvedValue(res as unknown as Response)
    const success = vi.spyOn(toast, 'success')

    const anchors: HTMLAnchorElement[] = []
    const orig = document.createElement.bind(document)
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      const el = orig(tag as 'a')
      if (tag === 'a') anchors.push(el as HTMLAnchorElement)
      return el
    })

    await exportCsv('/api/export', { successMessage: 'done' })

    expect(fetch).toHaveBeenCalledOnce()
    expect(res.blob).toHaveBeenCalledOnce()
    expect(success).toHaveBeenCalledWith('done')
    expect(anchors[0]?.download).toBe('out.csv')
  })

  it('prefers the UTF-8 encoded filename', async () => {
    const res = makeResponse({
      contentType: 'text/csv',
      contentDisposition: "attachment; filename*=UTF-8''encoded%20name.csv",
    })
    vi.mocked(fetch).mockResolvedValue(res as unknown as Response)

    const anchors: HTMLAnchorElement[] = []
    const orig = document.createElement.bind(document)
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      const el = orig(tag as 'a')
      if (tag === 'a') anchors.push(el as HTMLAnchorElement)
      return el
    })

    await exportCsv('/api/export', { filename: 'fallback.csv' })
    expect(anchors[0]?.download).toBe('encoded name.csv')
  })

  it('falls back to the provided filename when CD is missing', async () => {
    const res = makeResponse({ contentType: 'text/csv' })
    vi.mocked(fetch).mockResolvedValue(res as unknown as Response)

    const anchors: HTMLAnchorElement[] = []
    const orig = document.createElement.bind(document)
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      const el = orig(tag as 'a')
      if (tag === 'a') anchors.push(el as HTMLAnchorElement)
      return el
    })

    await exportCsv('/api/export', { filename: 'fallback.csv' })
    expect(anchors[0]?.download).toBe('fallback.csv')
  })

  it('shows server message on JSON error payload', async () => {
    const res = makeResponse({
      contentType: 'application/json',
      jsonBody: { message: 'nope' },
    })
    vi.mocked(fetch).mockResolvedValue(res as unknown as Response)
    const err = vi.spyOn(toast, 'error')

    await exportCsv('/api/export')

    expect(res.json).toHaveBeenCalledOnce()
    expect(err).toHaveBeenCalledWith('nope')
  })

  it('shows fallback error when response not ok', async () => {
    const res = makeResponse({ contentType: 'text/csv', ok: false })
    vi.mocked(fetch).mockResolvedValue(res as unknown as Response)
    const err = vi.spyOn(toast, 'error')

    await exportCsv('/api/export', { errorMessage: 'failed' })

    expect(err).toHaveBeenCalledWith('failed')
  })

  it('shows retry error when fetch throws', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('network'))
    const err = vi.spyOn(toast, 'error')

    await exportCsv('/api/export')

    expect(err).toHaveBeenCalledWith('导出失败，请稍后重试')
  })

  it('appends query params when provided', async () => {
    const res = makeResponse({ contentType: 'text/csv' })
    vi.mocked(fetch).mockResolvedValue(res as unknown as Response)
    const opts: ExportCsvOptions = { params: { a: 1, b: '', c: null } }

    await exportCsv('/api/export', opts)

    const calledUrl = vi.mocked(fetch).mock.calls[0][0] as string
    expect(calledUrl).toContain('a=1')
    expect(calledUrl).not.toContain('b=')
    expect(calledUrl).not.toContain('c=')
  })
})
