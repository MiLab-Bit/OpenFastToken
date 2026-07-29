import { describe, it, expect } from 'vitest'
import {
  tryParseRequestRuleExpr,
  splitBillingExprAndRequestRules,
  buildRequestRuleExpr,
  normalizeCondition,
  SOURCE_TIME,
  SOURCE_PARAM,
  SOURCE_HEADER,
  MATCH_EQ,
  MATCH_CONTAINS,
  MATCH_EXISTS,
  MATCH_RANGE,
  MATCH_GT,
  MATCH_GTE,
  MATCH_LT,
  MATCH_LTE,
} from './billing-expr'

// ---------------------------------------------------------------------------
// Gap coverage: request-rule parsing of time conditions and the full
// param/header condition matrix that build*Expr can emit.
// ---------------------------------------------------------------------------

describe('tryParseRequestRuleExpr — time conditions', () => {
  it('parses an open (unwrapped) time range', () => {
    const groups = tryParseRequestRuleExpr(
      '(hour("UTC") >= 9 || hour("UTC") < 17 ? 1.5 : 1)'
    )!
    expect(groups[0].conditions[0]).toEqual({
      source: 'time',
      timeFunc: 'hour',
      timezone: 'UTC',
      mode: MATCH_RANGE,
      value: '',
      rangeStart: '9',
      rangeEnd: '17',
    })
  })

  it('parses a wrapped time range', () => {
    const groups = tryParseRequestRuleExpr(
      '((hour("UTC") >= 9 || hour("UTC") < 17) ? 1.5 : 1)'
    )!
    expect(groups[0].conditions[0].mode).toBe(MATCH_RANGE)
    expect((groups[0].conditions[0] as { rangeStart: string }).rangeStart).toBe('9')
  })

  it('parses a time gte condition', () => {
    const groups = tryParseRequestRuleExpr(
      '(hour("Asia/Shanghai") >= 8 ? 1.2 : 1)'
    )!
    expect(groups[0].conditions[0]).toMatchObject({
      source: 'time',
      timeFunc: 'hour',
      timezone: 'Asia/Shanghai',
      mode: MATCH_GTE,
      value: '8',
    })
  })

  it('parses a time lt condition', () => {
    const groups = tryParseRequestRuleExpr('(minute("UTC") < 30 ? 1.1 : 1)')!
    expect(groups[0].conditions[0]).toMatchObject({
      source: 'time',
      timeFunc: 'minute',
      timezone: 'UTC',
      mode: MATCH_LT,
      value: '30',
    })
  })

  it('parses time eq / gte with other time funcs', () => {
    expect(
      tryParseRequestRuleExpr('(weekday("UTC") == 2 ? 1 : 1)')![0].conditions[0]
    ).toMatchObject({ timeFunc: 'weekday', mode: MATCH_EQ, value: '2' })
    expect(
      tryParseRequestRuleExpr('(month("UTC") >= 1 ? 1 : 1)')![0].conditions[0]
    ).toMatchObject({ timeFunc: 'month', mode: MATCH_GTE, value: '1' })
  })
})

