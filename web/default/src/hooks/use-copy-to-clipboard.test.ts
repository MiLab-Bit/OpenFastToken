/*
 * Unit tests for hooks/use-copy-to-clipboard.ts
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { toast } from 'sonner'
import { useCopyToClipboard } from './use-copy-to-clipboard'

describe('useCopyToClipboard', () => {
  beforeEach(() => {
    // navigator.clipboard.writeText is stubbed in the global setup
    vi.spyOn(toast, 'success')
    vi.spyOn(toast, 'error')
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('copies text and records copiedText', async () => {
    const { result } = renderHook(() => useCopyToClipboard())
    let ok = false
    await act(async () => {
      ok = await result.current.copyToClipboard('hello')
    })
    expect(ok).toBe(true)
    expect(result.current.copiedText).toBe('hello')
    expect(toast.success).toHaveBeenCalled()
  })

  it('does not notify when notify=false', async () => {
    const { result } = renderHook(() => useCopyToClipboard({ notify: false }))
    await act(async () => {
      await result.current.copyToClipboard('x')
    })
    expect(toast.success).not.toHaveBeenCalled()
  })

  it('uses a custom success message', async () => {
    const { result } = renderHook(() =>
      useCopyToClipboard({ successMessage: 'copied!' })
    )
    await act(async () => {
      await result.current.copyToClipboard('x')
    })
    expect(toast.success).toHaveBeenCalledWith('copied!')
  })

  it('reports failure and shows an error when copy fails', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: () => Promise.reject(new Error('nope')) },
      configurable: true,
    })
    // @ts-expect-error test stub
    document.execCommand = () => false

    const { result } = renderHook(() =>
      useCopyToClipboard({ errorMessage: 'copy failed' })
    )
    let ok = true
    await act(async () => {
      ok = await result.current.copyToClipboard('x')
    })
    expect(ok).toBe(false)
    expect(result.current.copiedText).toBeNull()
    expect(toast.error).toHaveBeenCalledWith('copy failed')

    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: () => Promise.resolve() },
      configurable: true,
    })
  })
})
