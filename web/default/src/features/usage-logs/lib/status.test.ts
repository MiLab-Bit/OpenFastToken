import { describe, it, expect } from 'vitest'
import { createStatusMapper } from './status'

describe('createStatusMapper', () => {
  const mapper = createStatusMapper({
    active: { label: 'Active', variant: 'success' },
    disabled: { label: 'Disabled', variant: 'danger' },
  })

  it('returns the mapped label', () => {
    expect(mapper.getLabel('active')).toBe('Active')
  })
  it('falls back to a default label for unknown status', () => {
    expect(mapper.getLabel('ghost')).toBe('Unknown')
    expect(mapper.getLabel('ghost', 'N/A')).toBe('N/A')
  })
  it('returns the mapped variant', () => {
    expect(mapper.getVariant('disabled')).toBe('danger')
  })
  it('falls back to a default variant for unknown status', () => {
    expect(mapper.getVariant('ghost')).toBe('neutral')
    expect(mapper.getVariant('ghost', 'warning')).toBe('warning')
  })
})
