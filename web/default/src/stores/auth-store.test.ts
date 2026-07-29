import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore, type AuthUser } from './auth-store'

const sampleUser: AuthUser = {
  id: 1,
  username: 'alice',
  role: 1,
  group: 'default',
}

describe('auth-store', () => {
  beforeEach(() => {
    window.localStorage.clear()
    // reset user to null for a clean slate
    useAuthStore.getState().auth.reset()
  })

  it('starts with no user', () => {
    expect(useAuthStore.getState().auth.user).toBeNull()
  })

  it('setUser stores and persists the user to localStorage', () => {
    useAuthStore.getState().auth.setUser(sampleUser)
    expect(useAuthStore.getState().auth.user).toEqual(sampleUser)
    const raw = window.localStorage.getItem('user')
    expect(raw).toBeTruthy()
    expect(JSON.parse(raw as string).username).toBe('alice')
  })

  it('setUser(null) clears the user', () => {
    useAuthStore.getState().auth.setUser(sampleUser)
    useAuthStore.getState().auth.setUser(null)
    expect(useAuthStore.getState().auth.user).toBeNull()
  })

  it('reset clears the persisted user', () => {
    useAuthStore.getState().auth.setUser(sampleUser)
    useAuthStore.getState().auth.reset()
    expect(useAuthStore.getState().auth.user).toBeNull()
    expect(window.localStorage.getItem('user')).toBeNull()
  })
})
