/*
 * Unit tests for lib/format.ts — number / quota / timestamp / color formatting.
 */
import { describe, it, expect } from 'vitest'
import {
  formatNumber,
  formatCompactNumber,
  formatPercent,
  formatCurrencyUSD,
  formatQuota,
  parseQuotaFromDollars,
  quotaUnitsToDollars,
  formatTimestamp,
  formatTimestampToDate,
  formatDateTimeStr,
  formatDateStr,
  formatTimeStr,
  formatLogQuota,
  formatTokens,
  formatUseTime,
  formatTimestampForInput,
  parseTimestampFromInput,
  stringToColor,
  formatTimestampRelative,
} from './format'

describe('number formatting', () => {
  it('returns "-" for null / undefined / NaN', () => {
    expect(formatNumber(null)).toBe('-')
    expect(formatNumber(undefined)).toBe('-')
    expect(formatNumber(Number.NaN)).toBe('-')
    expect(formatCompactNumber(null)).toBe('-')
    expect(formatPercent(undefined)).toBe('-')
  })
  it('formats a regular number', () => {
    expect(formatNumber(1234567)).toMatch(/1/)
    expect(formatCompactNumber(1500)).toMatch(/[0-9]/)
  })
  it('formats percent (value is already in percent units)', () => {
    expect(formatPercent(50)).toContain('%')
  })
})

describe('currency / quota formatting', () => {
  it('formats USD via CNY display', () => {
    expect(formatCurrencyUSD(720000)).toContain('¥')
    expect(formatCurrencyUSD(null)).toBe('-')
  })
  it('formatQuota yields a currency string', () => {
    const s = formatQuota(720000)
    expect(s).toContain('¥')
  })
  it('round-trips dollars <-> quota units (CNY, rate 1)', () => {
    expect(parseQuotaFromDollars(1)).toBe(500000)
    expect(quotaUnitsToDollars(500000)).toBe(1)
  })
  it('parseQuotaFromDollars guards non-finite input', () => {
    expect(parseQuotaFromDollars(Number.NaN)).toBe(0)
  })
  it('formatLogQuota yields a currency string', () => {
    expect(formatLogQuota(720000)).toContain('¥')
  })
})

describe('timestamp formatting', () => {
  it('formatTimestamp(-1) returns a (translated) "never" string', () => {
    expect(typeof formatTimestamp(-1)).toBe('string')
    expect(formatTimestamp(-1).length).toBeGreaterThan(0)
  })
  it('formatTimestampToDate treats 0/-1/undefined as "-"', () => {
    expect(formatTimestampToDate(0)).toBe('-')
    expect(formatTimestampToDate(-1)).toBe('-')
    expect(formatTimestampToDate(undefined)).toBe('-')
  })
  it('formatTimestampToDate formats seconds', () => {
    expect(formatTimestampToDate(1_700_000_000)).toMatch(
      /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/
    )
  })
  it('formatDateTimeStr / formatDateStr / formatTimeStr', () => {
    const d = new Date(0)
    expect(formatDateTimeStr(d)).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
    expect(formatDateStr(d)).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    expect(formatTimeStr(d)).toMatch(/^\d{2}:\d{2}:\d{2}$/)
  })
})

describe('token / use-time formatting', () => {
  it('formatTokens uses K/M suffixes', () => {
    expect(formatTokens(0)).toBe('-')
    expect(formatTokens(500)).toBe('500')
    expect(formatTokens(1500)).toBe('1.5K')
    expect(formatTokens(2_000_000)).toBe('2.00M')
  })
  it('formatUseTime picks units', () => {
    expect(formatUseTime(30)).toBe('30.0s')
    expect(formatUseTime(90)).toBe('1m 30s')
  })
})

describe('input <-> timestamp', () => {
  it('formatTimestampForInput(-1) is empty', () => {
    expect(formatTimestampForInput(-1)).toBe('')
  })
  it('parseTimestampFromInput("") is -1', () => {
    expect(parseTimestampFromInput('')).toBe(-1)
  })
  it('round-trips a datetime-local value', () => {
    const value = '2020-01-01T00:00'
    const parsed = parseTimestampFromInput(value)
    expect(parsed).toBe(Math.floor(new Date(value).getTime() / 1000))
  })
})

describe('stringToColor (hsl)', () => {
  it('returns gray for empty string', () => {
    expect(stringToColor('')).toBe('gray')
  })
  it('returns an hsl color', () => {
    expect(stringToColor('gpt-4')).toMatch(/^hsl\(/)
  })
})

describe('formatTimestampRelative', () => {
  it('returns "-" for zero', () => {
    expect(formatTimestampRelative(0)).toBe('-')
  })
  it('returns a non-empty relative string for an old timestamp', () => {
    const old = Math.floor(new Date('2000-01-01').getTime() / 1000)
    expect(formatTimestampRelative(old).length).toBeGreaterThan(0)
  })
})
