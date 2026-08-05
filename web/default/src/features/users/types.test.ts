/*
Copyright (C) 2023-2026 FastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import { describe, it, expect } from 'vitest'
import {
  userSchema,
  userStatusSchema,
  userRoleSchema,
  USER_STATUS_MAP,
  USER_ROLE_MAP,
} from './types'

describe('userStatusSchema', () => {
  it('accepts numeric statuses', () => {
    expect(userStatusSchema.safeParse(1).success).toBe(true)
    expect(userStatusSchema.safeParse(2).success).toBe(true)
    expect(userStatusSchema.safeParse(100).success).toBe(true)
  })

  it('rejects non-numeric values', () => {
    expect(userStatusSchema.safeParse('enabled').success).toBe(false)
    expect(userStatusSchema.safeParse(null).success).toBe(false)
  })
})

describe('userRoleSchema', () => {
  it('accepts numeric roles', () => {
    expect(userRoleSchema.safeParse(1).success).toBe(true)
    expect(userRoleSchema.safeParse(10).success).toBe(true)
  })
})

describe('userSchema', () => {
  const baseUser = {
    id: 1,
    username: 'alice',
    display_name: 'Alice',
    quota: 1000,
    used_quota: 0,
    request_count: 0,
    group: 'default',
    status: 1,
    role: 1,
  }

  it('parses a valid user', () => {
    const res = userSchema.safeParse(baseUser)
    expect(res.success).toBe(true)
  })

  it('rejects a non-numeric status', () => {
    const res = userSchema.safeParse({ ...baseUser, status: 'active' })
    expect(res.success).toBe(false)
  })

  it('allows an optional DeletedAt of any shape (z.unknown)', () => {
    const withString = userSchema.safeParse({
      ...baseUser,
      DeletedAt: '2024-01-01T00:00:00Z',
    })
    expect(withString.success).toBe(true)

    const withNull = userSchema.safeParse({ ...baseUser, DeletedAt: null })
    expect(withNull.success).toBe(true)

    const withObject = userSchema.safeParse({
      ...baseUser,
      DeletedAt: { deleted_at: '2024-01-01' },
    })
    expect(withObject.success).toBe(true)
  })
})

describe('USER_STATUS_MAP / USER_ROLE_MAP', () => {
  it('maps known status codes to labels', () => {
    expect(USER_STATUS_MAP[1]).toBe('enabled')
    expect(USER_STATUS_MAP[2]).toBe('disabled')
    expect(USER_STATUS_MAP[3]).toBe('pending')
  })

  it('maps known role codes to labels', () => {
    expect(USER_ROLE_MAP[1]).toBe('common')
    expect(USER_ROLE_MAP[10]).toBe('admin')
    expect(USER_ROLE_MAP[100]).toBe('root')
  })
})
