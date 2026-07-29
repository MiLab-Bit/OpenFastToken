import { describe, it, expect } from 'vitest'
import {
  collectInvalidStatusCodeEntries,
  collectDisallowedStatusCodeRedirects,
  collectNewDisallowedStatusCodeRedirects,
} from './status-code-risk-guard'

describe('collectInvalidStatusCodeEntries', () => {
  it('returns [] for empty input', () => {
    expect(collectInvalidStatusCodeEntries('')).toEqual([])
  })
  it('returns [] for invalid JSON', () => {
    expect(collectInvalidStatusCodeEntries('not json')).toEqual([])
  })
  it('returns [] for non-object JSON', () => {
    expect(collectInvalidStatusCodeEntries('[1,2,3]')).toEqual([])
  })
  it('flags entries with bad keys or values', () => {
    const res = collectInvalidStatusCodeEntries('{"200":502,"abc":"502","200":"xyz"}')
    expect(res).toContain('abc → 502')
    expect(res).toContain('200 → xyz')
  })
  it('returns [] when all entries are valid', () => {
    expect(collectInvalidStatusCodeEntries('{"200":502,"404":503}')).toEqual([])
  })
})

describe('collectDisallowedStatusCodeRedirects', () => {
  it('returns [] for empty input', () => {
    expect(collectDisallowedStatusCodeRedirects('')).toEqual([])
  })
  it('flags 504/524 source codes mapped elsewhere', () => {
    const res = collectDisallowedStatusCodeRedirects('{"504":502,"524":500}')
    expect(res).toContain('504 -> 502')
    expect(res).toContain('524 -> 500')
  })
  it('ignores same-code mappings and non-504/524 sources', () => {
    expect(collectDisallowedStatusCodeRedirects('{"504":504,"200":502}')).toEqual([])
  })
  it('dedupes and sorts', () => {
    const res = collectDisallowedStatusCodeRedirects('{"524":500,"504":502}')
    expect(res).toEqual(['504 -> 502', '524 -> 500'])
  })
})

describe('collectNewDisallowedStatusCodeRedirects', () => {
  it('returns [] when current has no risky mappings', () => {
    expect(collectNewDisallowedStatusCodeRedirects('', '{"200":502}')).toEqual([])
  })
  it('returns only mappings newly introduced vs original', () => {
    const original = '{"504":502}'
    const current = '{"504":502,"524":500}'
    expect(collectNewDisallowedStatusCodeRedirects(original, current)).toEqual([
      '524 -> 500',
    ])
  })
  it('returns [] when nothing new', () => {
    const original = '{"504":502,"524":500}'
    expect(
      collectNewDisallowedStatusCodeRedirects(original, '{"504":502,"524":500}')
    ).toEqual([])
  })
})
