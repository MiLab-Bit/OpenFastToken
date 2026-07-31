import { describe, it, expect } from 'vitest'
import {
  isValidBackupCode,
  formatBackupCode,
  cleanBackupCode,
  isValidEmail,
  isValidOTP,
} from './validation'

describe('isValidBackupCode', () => {
  it('accepts XXXX-XXXX format', () => {
    expect(isValidBackupCode('ABCD-1234')).toBe(true)
    expect(isValidBackupCode('abcd-1234')).toBe(true)
  })
  it('rejects without hyphen or wrong length', () => {
    expect(isValidBackupCode('ABCD1234')).toBe(false)
    expect(isValidBackupCode('ABC-123')).toBe(false)
    expect(isValidBackupCode('')).toBe(false)
  })
})

describe('formatBackupCode', () => {
  it('adds a hyphen after 4 chars', () => {
    expect(formatBackupCode('abcd1234')).toBe('ABCD-1234')
  })
  it('uppercases and strips non-alphanumeric', () => {
    expect(formatBackupCode('ab-cd-12-34')).toBe('ABCD-1234')
  })
  it('truncates to 8 characters', () => {
    expect(formatBackupCode('abcdefghij')).toBe('ABCD-EFGH')
  })
  it('leaves short input without hyphen', () => {
    expect(formatBackupCode('abc')).toBe('ABC')
  })
})

describe('cleanBackupCode', () => {
  it('removes hyphens', () => {
    expect(cleanBackupCode('ABCD-1234')).toBe('ABCD1234')
  })
})

describe('isValidEmail', () => {
  it('accepts valid emails', () => {
    expect(isValidEmail('a@b.com')).toBe(true)
    expect(isValidEmail('user.name@sub.domain.io')).toBe(true)
  })
  it('rejects invalid emails', () => {
    expect(isValidEmail('a@b')).toBe(false)
    expect(isValidEmail('abc')).toBe(false)
    expect(isValidEmail('a b@c.com')).toBe(false)
  })
})

describe('isValidOTP', () => {
  it('accepts 4-8 digit codes', () => {
    expect(isValidOTP('1234')).toBe(true)
    expect(isValidOTP('12345678')).toBe(true)
  })
  it('rejects too short, too long, or non-numeric', () => {
    expect(isValidOTP('123')).toBe(false)
    expect(isValidOTP('123456789')).toBe(false)
    expect(isValidOTP('12a4')).toBe(false)
  })
})
