import { describe, it, expect } from 'vitest'
import type { Cell, Table } from '@tanstack/react-table'
import { getCellLabel, renderCellContent, tableHasCompactMeta } from './card-cell-utils'

// Minimal structural stand-ins for the bits of `Cell`/`Table` we touch.
type AnyCell = Cell<unknown, unknown>

function makeCell(opts: {
  header?: unknown
  metaLabel?: string
  cellRenderer?: (ctx: unknown) => unknown
  value?: unknown
}): AnyCell {
  return {
    column: {
      columnDef: {
        header: opts.header,
        meta: opts.metaLabel ? { label: opts.metaLabel } : undefined,
        cell: opts.cellRenderer,
      },
    },
    getContext: () => ({}) as never,
    getValue: () => opts.value,
  } as unknown as AnyCell
}

describe('getCellLabel', () => {
  it('returns the string header directly', () => {
    expect(getCellLabel(makeCell({ header: 'Name' }))).toBe('Name')
  })

  it('prefers meta.label over a non-string header', () => {
    expect(getCellLabel(makeCell({ header: () => 'x', metaLabel: 'Slug' }))).toBe(
      'Slug'
    )
  })

  it('falls back to null when neither is usable', () => {
    expect(getCellLabel(makeCell({ header: () => 'x' }))).toBeNull()
  })
})

describe('renderCellContent', () => {
  it('uses the column cell renderer when present', () => {
    const cell = makeCell({ cellRenderer: () => 'rendered!' })
    // flexRender wraps the renderer in a React element rather than returning
    // the raw string, so we assert a non-null element distinct from the
    // value-based fallback.
    const out = renderCellContent(cell)
    expect(out).toBeTruthy()
    expect(out).not.toBe(42)
  })

  it('falls back to the raw value when there is no renderer', () => {
    const cell = makeCell({ value: 42 })
    expect(renderCellContent(cell)).toBe(42)
  })
})

describe('tableHasCompactMeta', () => {
  function makeTable(columns: Array<{ mobileTitle?: string; mobileBadge?: string }>) {
    return {
      getVisibleLeafColumns: () =>
        columns.map((meta) => ({ columnDef: { meta } })),
    } as unknown as Table<unknown>
  }

  it('is false when no column declares compact meta', () => {
    expect(tableHasCompactMeta(makeTable([{}, {}]))).toBe(false)
  })

  it('is true when a column declares mobileTitle', () => {
    expect(tableHasCompactMeta(makeTable([{ mobileTitle: 'T' }]))).toBe(true)
  })

  it('is true when a column declares mobileBadge', () => {
    expect(tableHasCompactMeta(makeTable([{ mobileBadge: 'B' }]))).toBe(true)
  })
})
