import { describe, it, expect } from 'vitest'
import {
  getResolvedColumnClassName,
  getResolvedColumnClassNameFromMap,
  getPinnedColumnMap,
} from './column-pinning'
import type { DataTableColumnClassName, DataTablePinnedColumn } from './types'

const pinnedLeft: DataTablePinnedColumn = {
  columnId: 'name',
  side: 'left',
  className: 'extra',
  headerClassName: 'hdr',
  cellClassName: 'cell',
}

const pinnedRight: DataTablePinnedColumn = {
  columnId: 'name',
  side: 'right',
  className: 'extra',
  headerClassName: 'hdr',
  cellClassName: 'cell',
}

describe('getPinnedColumnMap', () => {
  it('returns undefined when no pinned columns', () => {
    expect(getPinnedColumnMap(undefined)).toBeUndefined()
    expect(getPinnedColumnMap([])).toBeUndefined()
  })

  it('maps columnId -> pinned column entry', () => {
    const map = getPinnedColumnMap([pinnedLeft])
    expect(map?.get('name')).toBe(pinnedLeft)
    expect(map?.get('missing')).toBeUndefined()
  })
})

describe('getResolvedColumnClassName', () => {
  it('returns only custom class when nothing is pinned', () => {
    const fn = getResolvedColumnClassName((id) => `custom-${id}`, undefined)
    expect(fn('name', 'header')).toBe('custom-name')
  })

  it('returns only custom class when column is not pinned', () => {
    const fn = getResolvedColumnClassName(
      (id) => `custom-${id}`,
      [{ columnId: 'other', side: 'left' }]
    )
    expect(fn('name', 'header')).toBe('custom-name')
  })

  it('merges sticky styling for a left-pinned header', () => {
    const fn = getResolvedColumnClassName((id) => `custom-${id}`, [pinnedLeft])
    const out = fn('name', 'header')
    expect(out).toContain('custom-name')
    expect(out).toContain('sticky')
    expect(out).toContain('left-0')
    expect(out).toContain('extra')
    expect(out).toContain('hdr')
    // cell-only class must NOT be present on a header
    expect(out).not.toContain('cell')
  })

  it('merges sticky styling for a right-pinned cell', () => {
    const fn = getResolvedColumnClassName(
      (id) => `custom-${id}`,
      [pinnedRight]
    )
    const out = fn('name', 'cell')
    expect(out).toContain('right-0')
    expect(out).toContain('extra')
    expect(out).toContain('cell')
    // header-only class must NOT be present on a cell
    expect(out).not.toContain('hdr')
  })

  it('works with an undefined custom class name', () => {
    const fn = getResolvedColumnClassName(undefined, [pinnedLeft])
    const out = fn('name', 'header')
    expect(out).toContain('sticky')
    expect(out).toContain('left-0')
  })
})

describe('getResolvedColumnClassNameFromMap', () => {
  it('uses a provided map', () => {
    const map = getPinnedColumnMap([pinnedLeft])
    const fn = getResolvedColumnClassNameFromMap((id) => `x-${id}`, map)
    expect(fn('name', 'header')).toContain('sticky')
    expect(fn('other', 'header')).toBe('x-other')
  })

  it('tolerates an undefined map', () => {
    const fn = getResolvedColumnClassNameFromMap(
      (id) => `x-${id}`,
      undefined
    )
    expect(fn('a', 'header')).toBe('x-a')
  })
})
