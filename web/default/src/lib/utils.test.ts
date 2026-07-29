/*
 * Unit tests for lib/utils.ts — general UI helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  cn,
  sanitizeCssVariableName,
  getPageNumbers,
  truncateText,
  tryPrettyJson,
} from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('a', 'b')).toBe('a b')
  })
  it('dedupes conflicting tailwind classes via tailwind-merge', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })
  it('handles conditional falsy values', () => {
    expect(cn('a', false, null, undefined, 'b')).toBe('a b')
  })
})

describe('sanitizeCssVariableName', () => {
  it('replaces dots, spaces and slashes with hyphens', () => {
    expect(sanitizeCssVariableName('gpt-3.5-turbo')).toBe('gpt-3-5-turbo')
    expect(sanitizeCssVariableName('a b/c')).toBe('a-b-c')
  })
  it('strips other disallowed characters', () => {
    expect(sanitizeCssVariableName('model@name!')).toBe('modelname')
  })
})

describe('getPageNumbers', () => {
  it('shows all pages when total <= 5', () => {
    expect(getPageNumbers(1, 5)).toEqual([1, 2, 3, 4, 5])
    expect(getPageNumbers(3, 3)).toEqual([1, 2, 3])
  })
  it('near the beginning: 1 2 3 4 ... N', () => {
    expect(getPageNumbers(1, 10)).toEqual([1, 2, 3, 4, '...', 10])
    expect(getPageNumbers(3, 10)).toEqual([1, 2, 3, 4, '...', 10])
  })
  it('near the end: 1 ... N-3 N-2 N-1 N', () => {
    expect(getPageNumbers(10, 10)).toEqual([1, '...', 7, 8, 9, 10])
    expect(getPageNumbers(8, 10)).toEqual([1, '...', 7, 8, 9, 10])
  })
  it('in the middle: 1 ... c-1 c c+1 ... N', () => {
    expect(getPageNumbers(5, 10)).toEqual([1, '...', 4, 5, 6, '...', 10])
  })
})

describe('truncateText', () => {
  it('returns text unchanged when shorter than max', () => {
    expect(truncateText('abc', 5)).toBe('abc')
  })
  it('returns null/empty as-is', () => {
    expect(truncateText('', 5)).toBe('')
    expect(truncateText(null as unknown as string, 5)).toBe(null)
  })
  it('appends ellipsis when longer', () => {
    expect(truncateText('hello world', 5)).toBe('hello...')
  })
})

describe('tryPrettyJson', () => {
  it('returns empty string for empty input', () => {
    expect(tryPrettyJson('')).toBe('')
    expect(tryPrettyJson(null as unknown as string)).toBe('')
  })
  it('pretty prints valid JSON', () => {
    const out = tryPrettyJson('{"a":1}')
    expect(out).toContain('\n')
    expect(out).toContain('"a": 1')
  })
  it('falls back to raw text for invalid JSON', () => {
    expect(tryPrettyJson('not json {')).toBe('not json {')
  })
})
