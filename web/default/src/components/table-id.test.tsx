import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TableId } from './table-id'

describe('TableId', () => {
  it('renders a numeric id', () => {
    render(<TableId value={123} />)
    expect(screen.getByText('123')).toBeInTheDocument()
  })

  it('renders a string id', () => {
    render(<TableId value='abc-1' />)
    expect(screen.getByText('abc-1')).toBeInTheDocument()
  })

  it('applies a custom className alongside the base classes', () => {
    const { container } = render(<TableId value={1} className='my-id' />)
    const el = container.firstChild as HTMLElement
    expect(el.className).toContain('font-mono')
    expect(el.className).toContain('my-id')
  })
})
