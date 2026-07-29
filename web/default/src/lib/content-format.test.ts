/*
 * Unit tests for lib/content-format.ts — content sniffing helpers.
 */
import { describe, it, expect } from 'vitest'
import { isHttpUrl, isLikelyHtml } from './content-format'

describe('isHttpUrl', () => {
  it('accepts http and https URLs', () => {
    expect(isHttpUrl('https://example.com/a')).toBe(true)
    expect(isHttpUrl('http://example.com')).toBe(true)
  })
  it('rejects non-http protocols and bare strings', () => {
    expect(isHttpUrl('ftp://example.com')).toBe(false)
    expect(isHttpUrl('example.com')).toBe(false)
    expect(isHttpUrl('not a url')).toBe(false)
    expect(isHttpUrl('')).toBe(false)
  })
})

describe('isLikelyHtml', () => {
  it('detects html by doctype and tags', () => {
    expect(isLikelyHtml('<!doctype html><html><head></head>')).toBe(true)
    expect(isLikelyHtml('<body><p>hi</p></body>')).toBe(true)
    expect(isLikelyHtml('<script>alert(1)</script>')).toBe(true)
    expect(isLikelyHtml('<div>plain</div>')).toBe(true)
  })
  it('returns false for plain text', () => {
    expect(isLikelyHtml('just some text')).toBe(false)
    expect(isLikelyHtml('{"a":1}')).toBe(false)
  })
})
