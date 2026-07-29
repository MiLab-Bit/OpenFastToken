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
import {
  formatCurrencyFromUSD,
  formatBillingCurrencyFromUSD,
  formatQuotaWithCurrency,
  formatLocalCurrencyAmount,
  getCurrencyLabel,
  isCurrencyDisplayEnabled,
  getCurrencyDisplay,
  getDisplayMeta,
  isCurrencyDisplayType,
  parseCurrencyDisplayType,
  type CurrencyDisplayConfig,
} from './currency'

describe('currency formatting (system units → 元)', () => {
  it('formats a positive quota with two decimals', () => {
    // 720000 system units = 1.44 元 (1 元 = 500000 system units)
    expect(formatCurrencyFromUSD(720000)).toBe('¥1.44')
  })

  it('returns "-" for null / undefined / NaN input', () => {
    expect(formatCurrencyFromUSD(null)).toBe('-')
    expect(formatCurrencyFromUSD(undefined)).toBe('-')
    expect(formatCurrencyFromUSD(Number.NaN)).toBe('-')
  })

  it('formats zero using small-digit precision', () => {
    expect(formatCurrencyFromUSD(0)).toBe('¥0')
  })

  it('formats negative quotas', () => {
    expect(formatCurrencyFromUSD(-720000)).toBe('-¥1.44')
  })

  it('billing and quota helpers behave like the base formatter', () => {
    expect(formatBillingCurrencyFromUSD(720000)).toBe('¥1.44')
    expect(formatQuotaWithCurrency(720000)).toBe('¥1.44')
  })

  it('formats an already-yuan amount', () => {
    expect(formatLocalCurrencyAmount(1440)).toBe('¥1,440')
    expect(formatLocalCurrencyAmount(null)).toBe('-')
  })
})

describe('currency display config', () => {
  it('returns the static currency label', () => {
    expect(getCurrencyLabel()).toBe('元')
  })

  it('reports display as enabled', () => {
    expect(isCurrencyDisplayEnabled()).toBe(true)
  })

  it('returns the fixed display structure', () => {
    expect(getCurrencyDisplay()).toEqual({
      config: { quotaPerUnit: 500000 },
      meta: { kind: 'currency', exchangeRate: 1 },
    })
  })

  it('getDisplayMeta returns the currency kind regardless of config', () => {
    const config: CurrencyDisplayConfig = { kind: 'tokens', quotaPerUnit: 100 }
    expect(getDisplayMeta(config)).toEqual({ kind: 'currency' })
  })
})

describe('currency display type guards', () => {
  it('isCurrencyDisplayType recognizes CNY / TOKENS', () => {
    expect(isCurrencyDisplayType('CNY')).toBe(true)
    expect(isCurrencyDisplayType('TOKENS')).toBe(true)
    expect(isCurrencyDisplayType('USD')).toBe(false)
    expect(isCurrencyDisplayType(123)).toBe(false)
  })

  it('parseCurrencyDisplayType falls back to CNY', () => {
    expect(parseCurrencyDisplayType('TOKENS')).toBe('TOKENS')
    expect(parseCurrencyDisplayType('unknown')).toBe('CNY')
    expect(parseCurrencyDisplayType(undefined)).toBe('CNY')
  })
})
