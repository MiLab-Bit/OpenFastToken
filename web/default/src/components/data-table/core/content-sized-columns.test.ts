import { describe, it, expect } from 'vitest'
import { isContentSizedColumn } from './content-sized-columns'

describe('isContentSizedColumn', () => {
  it('is true only for the actions column', () => {
    expect(isContentSizedColumn('actions')).toBe(true)
  })

  it('is false for any other column id', () => {
    expect(isContentSizedColumn('name')).toBe(false)
    expect(isContentSizedColumn('created_at')).toBe(false)
    expect(isContentSizedColumn('')).toBe(false)
  })
})
