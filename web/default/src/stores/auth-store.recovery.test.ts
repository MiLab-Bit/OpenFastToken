import { describe, it, expect, beforeEach, vi } from 'vitest'

describe('auth-store recovery from localStorage', () => {
  beforeEach(() => {
    window.localStorage.clear()
    vi.resetModules()
  })

  it('restores a valid saved user on init', async () => {
    const saved = { id: 1, username: 'alice', role: 1, group: 'default' }
    window.localStorage.setItem('user', JSON.stringify(saved))
    const mod = await import('./auth-store')
    expect(mod.useAuthStore.getState().auth.user).toEqual(saved)
  })

  it('falls back to null and clears corrupt user JSON on init', async () => {
    window.localStorage.setItem('user', '{ this is not valid json')
    const mod = await import('./auth-store')
    expect(mod.useAuthStore.getState().auth.user).toBeNull()
    expect(window.localStorage.getItem('user')).toBeNull()
  })

  it('falls back to null for a non-JSON string on init', async () => {
    window.localStorage.setItem('user', 'plain-string-not-json')
    const mod = await import('./auth-store')
    expect(mod.useAuthStore.getState().auth.user).toBeNull()
    expect(window.localStorage.getItem('user')).toBeNull()
  })
})
