import { describe, it, expect, vi } from 'vitest'
import {
  formatQuota,
  getChannelStatusBadge,
  getMultiKeyStatusBadge,
  formatChannelKey,
  formatKeyPreview,
  formatResponseTime,
  getResponseTimeConfig,
  channelNeedsAttention,
  getAttentionReason,
  aggregateChannelsByTag,
  deduplicateKeys,
  getWeightDisplay,
} from './channel-utils'
import { RESPONSE_TIME_CONFIG, RESPONSE_TIME_THRESHOLDS, CHANNEL_STATUS_CONFIG, MULTI_KEY_STATUS_CONFIG } from '../constants'

describe('quota formatting', () => {
  it('formats a quota as a non-empty currency string', () => {
    const out = formatQuota(1000)
    expect(typeof out).toBe('string')
    expect(out.length).toBeGreaterThan(0)
  })
  it('formats a zero quota without throwing', () => {
    expect(typeof formatQuota(0)).toBe('string')
  })
})

describe('status badge fallbacks', () => {
  it('falls back to status 0 config for unknown channel status', () => {
    expect(getChannelStatusBadge(99999)).toBe(CHANNEL_STATUS_CONFIG[0])
  })
  it('falls back to multi-key config 1 for unknown status', () => {
    expect(getMultiKeyStatusBadge(99999)).toBe(MULTI_KEY_STATUS_CONFIG[1])
  })
})

describe('key formatting — branches', () => {
  it('masks short single keys (<=16 chars)', () => {
    const out = formatChannelKey('abcdefghijklmnop', false)
    expect(out).toBe('abcd...mnop')
  })
  it('returns short keys unchanged when within maxLength', () => {
    expect(formatKeyPreview('abc', 10)).toBe('abc')
    expect(formatKeyPreview('short')).toBe('short')
  })
})

describe('formatResponseTime — i18n t function', () => {
  const t = vi.fn((key: string, opts?: { value?: number | string }) =>
    opts && opts.value !== undefined ? `${key}:${opts.value}` : key
  )
  it('delegates "Not tested"', () => {
    expect(formatResponseTime(0, t)).toBe('Not tested')
  })
  it('delegates milliseconds', () => {
    expect(formatResponseTime(500, t)).toBe('{{value}}ms:500')
  })
  it('delegates seconds', () => {
    expect(formatResponseTime(1500, t)).toBe('{{value}}s:1.50')
  })
})

describe('getResponseTimeConfig — GOOD and FAIR tiers', () => {
  const { EXCELLENT, GOOD, FAIR, POOR } = RESPONSE_TIME_THRESHOLDS
  it('ranks a value between EXCELLENT and GOOD as GOOD', () => {
    expect(getResponseTimeConfig(Math.floor((EXCELLENT + GOOD) / 2))).toBe(
      RESPONSE_TIME_CONFIG.GOOD
    )
  })
  it('ranks a value between GOOD and FAIR as FAIR', () => {
    expect(getResponseTimeConfig(Math.floor((GOOD + FAIR) / 2))).toBe(
      RESPONSE_TIME_CONFIG.FAIR
    )
  })
  it('returns POOR at and beyond the POOR threshold', () => {
    expect(getResponseTimeConfig(POOR)).toBe(RESPONSE_TIME_CONFIG.POOR)
    expect(getResponseTimeConfig(POOR + 1000)).toBe(RESPONSE_TIME_CONFIG.POOR)
  })
})

describe('channel attention — multi-key all-disabled', () => {
  const multiKeyAllDisabled = {
    status: 1,
    balance: 50,
    channel_info: {
      is_multi_key: true,
      multi_key_size: 2,
      multi_key_status_list: { '0': 0, '1': 0 },
    },
  } as never

  it('needs attention when every key is disabled', () => {
    expect(channelNeedsAttention(multiKeyAllDisabled)).toBe(true)
  })
  it('reports the all-keys-disabled reason', () => {
    expect(getAttentionReason(multiKeyAllDisabled)).toBe('All keys disabled')
  })
})

