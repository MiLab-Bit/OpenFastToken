import { describe, it, expect } from 'vitest'
import { buildQueryParams } from './query-params'

describe('buildQueryParams', () => {
  it('filters out undefined, null, and empty string but keeps 0', () => {
    const params = buildQueryParams({
      a: 1,
      b: '',
      c: null,
      d: undefined,
      e: 0,
      f: 'x',
    })
    expect(params.get('a')).toBe('1')
    expect(params.get('e')).toBe('0')
    expect(params.get('f')).toBe('x')
    expect(params.has('b')).toBe(false)
    expect(params.has('c')).toBe(false)
    expect(params.has('d')).toBe(false)
  })
  it('returns an empty param set for an empty object', () => {
    expect(buildQueryParams({}).toString()).toBe('')
  })
})
