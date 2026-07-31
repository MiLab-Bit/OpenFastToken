/*
 * Unit tests for hooks/use-minimum-loading-time.ts
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useMinimumLoadingTime } from './use-minimum-loading-time'

describe('useMinimumLoadingTime', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('shows the skeleton immediately while loading', () => {
    const { result } = renderHook(() => useMinimumLoadingTime(true, 1000))
    expect(result.current).toBe(true)
  })

  it('never shows the skeleton when loading is always false', () => {
    const { result } = renderHook(() => useMinimumLoadingTime(false, 1000))
    expect(result.current).toBe(false)
  })

  it('keeps the skeleton until the minimum time elapses after loading stops', () => {
    const { result, rerender } = renderHook(
      ({ l }: { l: boolean }) => useMinimumLoadingTime(l, 1000),
      { initialProps: { l: true } }
    )
    expect(result.current).toBe(true)
    rerender({ l: false })
    // elapsed ~0 -> still within minimum window
    expect(result.current).toBe(true)
    act(() => vi.advanceTimersByTime(1000))
    expect(result.current).toBe(false)
  })
})