describe('tryParseRequestRuleExpr — param/header condition matrix', () => {
  it('parses a header exists condition', () => {
    const groups = tryParseRequestRuleExpr('(header("x-app") != "" ? 2 : 1)')!
    expect(groups[0].conditions[0]).toEqual({
      source: 'header',
      path: 'x-app',
      mode: MATCH_EXISTS,
      value: '',
    })
  })

  it('round-trips a param contains rule (compound parsed as one CONTAINS condition)', () => {
    const groups = tryParseRequestRuleExpr(
      '(param("m") != nil && has(param("m"), "v") ? 1.2 : 1)'
    )!
    expect(groups).toHaveLength(1)
    expect(groups[0].conditions).toEqual([
      { source: 'param', path: 'm', mode: MATCH_CONTAINS, value: 'v' },
    ])
    expect(groups[0].multiplier).toBe('1.2')
    // round-trips back to the exact original string
    expect(buildRequestRuleExpr(groups)).toBe(
      '(param("m") != nil && has(param("m"), "v") ? 1.2 : 1)'
    )
  })

  it('round-trips a param numeric rule (compound parsed as one numeric condition)', () => {
    const groups = tryParseRequestRuleExpr(
      '(param("n") != nil && param("n") >= 5 ? 1.3 : 1)'
    )!
    expect(groups[0].conditions).toEqual([
      { source: 'param', path: 'n', mode: MATCH_GTE, value: '5' },
    ])
    expect(buildRequestRuleExpr(groups)).toBe(
      '(param("n") != nil && param("n") >= 5 ? 1.3 : 1)'
    )
  })

  it('round-trips a param LT numeric rule', () => {
    const groups = tryParseRequestRuleExpr(
      '(param("n") != nil && param("n") < 3 ? 1.1 : 1)'
    )!
    expect(groups[0].conditions).toEqual([
      { source: 'param', path: 'n', mode: MATCH_LT, value: '3' },
    ])
    expect(buildRequestRuleExpr(groups)).toBe(
      '(param("n") != nil && param("n") < 3 ? 1.1 : 1)'
    )
  })

  it('round-trips a param GT numeric rule', () => {
    const groups = tryParseRequestRuleExpr(
      '(param("n") != nil && param("n") > 5 ? 1.1 : 1)'
    )!
    expect(groups[0].conditions).toEqual([
      { source: 'param', path: 'n', mode: MATCH_GT, value: '5' },
    ])
    expect(buildRequestRuleExpr(groups)).toBe(
      '(param("n") != nil && param("n") > 5 ? 1.1 : 1)'
    )
  })

  it('round-trips a param LTE numeric rule', () => {
    const groups = tryParseRequestRuleExpr(
      '(param("n") != nil && param("n") <= 5 ? 1.1 : 1)'
    )!
    expect(groups[0].conditions).toEqual([
      { source: 'param', path: 'n', mode: MATCH_LTE, value: '5' },
    ])
    expect(buildRequestRuleExpr(groups)).toBe(
      '(param("n") != nil && param("n") <= 5 ? 1.1 : 1)'
    )
  })

  it('still returns null for a truly unparseable compound', () => {
    expect(
      tryParseRequestRuleExpr('(param("x") > 5 ? 1 : 1)')
    ).toBeNull()
    expect(
      tryParseRequestRuleExpr('(param("x") == notAValidLiteral ? 1 : 1)')
    ).toBeNull()
  })

  it('parses equality with string / numeric / boolean literals', () => {
    expect(
      tryParseRequestRuleExpr('(header("h") == "v" ? 1.1 : 1)')![0].conditions[0]
    ).toEqual({
      source: 'header',
      path: 'h',
      mode: MATCH_EQ,
      value: 'v',
    })
    expect(
      tryParseRequestRuleExpr('(param("n") == 3 ? 1 : 1)')![0].conditions[0]
    ).toMatchObject({ mode: MATCH_EQ, value: '3' })
    expect(
      tryParseRequestRuleExpr('(param("n") == true ? 1 : 1)')![0].conditions[0]
    ).toMatchObject({ mode: MATCH_EQ, value: 'true' })
  })

  it('returns null when an equality literal is not parseable', () => {
    expect(
      tryParseRequestRuleExpr('(param("x") == notAValidLiteral ? 1 : 1)')
    ).toBeNull()
  })

  it('parses a multi-condition AND group', () => {
    const groups = tryParseRequestRuleExpr(
      '(param("a") != nil && header("b") != "" ? 1.5 : 1)'
    )!
    expect(groups[0].conditions).toHaveLength(2)
    expect(groups[0].conditions[0]).toMatchObject({ source: 'param', mode: MATCH_EXISTS })
    expect(groups[0].conditions[1]).toMatchObject({ source: 'header', mode: MATCH_EXISTS })
  })
})

describe('splitBillingExprAndRequestRules — multi-rule and fallbacks', () => {
  it('splits multiple rule factors', () => {
    const res = splitBillingExprAndRequestRules(
      '(p*10 + c*20) * (param("a") != nil ? 1.5 : 1) * (header("b") != "" ? 1.2 : 1)'
    )
    expect(res.billingExpr).toBe('p*10 + c*20')
    expect(res.requestRuleExpr).toBe(
      '(param("a") != nil ? 1.5 : 1) * (header("b") != "" ? 1.2 : 1)'
    )
  })

  it('keeps a base without outer parens as-is', () => {
    const res = splitBillingExprAndRequestRules(
      'p*10 * (param("m") != nil ? 1.5 : 1)'
    )
    expect(res.billingExpr).toBe('p*10')
    expect(res.requestRuleExpr).toBe('(param("m") != nil ? 1.5 : 1)')
  })

  it('returns the whole expr when base parts are not a single rule', () => {
    expect(splitBillingExprAndRequestRules('(p*10) * (c*20)')).toEqual({
      billingExpr: '(p*10) * (c*20)',
      requestRuleExpr: '',
    })
  })

  it('does not unwrap a base part whose outer parens close before the end', () => {
    // `(p*10) + (c*20)` has content after the first balanced `)`, so
    // hasFullOuterParens must reject it and leave the base unchanged.
    const res = splitBillingExprAndRequestRules(
      '(p*10) + (c*20) * (param("m") != nil ? 1.5 : 1)'
    )
    expect(res.billingExpr).toBe('(p*10) + (c*20)')
    expect(res.requestRuleExpr).toBe('(param("m") != nil ? 1.5 : 1)')
  })
})

