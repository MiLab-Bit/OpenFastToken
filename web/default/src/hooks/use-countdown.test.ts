/*
 * Unit tests for hooks/use-countdown.ts
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useCountdown } from './use-countdown'

describe('useCountdown', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('initializes with the configured seconds and inactive state', () => {
    const { result } = renderHook(() =>
      useCountdown({ initialSeconds: 3, autoStart: false })
    )
    expect(result.current.secondsLeft).toBe(3)
    expect(result.current.isActive).toBe(false)
  })

  it('starts and counts down, then resets and stops at 1', () => {
    const { result } = renderHook(() =>
      useCountdown({ initialSeconds: 3, autoStart: false })
    )
    act(() => result.current.start())
    expect(result.current.isActive).toBe(true)
    expect(result.current.secondsLeft).toBe(3)

    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.secondsLeft).toBe(2)
    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.secondsLeft).toBe(1)
    // next tick hits <=1: resets to initial and stops
    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.secondsLeft).toBe(3)
    expect(result.current.isActive).toBe(false)
  })

  it('accepts an explicit start duration', () => {
    const { result } = renderHook(() => useCountdown({ initialSeconds: 10 }))
    act(() => result.current.start(5))
    expect(result.current.secondsLeft).toBe(5)
  })

  it('stop halts the countdown', () => {
    const { result } = renderHook(() => useCountdown({ initialSeconds: 5 }))
    act(() => result.current.start())
    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.secondsLeft).toBe(4)
    act(() => result.current.stop())
    expect(result.current.isActive).toBe(false)
    const after = result.current.secondsLeft
    act(() => vi.advanceTimersByTime(2000))
    expect(result.current.secondsLeft).toBe(after)
  })

  it('reset returns to the initial seconds and inactive', () => {
    const { result } = renderHook(() => useCountdown({ initialSeconds: 5 }))
    act(() => result.current.start())
    act(() => vi.advanceTimersByTime(2000))
    expect(result.current.secondsLeft).toBe(3)
    act(() => result.current.reset())
    expect(result.current.secondsLeft).toBe(5)
    expect(result.current.isActive).toBe(false)
  })
})
