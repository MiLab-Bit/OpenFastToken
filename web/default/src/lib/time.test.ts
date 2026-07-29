/*
 * Unit tests for lib/time.ts — date/time helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  dateToUnixTimestamp,
  toStartOfDay,
  getStartOfDay,
  getEndOfDay,
  getNormalizedDateRange,
  getRollingDateRange,
  computeTimeRange,
  formatDate,
  formatDateTimeObject,
  formatChartTime,
  addTimeToDate,
} from './time'

describe('dateToUnixTimestamp', () => {
  it('converts epoch Date to 0', () => {
    expect(dateToUnixTimestamp(new Date(0))).toBe(0)
  })
})

describe('toStartOfDay', () => {
  it('keeps the same calendar day', () => {
    const ts = 1_700_000_000
    const start = toStartOfDay(ts)
    expect(formatDate(start)).toBe(formatDate(ts))
    expect(start).toBeLessThanOrEqual(ts)
  })
})

describe('getStartOfDay / getEndOfDay', () => {
  it('zeroes the time for start of day', () => {
    const d = getStartOfDay(new Date('2023-06-15T13:45:30'))
    expect(d.getHours()).toBe(0)
    expect(d.getMinutes()).toBe(0)
    expect(d.getSeconds()).toBe(0)
    expect(d.getMilliseconds()).toBe(0)
  })
  it('sets end of day to 23:59:59.999', () => {
    const d = getEndOfDay(new Date('2023-06-15T01:02:03'))
    expect(d.getHours()).toBe(23)
    expect(d.getMinutes()).toBe(59)
    expect(d.getSeconds()).toBe(59)
    expect(d.getMilliseconds()).toBe(999)
  })
})

describe('getNormalizedDateRange', () => {
  it('returns start <= end as Date objects spanning the range', () => {
    const { start, end } = getNormalizedDateRange(7, new Date('2023-06-15T12:00:00'))
    expect(start).toBeInstanceOf(Date)
    expect(end).toBeInstanceOf(Date)
    expect(end.getTime()).toBeGreaterThanOrEqual(start.getTime())
  })
})

describe('getRollingDateRange', () => {
  it('spans exactly the requested number of days', () => {
    const from = new Date('2023-06-15T12:00:00')
    const { start, end } = getRollingDateRange(3, from)
    expect(end.getTime() - start.getTime()).toBe(3 * 24 * 60 * 60 * 1000)
    expect(end.getTime()).toBe(from.getTime())
  })
})

describe('computeTimeRange', () => {
  it('without day normalization: end = start + days', () => {
    const { start_timestamp, end_timestamp } = computeTimeRange(2)
    expect(end_timestamp - start_timestamp).toBe(2 * 24 * 3600)
  })
  it('with day normalization end includes the full last day', () => {
    const now = Math.floor(Date.now() / 1000)
    const { start_timestamp, end_timestamp } = computeTimeRange(2, undefined, undefined, true)
    // end_timestamp = toStartOfDay(now) + 86400 - 1
    expect(end_timestamp - start_timestamp).toBe(2 * 24 * 3600 + 24 * 3600 - 1)
  })
  it('uses provided start/end dates in day-normalized mode', () => {
    const start = new Date('2023-06-01T00:00:00')
    const end = new Date('2023-06-10T00:00:00')
    const { start_timestamp, end_timestamp } = computeTimeRange(
      2,
      start,
      end,
      true
    )
    expect(end_timestamp).toBeGreaterThan(start_timestamp)
  })
})

describe('formatDate / formatDateTimeObject', () => {
  it('formatDate returns YYYY-MM-DD shape', () => {
    expect(formatDate(0)).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })
  it('formatDateTimeObject returns YYYY-MM-DD HH:mm:ss shape', () => {
    expect(formatDateTimeObject(new Date(0))).toMatch(
      /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/
    )
  })
})

describe('formatChartTime', () => {
  it('day granularity returns MM-DD', () => {
    const s = formatChartTime(1_700_000_000, 'day')
    expect(s).toMatch(/^\d{2}-\d{2}$/)
  })
  it('hour granularity appends :00', () => {
    expect(formatChartTime(1_700_000_000, 'hour')).toContain(':00')
  })
  it('week granularity appends a range', () => {
    expect(formatChartTime(1_700_000_000, 'week')).toContain(' - ')
  })
})

describe('addTimeToDate', () => {
  it('returns undefined when all deltas are zero', () => {
    expect(addTimeToDate(0, 0, 0, new Date('2023-06-15T12:00:00'))).toBeUndefined()
  })
  it('adds months', () => {
    const base = new Date('2023-06-15T12:00:00')
    const out = addTimeToDate(1, 0, 0, base)!
    expect(out.getMonth()).toBe(6) // July (0-indexed)
  })
  it('adds days and hours', () => {
    const base = new Date('2023-06-15T12:00:00')
    const out = addTimeToDate(0, 2, 3, base)!
    expect(out.getDate()).toBe(17)
    expect(out.getHours()).toBe(15)
  })
})
