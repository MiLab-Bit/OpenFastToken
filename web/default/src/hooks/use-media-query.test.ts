/*
 * Unit tests for hooks/use-media-query.ts
 */
import { describe, it, expect, afterEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useMediaQuery } from './use-media-query'

function stubMatchMedia(matches: boolean) {
  const original = window.matchMedia
  window.matchMedia = ((query: string) => ({
    matches,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  })) as typeof window.matchMedia
  return () => {
    window.matchMedia = original
  }
}

describe('useMediaQuery', () => {
  const restoreFns: Array<() => void> = []
  afterEach(() => {
    while (restoreFns.length) restoreFns.pop()?.()
  })

  it('returns false by default (non-matching stub)', () => {
    restoreFns.push(stubMatchMedia(false))
    const { result } = renderHook(() => useMediaQuery('(max-width: 600px)'))
    expect(result.current).toBe(false)
  })

  it('reflects a matching media query', () => {
    restoreFns.push(stubMatchMedia(true))
    const { result } = renderHook(() => useMediaQuery('(max-width: 600px)'))
    expect(result.current).toBe(true)
  })
})
