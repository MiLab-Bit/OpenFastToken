import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Circle } from 'lucide-react'
import { StatusBadge, StatusBadgeList, statusPresets } from './status-badge'

function renderBadge(props: React.ComponentProps<typeof StatusBadge>) {
  return render(<StatusBadge {...props} />)
}

describe('StatusBadge', () => {
  it('renders the label text', () => {
    renderBadge({ label: 'Active' })
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('renders children when provided instead of label', () => {
    render(<StatusBadge>child-node</StatusBadge>)
    expect(screen.getByText('child-node')).toBeInTheDocument()
  })

  it('renders the icon when provided', () => {
    const { container } = renderBadge({ label: 'x', icon: Circle })
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('applies the variant surface class', () => {
    const { container } = renderBadge({ label: 'ok', variant: 'success' })
    expect(container.firstChild?.className).toContain('bg-success/10')
  })

  it('uses autoColor to pick a variant via stringToColor', () => {
    const { container } = renderBadge({ label: 'x', autoColor: 'OpenAI' })
    // A concrete surface class from badgeSurfaceMap should be applied.
    expect(container.firstChild?.className).toMatch(/bg-(success|warning|destructive|info|neutral|chart-\d)\/10/)
  })

  it('honors the size variant', () => {
    const { container } = renderBadge({ label: 'x', size: 'lg' })
    expect(container.firstChild?.className).toContain('h-7')
  })

  it('adds the pulse animation when pulse is set', () => {
    const { container } = renderBadge({ label: 'x', pulse: true })
    expect(container.firstChild?.className).toContain('animate-pulse')
  })

  it('is copyable by default (cursor-copy class + title)', () => {
    const { container } = renderBadge({ label: 'copy-me', copyText: 'COPY' })
    const el = container.firstChild as HTMLElement
    expect(el.className).toContain('cursor-copy')
    expect(el.getAttribute('title')).toContain('COPY')
  })

  it('disables copy affordances when copyable is false', () => {
    const { container } = renderBadge({ label: 'x', copyable: false })
    const el = container.firstChild as HTMLElement
    expect(el.className).not.toContain('cursor-copy')
    expect(el.getAttribute('title')).toBeNull()
  })

  it('fires the onClick handler without copying when copyable is false', () => {
    const onClick = vi.fn()
    renderBadge({ label: 'x', copyable: false, onClick })
    fireEvent.click(screen.getByText('x'))
    expect(onClick).toHaveBeenCalledOnce()
  })
})

describe('StatusBadgeList', () => {
  const renderItem = (item: string) => <StatusBadge key={item} label={item} copyable={false} />

  it('renders the empty node when there are no items', () => {
    render(<StatusBadgeList items={[]} renderItem={renderItem} />)
    expect(screen.getByText('-')).toBeInTheDocument()
  })

  it('renders the first `max` items plus an overflow badge', () => {
    render(
      <StatusBadgeList items={['a', 'b', 'c']} renderItem={renderItem} />
    )
    expect(screen.getByText('a')).toBeInTheDocument()
    expect(screen.getByText('b')).toBeInTheDocument()
    expect(screen.getByText('+1')).toBeInTheDocument()
  })

  it('respects a custom max', () => {
    render(
      <StatusBadgeList items={['a', 'b', 'c', 'd']} max={1} renderItem={renderItem} />
    )
    expect(screen.getByText('a')).toBeInTheDocument()
    expect(screen.getByText('+3')).toBeInTheDocument()
  })

  it('uses a custom moreLabel', () => {
    render(
      <StatusBadgeList
        items={['a', 'b', 'c']}
        max={1}
        moreLabel={(n) => `more ${n}`}
        renderItem={renderItem}
      />
    )
    expect(screen.getByText('more 2')).toBeInTheDocument()
  })

  it('honors a custom getKey', () => {
    render(
      <StatusBadgeList
        items={['a', 'b']}
        getKey={(item) => `k-${item}`}
        renderItem={renderItem}
      />
    )
    expect(screen.getByText('a')).toBeInTheDocument()
  })
})

describe('statusPresets', () => {
  it('exposes the documented preset variants', () => {
    expect(statusPresets.active.variant).toBe('success')
    expect(statusPresets.inactive.variant).toBe('neutral')
    expect(statusPresets.suspended.variant).toBe('danger')
    expect(statusPresets.pending.pulse).toBe(true)
  })
})
