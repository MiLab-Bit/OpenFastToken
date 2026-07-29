import { describe, it, expect } from 'vitest'
import { RESPONSE_TIME_CONFIG } from '../constants'
import {
  getChannelTypeLabel,
  getChannelTypeIcon,
  getChannelStatusBadge,
  getMultiKeyStatusBadge,
  isChannelEnabled,
  isMultiKeyChannel,
  formatChannelKey,
  formatKeyPreview,
  countKeys,
  parseModelsList,
  parseGroupsList,
  formatModelsString,
  formatGroupsString,
  parseChannelSettings,
  parseChannelOtherSettings,
  validateChannelSettings,
  formatBalance,
  getBalanceVariant,
  formatResponseTime,
  getResponseTimeConfig,
  formatRelativeTime,
  formatTimestamp,
  getPriorityDisplay,
  getWeightDisplay,
  validateChannelName,
  validateApiKey,
  validateModels,
  validateGroups,
  channelNeedsAttention,
  getAttentionReason,
  isTagAggregateRow,
  aggregateChannelsByTag,
  deduplicateKeys,
  getKeyPromptForType,
} from './channel-utils'

describe('channel type helpers', () => {
  it('labels known and unknown types', () => {
    expect(getChannelTypeLabel(1)).not.toBe('Unknown')
    expect(getChannelTypeLabel(99999)).toBe('Unknown')
  })
  it('maps type to an icon name (defaulting to OpenAI)', () => {
    expect(getChannelTypeIcon(1)).toBe('OpenAI')
    expect(getChannelTypeIcon(14)).toBe('Claude')
    expect(getChannelTypeIcon(99999)).toBe('OpenAI')
  })
  it('resolves status badges', () => {
    expect(getChannelStatusBadge(1)).toBeTruthy()
    expect(getMultiKeyStatusBadge(1)).toBeTruthy()
  })
})

describe('channel predicates', () => {
  it('isChannelEnabled', () => {
    expect(isChannelEnabled({ status: 1 } as never)).toBe(true)
    expect(isChannelEnabled({ status: 0 } as never)).toBe(false)
  })
  it('isMultiKeyChannel', () => {
    expect(
      isMultiKeyChannel({ channel_info: { is_multi_key: true } } as never)
    ).toBe(true)
    expect(isMultiKeyChannel({} as never)).toBe(false)
  })
})

describe('key formatting', () => {
  it('formatChannelKey masks single keys and counts multi keys', () => {
    expect(formatChannelKey('', false)).toBe('')
    expect(formatChannelKey('abcdefghijklmnopQR', false)).toContain('...')
    expect(formatChannelKey('a\nb\nc', true)).toBe('3 keys')
  })
  it('formatKeyPreview truncates long keys', () => {
    expect(formatKeyPreview('')).toBe('')
    expect(formatKeyPreview('abcdefghij', 5)).toBe('abcde...')
  })
  it('countKeys counts non-empty lines', () => {
    expect(countKeys('a\n\nb\nc')).toBe(3)
    expect(countKeys('')).toBe(0)
  })
})

describe('model/group parsing', () => {
  it('parseModelsList splits and trims', () => {
    expect(parseModelsList('a, b , c')).toEqual(['a', 'b', 'c'])
    expect(parseModelsList('')).toEqual([])
  })
  it('parseGroupsList puts default first', () => {
    expect(parseGroupsList('b,default,a')).toEqual(['default', 'a', 'b'])
  })
  it('format helpers join back', () => {
    expect(formatModelsString(['a', 'b'])).toBe('a,b')
    expect(formatGroupsString(['x', 'y'])).toBe('x,y')
  })
})

describe('settings parsing', () => {
  it('parseChannelSettings', () => {
    expect(parseChannelSettings('{"a":1}')).toEqual({ a: 1 })
    expect(parseChannelSettings('bad')).toEqual({})
    expect(parseChannelSettings(null)).toEqual({})
  })
  it('parseChannelOtherSettings', () => {
    expect(parseChannelOtherSettings('{"a":1}')).toEqual({ a: 1 })
    expect(parseChannelOtherSettings('{}')).toEqual({})
    expect(parseChannelOtherSettings('bad')).toEqual({})
  })
  it('validateChannelSettings', () => {
    expect(validateChannelSettings('{"a":1}')).toBe(true)
    expect(validateChannelSettings('bad')).toBe(false)
    expect(validateChannelSettings('')).toBe(true)
  })
})

