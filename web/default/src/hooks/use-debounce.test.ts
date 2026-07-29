/*
 * Unit tests for hooks/use-debounce.ts
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useDebounce } from './use-debounce'

describe('useDebounce', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('returns the initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('a', 500))
    expect(result.current).toBe('a')
  })

  it('updates only after the delay elapses', () => {
    const { result, rerender } = renderHook(
      ({ v }: { v: string }) => useDebounce(v, 500),
      { initialProps: { v: 'a' } }
    )
    rerender({ v: 'b' })
    expect(result.current).toBe('a') // not yet
    act(() => {
      vi.advanceTimersByTime(500)
    })
    expect(result.current).toBe('b')
  })

  it('cancels the previous timer on rapid changes', () => {
    const { result, rerender } = renderHook(
      ({ v }: { v: string }) => useDebounce(v, 500),
      { initialProps: { v: 'a' } }
    )
    rerender({ v: 'b' })
    rerender({ v: 'c' })
    act(() => {
      vi.advanceTimersByTime(500)
    })
    expect(result.current).toBe('c')
  })
})