describe('buildRequestRuleExpr — emit the full condition matrix', () => {
  it('emits header exists / contains', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'header', path: 'x', mode: MATCH_EXISTS, value: '' }], multiplier: '2' },
      ])
    ).toBe('(header("x") != "" ? 2 : 1)')
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'header', path: 'x', mode: MATCH_CONTAINS, value: 'v' }], multiplier: '2' },
      ])
    ).toBe('(has(header("x"), "v") ? 2 : 1)')
  })

  it('emits param contains / numeric / equality', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'm', mode: MATCH_CONTAINS, value: 'v' }], multiplier: '1.2' },
      ])
    ).toBe('(param("m") != nil && has(param("m"), "v") ? 1.2 : 1)')
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'n', mode: MATCH_GTE, value: '5' }], multiplier: '1.3' },
      ])
    ).toBe('(param("n") != nil && param("n") >= 5 ? 1.3 : 1)')
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'n', mode: MATCH_EQ, value: '3' }], multiplier: '1' },
      ])
    ).toBe('(param("n") == 3 ? 1 : 1)')
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'n', mode: MATCH_EQ, value: 'true' }], multiplier: '1' },
      ])
    ).toBe('(param("n") == true ? 1 : 1)')
  })

  it('returns empty when a condition path is blank', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: '', mode: MATCH_EXISTS, value: '' }], multiplier: '1' },
      ])
    ).toBe('')
  })

  it('emits time range / gte conditions', () => {
    expect(
      buildRequestRuleExpr([
        {
          conditions: [
            { source: 'time', timeFunc: 'hour', timezone: 'UTC', mode: MATCH_RANGE, value: '', rangeStart: '9', rangeEnd: '17' },
          ],
          multiplier: '1.5',
        },
      ])
    ).toBe('(hour("UTC") >= 9 || hour("UTC") < 17 ? 1.5 : 1)')
    expect(
      buildRequestRuleExpr([
        {
          conditions: [
            { source: 'time', timeFunc: 'hour', timezone: 'Asia/Shanghai', mode: MATCH_GTE, value: '8', rangeStart: '', rangeEnd: '' },
          ],
          multiplier: '1.2',
        },
      ])
    ).toBe('(hour("Asia/Shanghai") >= 8 ? 1.2 : 1)')
  })

  it('returns empty for non-numeric time operands', () => {
    expect(
      buildRequestRuleExpr([
        {
          conditions: [
            { source: 'time', timeFunc: 'hour', timezone: 'UTC', mode: MATCH_RANGE, value: '', rangeStart: 'abc', rangeEnd: '17' },
          ],
          multiplier: '1.5',
        },
      ])
    ).toBe('')
    expect(
      buildRequestRuleExpr([
        {
          conditions: [
            { source: 'time', timeFunc: 'hour', timezone: 'UTC', mode: MATCH_GTE, value: 'abc', rangeStart: '', rangeEnd: '' },
          ],
          multiplier: '1.2',
        },
      ])
    ).toBe('')
  })

  it('joins multiple groups with " * "', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'm', mode: MATCH_EXISTS, value: '' }], multiplier: '1.5' },
        { conditions: [{ source: 'header', path: 'b', mode: MATCH_EXISTS, value: '' }], multiplier: '1.2' },
      ])
    ).toBe('(param("m") != nil ? 1.5 : 1) * (header("b") != "" ? 1.2 : 1)')
  })

  it('wraps a condition containing " || " when combined with others', () => {
    expect(
      buildRequestRuleExpr([
        {
          conditions: [
            { source: 'time', timeFunc: 'hour', timezone: 'UTC', mode: MATCH_RANGE, value: '', rangeStart: '9', rangeEnd: '17' },
            { source: 'param', path: 'm', mode: MATCH_EXISTS, value: '' },
          ],
          multiplier: '1.5',
        },
      ])
    ).toBe('((hour("UTC") >= 9 || hour("UTC") < 17) && param("m") != nil ? 1.5 : 1)')
  })
})

