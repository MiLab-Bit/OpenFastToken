import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { render, screen } from '@testing-library/react'
import i18n from '@/l10n/config'
import { GroupBadge } from './group-badge'

// The app's i18n instance defaults to Chinese; pin it to English for this file
// so the group label assertions are deterministic and language-independent.
beforeAll(() => {
  i18n.changeLanguage('en')
})
afterAll(() => {
  i18n.changeLanguage('zh')
})

function ratioSpan(container: HTMLElement, text: string): HTMLElement | undefined {
  return Array.from(container.querySelectorAll('span')).find(
    (s) => s.textContent?.trim() === text
  )
}

describe('GroupBadge', () => {
  it('treats an empty group as "User Group" with a neutral variant', () => {
    const { container } = render(<GroupBadge group={null} />)
    expect(screen.getByText('User Group')).toBeInTheDocument()
    expect(container.querySelector('span')?.className).toContain('bg-muted')
  })

  it('labels the default group', () => {
    render(<GroupBadge group='default' />)
    expect(screen.getByText('Default')).toBeInTheDocument()
  })

  it('labels the auto group', () => {
    render(<GroupBadge group='auto' />)
    expect(screen.getByText('Auto')).toBeInTheDocument()
  })

  it('renders a normal group name as its label', () => {
    render(<GroupBadge group='my-team' />)
    expect(screen.getByText('my-team')).toBeInTheDocument()
  })

  it('prefers an explicit label override', () => {
    render(<GroupBadge group='my-team' label='Override' />)
    expect(screen.getByText('Override')).toBeInTheDocument()
  })

  it('renders only the badge when ratio is null', () => {
    const { container } = render(<GroupBadge group='my-team' />)
    // No ratio span (which would contain "x").
    expect(container.querySelector('span')?.textContent).not.toMatch(/x$/)
  })

  it('colors the ratio badge with warning when ratio > 1', () => {
    const { container } = render(<GroupBadge group='my-team' ratio={1.5} />)
    expect(screen.getByText('1.5x')).toBeInTheDocument()
    expect(ratioSpan(container, '1.5x')?.className).toContain('bg-warning/10')
  })

  it('colors the ratio badge with info when ratio < 1', () => {
    const { container } = render(<GroupBadge group='my-team' ratio={0.5} />)
    expect(screen.getByText('0.5x')).toBeInTheDocument()
    expect(ratioSpan(container, '0.5x')?.className).toContain('bg-info/10')
  })

  it('uses the muted class when ratio equals 1', () => {
    const { container } = render(<GroupBadge group='my-team' ratio={1} />)
    expect(ratioSpan(container, '1x')?.className).toContain('bg-muted')
  })
})
