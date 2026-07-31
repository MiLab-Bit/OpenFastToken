import { describe, it, expect, vi, beforeEach } from 'vitest'
import { formatRelativeTime, formatTimestamp } from './channel-utils'
import dayjs from '@/lib/dayjs'
import { formatTimestampToDate } from '@/lib/format'

vi.mock('@/lib/dayjs', () => ({
  default: vi.fn((ts: number) => ({ fromNow: () => 'some time ago' })),
}))

vi.mock('@/lib/format', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/format')>()
  return {
    ...actual,
    formatTimestampToDate: vi.fn((ts: number) => '2020-01-01'),
  }
})

beforeEach(() => {
  vi.mocked(dayjs).mockImplementation((ts: number) => ({
    fromNow: () => 'some time ago',
  }))
  vi.mocked(formatTimestampToDate).mockImplementation((ts: number) => '2020-01-01')
})

describe('formatRelativeTime — all branches', () => {
  it('returns "Never" for a zero / falsy timestamp', () => {
    expect(formatRelativeTime(0)).toBe('Never')
    expect(formatRelativeTime(undefined as never)).toBe('Never')
  })

  it('returns a relative string for a valid timestamp', () => {
    expect(formatRelativeTime(1_000_000)).toBe('some time ago')
  })

  it('returns "Unknown" when dayjs throws', () => {
    vi.mocked(dayjs).mockImplementation(() => {
      throw new Error('boom')
    })
    expect(formatRelativeTime(1_000_000)).toBe('Unknown')
  })
})

describe('formatTimestamp — all branches', () => {
  it('returns "N/A" for a zero / falsy timestamp', () => {
    expect(formatTimestamp(0)).toBe('N/A')
    expect(formatTimestamp(undefined as never)).toBe('N/A')
  })

  it('returns a formatted date for a valid timestamp', () => {
    expect(formatTimestamp(1_000_000)).toBe('2020-01-01')
  })

  it('returns "Invalid date" when formatTimestampToDate throws', () => {
    vi.mocked(formatTimestampToDate).mockImplementation(() => {
      throw new Error('boom')
    })
    expect(formatTimestamp(1_000_000)).toBe('Invalid date')
  })
})
