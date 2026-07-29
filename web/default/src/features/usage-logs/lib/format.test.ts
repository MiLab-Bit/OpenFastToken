import { describe, it, expect } from 'vitest'
import {
  getParamOverrideActionLabel,
  parseAuditLine,
  isViolationFeeLog,
  parseLogOther,
  getTimeColor,
  getFirstResponseTimeColor,
  getThroughputColor,
  getResponseTimeColor,
  formatModelName,
  decodeBillingExprB64,
  resolveMatchedTier,
  hasAnyCacheTokens,
  getTieredBillingSummary,
  formatDuration,
} from './format'
import { parseTiersFromExpr } from '@/features/pricing/lib/billing-expr'

const t = (k: string) => k

describe('getParamOverrideActionLabel', () => {
  it('maps a known lowercased action to its label', () => {
    expect(getParamOverrideActionLabel('SET', t)).toBe('Set')
    expect(getParamOverrideActionLabel('trim_prefix', t)).toBe('Trim Prefix')
  })
  it('returns the raw action for unknown actions', () => {
    expect(getParamOverrideActionLabel('unknown-op', t)).toBe('unknown-op')
  })
})

describe('parseAuditLine', () => {
  it('splits action and content on the first space', () => {
    expect(parseAuditLine('SET foo bar')).toEqual({
      action: 'SET',
      content: 'foo bar',
    })
  })
  it('returns the whole line when there is no space', () => {
    expect(parseAuditLine('NOBREAK')).toEqual({
      action: 'NOBREAK',
      content: 'NOBREAK',
    })
  })
  it('returns null for non-string input', () => {
    expect(parseAuditLine(undefined as never)).toBeNull()
  })
})

describe('isViolationFeeLog', () => {
  it('detects violation fee markers', () => {
    expect(isViolationFeeLog({ violation_fee: true })).toBe(true)
    expect(isViolationFeeLog({ violation_fee_code: 'X' })).toBe(true)
    expect(isViolationFeeLog({ violation_fee_marker: 1 })).toBe(true)
  })
  it('returns false when none present', () => {
    expect(isViolationFeeLog(null)).toBe(false)
    expect(isViolationFeeLog({})).toBe(false)
  })
})

describe('parseLogOther', () => {
  it('parses valid JSON', () => {
    expect(parseLogOther('{"a":1}')).toEqual({ a: 1 })
  })
  it('returns null for empty or invalid JSON', () => {
    expect(parseLogOther(undefined)).toBeNull()
    expect(parseLogOther('not json')).toBeNull()
  })
})

describe('time/throughput colors', () => {
  it('getTimeColor thresholds', () => {
    expect(getTimeColor(5)).toBe('success')
    expect(getTimeColor(20)).toBe('warning')
    expect(getTimeColor(60)).toBe('danger')
  })
  it('getFirstResponseTimeColor thresholds', () => {
    expect(getFirstResponseTimeColor(4)).toBe('success')
    expect(getFirstResponseTimeColor(8)).toBe('warning')
    expect(getFirstResponseTimeColor(20)).toBe('danger')
  })
  it('getThroughputColor thresholds', () => {
    expect(getThroughputColor(40)).toBe('success')
    expect(getThroughputColor(20)).toBe('warning')
    expect(getThroughputColor(5)).toBe('danger')
  })
  it('getResponseTimeColor falls back to time color with few tokens', () => {
    expect(getResponseTimeColor(5, 10)).toBe('success')
    expect(getResponseTimeColor(100, 4000)).toBe('success') // 4000/100=40 tps
  })
})

describe('formatModelName', () => {
  it('reports mapping when upstream model present', () => {
    const res = formatModelName({
      model_name: 'gpt-4',
      other: JSON.stringify({ is_model_mapped: true, upstream_model_name: 'gpt-4o' }),
    })
    expect(res).toEqual({
      name: 'gpt-4',
      isMapped: true,
      actualModel: 'gpt-4o',
    })
  })
  it('reports not mapped otherwise', () => {
    const res = formatModelName({ model_name: 'gpt-4', other: '{}' })
    expect(res.isMapped).toBe(false)
    expect(res.actualModel).toBeUndefined()
  })
})

describe('decodeBillingExprB64', () => {
  it('decodes a base64 billing expression', () => {
    const b64 = btoa('p*10 + c*20')
    expect(decodeBillingExprB64(b64)).toBe('p*10 + c*20')
  })
  it('returns empty for missing/malformed input', () => {
    expect(decodeBillingExprB64(undefined)).toBe('')
    expect(decodeBillingExprB64('!!!not base64')).toBe('')
  })
})

describe('resolveMatchedTier', () => {
  const tiers = parseTiersFromExpr('tier("0-100", p*10)')
  it('matches by normalized label', () => {
    expect(resolveMatchedTier(tiers, '0-100')).toBe(tiers[0])
  })
  it('returns null when no match', () => {
    expect(resolveMatchedTier(tiers, 'missing')).toBeNull()
    expect(resolveMatchedTier([], 'x')).toBeNull()
    expect(resolveMatchedTier(tiers, undefined)).toBeNull()
  })
})

describe('hasAnyCacheTokens', () => {
  it('detects cache token fields', () => {
    expect(hasAnyCacheTokens({ cache_tokens: 5 })).toBe(true)
    expect(hasAnyCacheTokens({ cache_creation_tokens: 5 })).toBe(true)
  })
  it('false when absent', () => {
    expect(hasAnyCacheTokens(null)).toBe(false)
    expect(hasAnyCacheTokens({})).toBe(false)
  })
})

describe('getTieredBillingSummary', () => {
  it('returns a summary for a valid tiered log', () => {
    const other = {
      billing_mode: 'tiered_expr',
      expr_b64: btoa('tier("x", p*10)'),
      matched_tier: 'x',
    }
    const summary = getTieredBillingSummary(other)
    expect(summary).not.toBeNull()
    expect(summary?.tier.inputPrice).toBe(10)
    expect(summary?.priceEntries[0].shortLabel).toBe('Input')
  })
  it('returns null for non-tiered or unparseable logs', () => {
    expect(getTieredBillingSummary(null)).toBeNull()
    expect(getTieredBillingSummary({ billing_mode: 'flat' })).toBeNull()
    expect(
      getTieredBillingSummary({ billing_mode: 'tiered_expr', expr_b64: 'bad' })
    ).toBeNull()
  })
})

describe('formatDuration', () => {
  it('returns null without both timestamps', () => {
    expect(formatDuration(undefined, 100)).toBeNull()
  })
  it('classifies duration color (ms)', () => {
    expect(formatDuration(1000, 2000)?.variant).toBe('green')
    expect(formatDuration(1000, 70000)?.variant).toBe('red')
  })
  it('classifies duration color (seconds)', () => {
    expect(formatDuration(1, 2, 'seconds')?.variant).toBe('green')
  })
})