describe('aggregateChannelsByTag — priority/weight null & multi-tag', () => {
  it('nulls priority when children differ', () => {
    const rows = aggregateChannelsByTag([
      { id: 1, tag: 'g', status: 1, used_quota: 10, response_time: 100, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
      { id: 2, tag: 'g', status: 1, used_quota: 20, response_time: 200, priority: 7, weight: 5, balance: 0, group: 'b', channel_info: {} } as never,
    ])
    const tagRow = rows[0] as { priority: number | null; weight: number | null }
    expect(tagRow.priority).toBeNull()
    expect(tagRow.weight).toBe(5)
  })

  it('nulls weight when children differ', () => {
    const rows = aggregateChannelsByTag([
      { id: 1, tag: 'g', status: 1, used_quota: 10, response_time: 100, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
      { id: 2, tag: 'g', status: 1, used_quota: 20, response_time: 200, priority: 5, weight: 8, balance: 0, group: 'b', channel_info: {} } as never,
    ])
    const tagRow = rows[0] as { weight: number | null }
    expect(tagRow.weight).toBeNull()
  })

  it('sets status from a non-enabled first child via the undefined branch', () => {
    const rows = aggregateChannelsByTag([
      { id: 1, tag: 'g', status: 2, used_quota: 10, response_time: 100, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
    ])
    expect((rows[0] as { status: number }).status).toBe(2)
  })

  it('produces one row per distinct tag', () => {
    const rows = aggregateChannelsByTag([
      { id: 1, tag: 'g', status: 1, used_quota: 10, response_time: 100, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
      { id: 2, tag: 'h', status: 1, used_quota: 20, response_time: 200, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
    ])
    expect(rows).toHaveLength(2)
  })

  it('does not duplicate an already-present group', () => {
    const rows = aggregateChannelsByTag([
      { id: 1, tag: 'g', status: 1, used_quota: 10, response_time: 100, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
      { id: 2, tag: 'g', status: 1, used_quota: 20, response_time: 200, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
    ])
    expect((rows[0] as { group: string }).group).toBe('a')
  })
})

describe('deduplicateKeys — whitespace-only lines', () => {
  it('skips blank lines but still counts them', () => {
    const res = deduplicateKeys('a\n   \nb\n')
    expect(res.deduplicatedText).toBe('a\nb')
    expect(res.beforeCount).toBe(4)
    expect(res.afterCount).toBe(2)
    expect(res.removedCount).toBe(2)
  })
})

describe('weight display — null handling', () => {
  it('returns "0" for null/undefined weight', () => {
    expect(getWeightDisplay(null)).toBe('0')
    expect(getWeightDisplay(undefined)).toBe('0')
  })
  it('returns the stringified weight number', () => {
    expect(getWeightDisplay(5)).toBe('5')
    expect(getWeightDisplay(0)).toBe('0')
  })
})

describe('getAttentionReason — all branches', () => {
  it('returns "Auto-disabled" when status is 3', () => {
    expect(
      getAttentionReason({ status: 3, balance: 50, channel_info: {} } as never)
    ).toBe('Auto-disabled')
  })
  it('returns "Low balance" when 0 < balance < 1', () => {
    expect(
      getAttentionReason({ status: 1, balance: 0.5, channel_info: {} } as never)
    ).toBe('Low balance')
  })
  it('returns null when the channel needs no attention', () => {
    expect(
      getAttentionReason({ status: 1, balance: 0, channel_info: {} } as never)
    ).toBeNull()
    expect(
      getAttentionReason({ status: 1, balance: 50, channel_info: {} } as never)
    ).toBeNull()
  })
})

describe('aggregateChannelsByTag — channel without a tag', () => {
  it('falls back to an empty-string tag when channel.tag is missing', () => {
    const rows = aggregateChannelsByTag([
      {
        id: 1,
        status: 1,
        used_quota: 10,
        response_time: 100,
        priority: 5,
        weight: 5,
        balance: 0,
        group: 'a',
        channel_info: {},
      } as never,
    ])
    expect(rows).toHaveLength(1)
    expect((rows[0] as { key: string }).key).toBe('')
  })
})
