import { describe, it, expect } from 'vitest'
import {
  parseTiersFromExpr,
  normalizeTierLabel,
  tryParseRequestRuleExpr,
  splitBillingExprAndRequestRules,
  combineBillingExpr,
  createEmptyCondition,
  createEmptyTimeCondition,
  createEmptyRuleGroup,
  createEmptyTimeRuleGroup,
  getRequestRuleMatchOptions,
  normalizeCondition,
  buildRequestRuleExpr,
  SOURCE_TIME,
  SOURCE_PARAM,
  SOURCE_HEADER,
  MATCH_EQ,
  MATCH_CONTAINS,
  MATCH_EXISTS,
  MATCH_RANGE,
} from './billing-expr'

describe('parseTiersFromExpr', () => {
  it('returns [] for empty input', () => {
    expect(parseTiersFromExpr('')).toEqual([])
  })
  it('parses a simple tier with coefficient mapping', () => {
    const tiers = parseTiersFromExpr('tier("0-100", p*10 + c*20)')
    expect(tiers).toHaveLength(1)
    expect(tiers[0].label).toBe('0-100')
    expect(tiers[0].inputPrice).toBe(10)
    expect(tiers[0].outputPrice).toBe(20)
  })
  it('parses a versioned expr', () => {
    const tiers = parseTiersFromExpr('v2:tier("x", p*5)')
    expect(tiers[0].label).toBe('x')
    expect(tiers[0].inputPrice).toBe(5)
  })
  it('parses tier conditions', () => {
    const tiers = parseTiersFromExpr(
      'p<100 && c<=50 ? tier("small", p*1) : tier("big", p*2)'
    )
    expect(tiers).toHaveLength(2)
    expect(tiers[0].conditions).toEqual([
      { var: 'p', op: '<', value: 100 },
      { var: 'c', op: '<=', value: 50 },
    ])
    expect(tiers[1].label).toBe('big')
    expect(tiers[1].inputPrice).toBe(2)
  })
  it('captures extra pricing vars', () => {
    const tiers = parseTiersFromExpr('tier("x", cr*3 + img*4)')
    expect(tiers[0].cacheReadPrice).toBe(3)
    expect(tiers[0].imagePrice).toBe(4)
  })
})

describe('normalizeTierLabel', () => {
  it('normalizes comparison glyphs to ascii', () => {
    expect(normalizeTierLabel('≤100')).toBe('<100')
    expect(normalizeTierLabel('≥50')).toBe('>50')
    expect(normalizeTierLabel('＜100')).toBe('<100')
  })
  it('lowercases and collapses whitespace', () => {
    expect(normalizeTierLabel('A  B')).toBe('ab')
  })
  it('returns empty for undefined', () => {
    expect(normalizeTierLabel(undefined)).toBe('')
  })
})

describe('tryParseRequestRuleExpr', () => {
  it('returns [] for empty input', () => {
    expect(tryParseRequestRuleExpr('')).toEqual([])
  })
  it('parses a single exists condition rule', () => {
    const groups = tryParseRequestRuleExpr('(param("model") != nil ? 1.5 : 1)')
    expect(groups).toHaveLength(1)
    expect(groups[0].conditions[0]).toEqual({
      source: 'param',
      path: 'model',
      mode: MATCH_EXISTS,
      value: '',
    })
    expect(groups[0].multiplier).toBe('1.5')
  })
  it('parses a header contains rule', () => {
    const groups = tryParseRequestRuleExpr(
      '(has(header("x-app"), "abc") ? 2 : 1)'
    )
    expect(groups[0].conditions[0]).toEqual({
      source: 'header',
      path: 'x-app',
      mode: MATCH_CONTAINS,
      value: 'abc',
    })
  })
  it('returns null for unparseable expression', () => {
    expect(tryParseRequestRuleExpr('garbage')).toBeNull()
  })
})

