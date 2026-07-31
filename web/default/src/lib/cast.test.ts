/*
 * Unit tests for lib/cast.ts — typed structural cast helper.
 */
import { describe, it, expect } from 'vitest'
import { unsafeCast } from './cast'

describe('unsafeCast', () => {
  it('returns the same reference for objects', () => {
    const obj = { a: 1 }
    expect(unsafeCast<{ a: number }>(obj)).toBe(obj)
  })

  it('returns the same value for primitives', () => {
    expect(unsafeCast<number>('5' as unknown)).toBe('5')
    expect(unsafeCast<string>(42 as unknown)).toBe(42)
  })

  it('can reinterpret a string as a numeric id', () => {
    const row = { id: '1001', name: 'x' }
    const casted = unsafeCast<{ id: number; name: string }>(row)
    expect(casted.name).toBe('x')
  })
})
