/*
Copyright (C) 2023-2026 OpenFastToken

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
import i18n from '@/l10n/config'
import { getRedemptionFormSchema } from './redemption-form'

// The real i18n instance's `t` is a base `TFunction`, matching the schema's
// `getRedemptionFormSchema(t: TFunction)` signature.
const t = i18n.t

describe('getRedemptionFormSchema', () => {
  it('accepts a valid form payload', () => {
    const schema = getRedemptionFormSchema(t)
    const res = schema.safeParse({ name: 'GIFT', quota_dollars: 10, count: 1 })
    expect(res.success).toBe(true)
  })

  it('rejects an empty name (min length 1)', () => {
    const schema = getRedemptionFormSchema(t)
    const res = schema.safeParse({ name: '', quota_dollars: 10 })
    expect(res.success).toBe(false)
  })

  it('rejects a name over the max length (20)', () => {
    const schema = getRedemptionFormSchema(t)
    const res = schema.safeParse({
      name: 'a'.repeat(21),
      quota_dollars: 10,
    })
    expect(res.success).toBe(false)
  })

  it('rejects a negative quota (min 0)', () => {
    const schema = getRedemptionFormSchema(t)
    const res = schema.safeParse({ name: 'GIFT', quota_dollars: -1 })
    expect(res.success).toBe(false)
  })

  it('rejects a count over the max (100)', () => {
    const schema = getRedemptionFormSchema(t)
    const res = schema.safeParse({ name: 'GIFT', quota_dollars: 10, count: 200 })
    expect(res.success).toBe(false)
  })

  it('allows optional count and expired_time to be omitted', () => {
    const schema = getRedemptionFormSchema(t)
    const res = schema.safeParse({ name: 'GIFT', quota_dollars: 5 })
    expect(res.success).toBe(true)
  })
})
