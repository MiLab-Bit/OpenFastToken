import { describe, it, expect } from 'vitest'
import {
  parseUserSettings,
  getDisplayName,
  getUserInitials,
} from './format'

describe('parseUserSettings', () => {
  it('parses valid JSON', () => {
    expect(parseUserSettings('{"a":1}')).toEqual({ a: 1 })
  })
  it('returns empty object for empty or invalid input', () => {
    expect(parseUserSettings(undefined)).toEqual({})
    expect(parseUserSettings('bad')).toEqual({})
  })
})

describe('getDisplayName', () => {
  it('prefers display_name then username', () => {
    expect(getDisplayName({ display_name: 'Jane', username: 'jane' })).toBe('Jane')
    expect(getDisplayName({ username: 'jane' })).toBe('jane')
  })
  it('returns empty for missing user', () => {
    expect(getDisplayName(undefined)).toBe('')
  })
})

describe('getUserInitials', () => {
  it('uses first letters of two words', () => {
    expect(getUserInitials({ display_name: 'Jane Doe' })).toBe('JD')
  })
  it('uses first two chars for a single word', () => {
    expect(getUserInitials({ display_name: 'Jane' })).toBe('JA')
  })
  it('returns ? for missing user', () => {
    expect(getUserInitials(undefined)).toBe('?')
  })
})
