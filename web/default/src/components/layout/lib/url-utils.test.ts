import { describe, it, expect } from 'vitest'
import { normalizeHref, checkIsActive } from './url-utils'

describe('normalizeHref', () => {
  it('strips query params', () => {
    expect(normalizeHref('/tokens?page=1')).toBe('/tokens')
  })

  it('removes trailing slashes for non-root paths', () => {
    expect(normalizeHref('/tokens/')).toBe('/tokens')
    expect(normalizeHref('/a/b/c/')).toBe('/a/b/c')
  })

  it('keeps the root slash', () => {
    expect(normalizeHref('/')).toBe('/')
  })
})

describe('checkIsActive', () => {
  it('matches an exact string url', () => {
    expect(checkIsActive('/tokens', { url: '/tokens' })).toBe(true)
  })

  it('matches on pathname when the current href carries query params', () => {
    expect(checkIsActive('/tokens?page=2', { url: '/tokens' })).toBe(true)
  })

  it('matches a sub-item url inside a collapsible group', () => {
    expect(
      checkIsActive('/tokens/new', { items: [{ url: '/tokens/new' }] })
    ).toBe(true)
  })

  it('matches a sub-item url carrying its own query only with an exact href', () => {
    expect(
      checkIsActive('/tokens/new', {
        items: [{ url: '/tokens/new?x=1' }],
      })
    ).toBe(false)
    expect(
      checkIsActive('/tokens/new?x=1', {
        items: [{ url: '/tokens/new?x=1' }],
      })
    ).toBe(true)
  })

  it('honors activeUrls', () => {
    expect(
      checkIsActive('/dashboard', { activeUrls: ['/dashboard', '/home'] })
    ).toBe(true)
  })

  it('matches first-level path for main navigation', () => {
    expect(checkIsActive('/tokens', { url: '/tokens/foo' }, true)).toBe(true)
    expect(checkIsActive('/channels', { url: '/tokens/foo' }, true)).toBe(false)
  })

  it('supports object urls (pathname only)', () => {
    expect(checkIsActive('/a', { url: { pathname: '/a' } })).toBe(true)
  })

  it('supports object urls with query (exact match required)', () => {
    expect(
      checkIsActive('/a?x=1', { url: { pathname: '/a', search: '?x=1' } })
    ).toBe(true)
    expect(
      checkIsActive('/a', { url: { pathname: '/a', search: '?x=1' } })
    ).toBe(false)
  })

  it('returns false when nothing matches', () => {
    expect(checkIsActive('/x', { url: '/y' })).toBe(false)
    expect(checkIsActive('/x', { items: [{ url: '/z' }] })).toBe(false)
  })

  it('returns false when there is no url to compare', () => {
    expect(checkIsActive('/x', {})).toBe(false)
  })
})