describe('splitBillingExprAndRequestRules', () => {
  it('returns empty when input is empty', () => {
    expect(splitBillingExprAndRequestRules('')).toEqual({
      billingExpr: '',
      requestRuleExpr: '',
    })
  })
  it('returns the whole expr when there is no rule factor', () => {
    expect(splitBillingExprAndRequestRules('p*10 + c*20')).toEqual({
      billingExpr: 'p*10 + c*20',
      requestRuleExpr: '',
    })
  })
  it('splits base and rule parts', () => {
    const res = splitBillingExprAndRequestRules(
      '(p*10 + c*20) * (param("m") != nil ? 1.5 : 1)'
    )
    expect(res.billingExpr).toBe('p*10 + c*20')
    expect(res.requestRuleExpr).toBe('(param("m") != nil ? 1.5 : 1)')
  })
})

describe('combineBillingExpr', () => {
  it('returns empty when base is empty', () => {
    expect(combineBillingExpr('', 'x')).toBe('')
  })
  it('returns base when rules are empty', () => {
    expect(combineBillingExpr('p*10', '')).toBe('p*10')
  })
  it('combines base and rules', () => {
    expect(combineBillingExpr('p*10', '(param("m") != nil ? 1.5 : 1)')).toBe(
      '(p*10) * (param("m") != nil ? 1.5 : 1)'
    )
  })
})

describe('createEmpty helpers', () => {
  it('creates empty param condition', () => {
    expect(createEmptyCondition()).toEqual({
      source: 'param',
      path: '',
      mode: MATCH_EQ,
      value: '',
    })
  })
  it('creates empty time condition', () => {
    const c = createEmptyTimeCondition()
    expect(c.source).toBe('time')
    expect(c.timeFunc).toBe('hour')
    expect(c.timezone).toBe('Asia/Shanghai')
  })
  it('creates empty rule groups', () => {
    expect(createEmptyRuleGroup().conditions).toHaveLength(1)
    expect(createEmptyTimeRuleGroup().conditions[0].source).toBe('time')
  })
})

describe('getRequestRuleMatchOptions', () => {
  it('offers range only for time source', () => {
    const opts = getRequestRuleMatchOptions(SOURCE_TIME)
    expect(opts.map((o) => o.value)).toContain(MATCH_RANGE)
    expect(opts.map((o) => o.value)).not.toContain(MATCH_CONTAINS)
  })
  it('offers base options for header', () => {
    const opts = getRequestRuleMatchOptions(SOURCE_HEADER)
    expect(opts.map((o) => o.value)).toEqual([
      MATCH_EQ,
      MATCH_CONTAINS,
      MATCH_EXISTS,
    ])
  })
  it('offers numeric comparators for param', () => {
    const opts = getRequestRuleMatchOptions(SOURCE_PARAM)
    expect(opts.map((o) => o.value)).toContain(MATCH_CONTAINS)
    expect(opts.map((o) => o.value)).toContain('gt')
  })
})

describe('normalizeCondition', () => {
  it('falls back to param for unknown source', () => {
    const c = normalizeCondition({ source: 'weird' as never, path: 'p' } as never)
    expect(c.source).toBe('param')
  })
  it('normalizes a time condition', () => {
    const c = normalizeCondition({
      source: 'time',
      timeFunc: 'hour',
      timezone: 'UTC',
      mode: 'range',
      value: '',
      rangeStart: '1',
      rangeEnd: '2',
    } as never)
    expect(c.source).toBe('time')
    expect(c.mode).toBe(MATCH_RANGE)
  })
  it('falls back to a default mode for an unknown mode', () => {
    const c = normalizeCondition({ source: 'param', path: 'p', mode: 'bogus' } as never)
    expect(c.mode).toBe(MATCH_EQ)
  })
})

describe('buildRequestRuleExpr', () => {
  it('builds a single exists rule', () => {
    const expr = buildRequestRuleExpr([
      {
        conditions: [{ source: 'param', path: 'model', mode: MATCH_EXISTS, value: '' }],
        multiplier: '1.5',
      },
    ])
    expect(expr).toBe('(param("model") != nil ? 1.5 : 1)')
  })
  it('returns empty string for empty groups', () => {
    expect(buildRequestRuleExpr([])).toBe('')
  })
  it('returns empty when multiplier is not numeric', () => {
    expect(
      buildRequestRuleExpr([
        {
          conditions: [{ source: 'param', path: 'm', mode: MATCH_EXISTS, value: '' }],
          multiplier: 'abc',
        },
      ])
    ).toBe('')
  })
})
