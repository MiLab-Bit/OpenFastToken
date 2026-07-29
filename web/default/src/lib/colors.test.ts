/*
 * Unit tests for lib/colors.ts — semantic color maps and helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  colorToBgClass,
  avatarColorMap,
  getAvatarColorClass,
  getBgColorClass,
  CHART_COLORS,
  getChartColor,
  ANNOUNCEMENT_TYPE_COLORS,
  getAnnouncementColorClass,
  stringToColor,
  type SemanticColor,
} from './colors'

const ALL_SEMANTIC: SemanticColor[] = [
  'blue',
  'green',
  'cyan',
  'purple',
  'pink',
  'red',
  'orange',
  'amber',
  'yellow',
  'lime',
  'light-green',
  'teal',
  'light-blue',
  'indigo',
  'violet',
  'grey',
  'slate',
]

describe('color maps are complete', () => {
  it('colorToBgClass covers every semantic color', () => {
    for (const c of ALL_SEMANTIC) {
      expect(colorToBgClass[c]).toBeTruthy()
    }
  })

  it('avatarColorMap covers every semantic color', () => {
    for (const c of ALL_SEMANTIC) {
      expect(avatarColorMap[c]).toBeTruthy()
    }
  })
})

describe('getAvatarColorClass', () => {
  it('maps a name to a valid avatar class', () => {
    const cls = getAvatarColorClass('gpt-4')
    const color = stringToColor('gpt-4')
    expect(cls).toBe(avatarColorMap[color])
  })

  it('is deterministic', () => {
    expect(getAvatarColorClass('alice')).toBe(getAvatarColorClass('alice'))
  })
})

describe('getBgColorClass', () => {
  it('falls back to blue when no color given', () => {
    expect(getBgColorClass()).toBe(colorToBgClass.blue)
    expect(getBgColorClass(undefined)).toBe(colorToBgClass.blue)
  })

  it('returns the mapped class for a known color', () => {
    expect(getBgColorClass('red')).toBe(colorToBgClass.red)
  })

  it('falls back to blue for an unknown color', () => {
    expect(getBgColorClass('not-a-color')).toBe(colorToBgClass.blue)
  })
})

describe('getChartColor', () => {
  it('cycles through the palette by index', () => {
    expect(getChartColor(0)).toBe(CHART_COLORS[0])
    expect(getChartColor(CHART_COLORS.length)).toBe(CHART_COLORS[0])
    expect(getChartColor(CHART_COLORS.length + 1)).toBe(CHART_COLORS[1])
  })
  it('palette has 12 entries', () => {
    expect(CHART_COLORS).toHaveLength(12)
  })
})

describe('announcement color class', () => {
  it('maps the default type', () => {
    expect(getAnnouncementColorClass()).toBe(ANNOUNCEMENT_TYPE_COLORS.default)
    expect(getAnnouncementColorClass('default')).toBe(
      ANNOUNCEMENT_TYPE_COLORS.default
    )
  })
  it('maps a valid type', () => {
    expect(getAnnouncementColorClass('success')).toBe(
      ANNOUNCEMENT_TYPE_COLORS.success
    )
  })
  it('falls back to default for unknown type', () => {
    expect(getAnnouncementColorClass('bogus')).toBe(
      ANNOUNCEMENT_TYPE_COLORS.default
    )
  })
})

describe('stringToColor (semantic)', () => {
  it('returns a semantic color from the tag palette', () => {
    const c = stringToColor('claude-3')
    expect(ALL_SEMANTIC).toContain(c)
  })
  it('is deterministic', () => {
    expect(stringToColor('gpt-4')).toBe(stringToColor('gpt-4'))
  })
  it('handles empty string without throwing', () => {
    expect(ALL_SEMANTIC).toContain(stringToColor(''))
  })
})