describe('balance', () => {
  it('formatBalance returns dash for non-finite', () => {
    expect(formatBalance(null)).toBe('-')
    expect(formatBalance(NaN)).toBe('-')
    // 5 is finite so it is currency-formatted (not the dash sentinel).
    expect(formatBalance(5)).not.toBe('-')
  })
  it('getBalanceVariant thresholds', () => {
    expect(getBalanceVariant(0)).toBe('neutral')
    expect(getBalanceVariant(0.5)).toBe('danger')
    expect(getBalanceVariant(5)).toBe('warning')
    expect(getBalanceVariant(50)).toBe('success')
  })
})

describe('response time', () => {
  it('formatResponseTime', () => {
    expect(formatResponseTime(0)).toBe('Not tested')
    expect(formatResponseTime(500)).toBe('500ms')
    expect(formatResponseTime(1500)).toBe('1.50s')
  })
  it('getResponseTimeConfig ranks by threshold', () => {
    expect(getResponseTimeConfig(0)).toBe(RESPONSE_TIME_CONFIG.UNKNOWN)
    expect(getResponseTimeConfig(100)).toBe(RESPONSE_TIME_CONFIG.EXCELLENT)
    expect(getResponseTimeConfig(100000)).toBe(RESPONSE_TIME_CONFIG.POOR)
  })
})

describe('time formatting', () => {
  it('formatRelativeTime', () => {
    expect(formatRelativeTime(0)).toBe('Never')
    expect(typeof formatRelativeTime(1000)).toBe('string')
  })
  it('formatTimestamp', () => {
    expect(formatTimestamp(0)).toBe('N/A')
    expect(typeof formatTimestamp(1700000000)).toBe('string')
  })
})

describe('priority/weight display', () => {
  it('default to 0 string', () => {
    expect(getPriorityDisplay(null)).toBe('0')
    expect(getWeightDisplay(undefined)).toBe('0')
    expect(getPriorityDisplay(5)).toBe('5')
  })
})

describe('validation', () => {
  it('validateChannelName / validateApiKey', () => {
    expect(validateChannelName('')).toBe(false)
    expect(validateChannelName('x')).toBe(true)
    expect(validateApiKey('')).toBe(false)
    expect(validateApiKey('k')).toBe(true)
  })
  it('validateModels / validateGroups', () => {
    expect(validateModels('a,b')).toBe(true)
    expect(validateModels('')).toBe(false)
    expect(validateGroups('g')).toBe(true)
    expect(validateGroups('')).toBe(false)
  })
})

describe('attention', () => {
  it('channelNeedsAttention', () => {
    expect(channelNeedsAttention({ status: 3 } as never)).toBe(true)
    expect(
      channelNeedsAttention({ status: 1, balance: 0.5 } as never)
    ).toBe(true)
    expect(channelNeedsAttention({ status: 1, balance: 50 } as never)).toBe(false)
  })
  it('getAttentionReason', () => {
    expect(getAttentionReason({ status: 3 } as never)).toBe('Auto-disabled')
    expect(
      getAttentionReason({ status: 1, balance: 0.5 } as never)
    ).toBe('Low balance')
  })
})

describe('tag aggregation', () => {
  it('isTagAggregateRow', () => {
    expect(isTagAggregateRow({ children: [] } as never)).toBe(true)
    expect(isTagAggregateRow({} as never)).toBe(false)
  })
  it('aggregateChannelsByTag groups and sums quotas', () => {
    const rows = aggregateChannelsByTag([
      { id: 1, tag: 'g', status: 1, used_quota: 10, response_time: 100, priority: 5, weight: 5, balance: 0, group: 'a', channel_info: {} } as never,
      { id: 2, tag: 'g', status: 1, used_quota: 20, response_time: 200, priority: 5, weight: 5, balance: 0, group: 'b', channel_info: {} } as never,
    ])
    expect(rows).toHaveLength(1)
    const tagRow = rows[0] as { children: unknown[]; used_quota: number; status: number }
    expect(tagRow.children).toHaveLength(2)
    expect(tagRow.used_quota).toBe(30)
    expect(tagRow.status).toBe(1)
  })
})

describe('key deduplication', () => {
  it('removes duplicate lines while keeping order', () => {
    const res = deduplicateKeys('a\nb\na\nc')
    expect(res.deduplicatedText).toBe('a\nb\nc')
    expect(res.beforeCount).toBe(4)
    expect(res.afterCount).toBe(3)
    expect(res.removedCount).toBe(1)
  })
  it('handles empty input', () => {
    expect(deduplicateKeys('')).toEqual({
      deduplicatedText: '',
      beforeCount: 0,
      afterCount: 0,
      removedCount: 0,
    })
  })
})

describe('getKeyPromptForType', () => {
  it('returns a non-empty prompt (defaulting when unknown)', () => {
    expect(getKeyPromptForType(1).length).toBeGreaterThan(0)
    expect(getKeyPromptForType(99999).length).toBeGreaterThan(0)
  })
})