describe('normalizeCondition — fallbacks and null handling', () => {
  it('falls back to hour for an unknown timeFunc', () => {
    const c = normalizeCondition({
      source: 'time',
      timeFunc: 'bogus',
      timezone: 'UTC',
      mode: 'range',
      value: '',
      rangeStart: '1',
      rangeEnd: '2',
    } as never)
    expect((c as { timeFunc: string }).timeFunc).toBe('hour')
  })

  it('falls back to GTE for an unknown time mode', () => {
    const c = normalizeCondition({
      source: 'time',
      timeFunc: 'hour',
      timezone: 'UTC',
      mode: 'bogus',
      value: '',
      rangeStart: '1',
      rangeEnd: '2',
    } as never)
    expect(c.mode).toBe(MATCH_GTE)
  })

  it('coerces null time fields to empty strings', () => {
    const c = normalizeCondition({
      source: 'time',
      timeFunc: 'hour',
      timezone: 'UTC',
      mode: 'gte',
      value: undefined,
      rangeStart: undefined,
      rangeEnd: undefined,
    } as never) as { value: string; rangeStart: string; rangeEnd: string }
    expect(c.value).toBe('')
    expect(c.rangeStart).toBe('')
    expect(c.rangeEnd).toBe('')
  })

  it('falls back to EQ for an unknown header mode', () => {
    const c = normalizeCondition({
      source: 'header',
      path: 'h',
      mode: 'bogus',
      value: 'x',
    } as never)
    expect(c.mode).toBe(MATCH_EQ)
  })

  it('coerces a null param value to an empty string', () => {
    const c = normalizeCondition({
      source: 'param',
      path: 'p',
      mode: 'eq',
      value: undefined,
    } as never) as { value: string }
    expect(c.value).toBe('')
  })

  it('keeps a valid header EXISTS mode', () => {
    const c = normalizeCondition({
      source: 'header',
      path: 'h',
      mode: 'exists',
      value: '',
    } as never)
    expect(c.mode).toBe(MATCH_EXISTS)
    expect(c.source).toBe(SOURCE_HEADER)
  })
})

describe('buildRequestRuleExpr — EQ plain-string literal (buildExprLiteral fallback)', () => {
  it('quotes a non-numeric, non-boolean param EQ value via JSON.stringify', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'm', mode: MATCH_EQ, value: 'hello' }], multiplier: '1' },
      ])
    ).toBe('(param("m") == "hello" ? 1 : 1)')
  })
  it('quotes a non-numeric, non-boolean header EQ value via JSON.stringify', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'header', path: 'h', mode: MATCH_EQ, value: 'world' }], multiplier: '1' },
      ])
    ).toBe('(header("h") == "world" ? 1 : 1)')
  })
})

describe('buildRequestRuleExpr — defensive fallbacks (100% branch coverage)', () => {
  it('returns empty for a numeric condition with a non-numeric value', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'param', path: 'n', mode: MATCH_GTE, value: 'abc' }], multiplier: '1' },
      ])
    ).toBe('')
  })

  it('returns empty when multiplier is undefined', () => {
    const expr = buildRequestRuleExpr([
      { conditions: [{ source: 'param', path: 'm', mode: MATCH_EXISTS, value: '' }], multiplier: undefined } as never,
    ])
    expect(expr).toBe('')
  })

  it('returns empty when conditions is undefined', () => {
    const expr = buildRequestRuleExpr([
      { conditions: undefined, multiplier: '1' } as never,
    ])
    expect(expr).toBe('')
  })

  it('returns empty when called with undefined groups', () => {
    expect(buildRequestRuleExpr(undefined as never)).toBe('')
  })
})

describe('buildRequestRuleExpr — timezone / empty-value fallbacks', () => {
  it('defaults a time condition timezone to Asia/Shanghai when omitted', () => {
    expect(
      buildRequestRuleExpr([
        {
          conditions: [
            { source: 'time', timeFunc: 'hour', timezone: undefined, mode: MATCH_GTE, value: '8', rangeStart: '', rangeEnd: '' } as never,
          ],
          multiplier: '1',
        },
      ])
    ).toBe('(hour("Asia/Shanghai") >= 8 ? 1 : 1)')
  })

  it('serializes a CONTAINS condition with an empty value via JSON.stringify', () => {
    expect(
      buildRequestRuleExpr([
        { conditions: [{ source: 'header', path: 'h', mode: MATCH_CONTAINS, value: '' }], multiplier: '1' },
      ])
    ).toBe('(has(header("h"), "") ? 1 : 1)')
  })
})
