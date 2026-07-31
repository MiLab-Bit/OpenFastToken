import { describe, it, expect } from 'vitest'
import {
  formatCurrency,
  getDiscountLabel,
  calculatePresetPricing,
  formatCreemPrice,
  formatQuotaShort,
} from './format'

describe('formatCurrency', () => {
  it('formats finite numbers', () => {
    expect(formatCurrency(1234.5)).toContain('1,234')
    expect(formatCurrency('99.9')).toContain('99.9')
  })
  it('returns dash for non-finite input', () => {
    expect(formatCurrency('abc')).toBe('-')
    expect(formatCurrency(NaN)).toBe('-')
  })
})

describe('getDiscountLabel', () => {
  it('returns empty for zero or default-rate discount', () => {
    expect(getDiscountLabel(0)).toBe('')
    expect(getDiscountLabel(1)).toBe('')
  })
  it('shows OFF for discounts below 1', () => {
    expect(getDiscountLabel(0.5)).toBe('50% OFF')
    expect(getDiscountLabel(0.8)).toBe('20% OFF')
  })
  it('shows bonus for discounts above 1', () => {
    expect(getDiscountLabel(1.2)).toBe('+20%')
  })
})

describe('calculatePresetPricing', () => {
  it('no-bonus: pay face value, no extra credit', () => {
    const r = calculatePresetPricing(100, 0.8, 0)
    expect(r.originalPrice).toBe(80)
    expect(r.actualPrice).toBe(80)
    expect(r.bonusCredit).toBe(80)
    expect(r.isBonus).toBe(false)
  })
  it('bonus: extra credit accumulates into到账额', () => {
    const r = calculatePresetPricing(100, 0.8, 20)
    expect(r.bonusCredit).toBe(100)
    expect(r.isBonus).toBe(true)
    expect(r.displayValue).toBe(100)
  })
})

describe('formatCreemPrice', () => {
  it('formats with currency style', () => {
    expect(formatCreemPrice(9.99, 'USD')).toContain('9.99')
  })
  it('returns dash for non-finite input', () => {
    expect(formatCreemPrice(NaN, 'USD')).toBe('-')
  })
})

describe('formatQuotaShort', () => {
  it('abbreviates large numbers', () => {
    expect(formatQuotaShort(1500)).toBe('1.5K')
    expect(formatQuotaShort(2_000_000)).toBe('2.0M')
  })
  it('leaves small numbers as-is', () => {
    expect(formatQuotaShort(500)).toBe('500')
  })
  it('returns dash for non-finite input', () => {
    expect(formatQuotaShort(NaN)).toBe('-')
  })
})
