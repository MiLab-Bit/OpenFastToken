import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'

// `@lobehub/icons` transitively imports `@emoji-mart/data` JSON, which Node's
// ESM loader rejects under vitest (needs an `import ... with { type: 'json' }`
// attribute). The icon library is irrelevant to ProviderBadge's own logic, so
// we mock it to keep the test suite loadable and focused.
vi.mock('@/lib/lobe-icon', () => ({
  getLobeIcon: vi.fn((name?: string | null) =>
    name ? <span data-testid='lobe-icon'>{name}</span> : null
  ),
}))

import { ProviderBadge } from './provider-badge'

describe('ProviderBadge', () => {
  it('renders the provider-badge slot and label', () => {
    const { container } = render(<ProviderBadge label='OpenAI' />)
    expect(container.querySelector('[data-slot="provider-badge"]')).toBeInTheDocument()
    expect(screen.getByText('OpenAI')).toBeInTheDocument()
  })

  it('renders an icon when iconKey is provided', () => {
    render(<ProviderBadge label='OpenAI' iconKey='OpenAI' />)
    expect(screen.getByTestId('lobe-icon')).toBeInTheDocument()
  })

  it('does not render an icon when iconKey is absent', () => {
    const { container } = render(<ProviderBadge label='OpenAI' />)
    const slot = container.querySelector('[data-slot="provider-badge"]') as HTMLElement
    // Only the badge itself should be present (no leading icon wrapper).
    expect(slot.children.length).toBe(1)
  })

  it('uses a neutral label when colorText is false', () => {
    const { container } = render(
      <ProviderBadge label='OpenAI' iconKey='OpenAI' colorText={false} />
    )
    // The badge root span carries `rounded-full`; the leading icon wrapper does not.
    const badge = Array.from(
      container.querySelectorAll('[data-slot="provider-badge"] span')
    ).find((s) => s.className.includes('rounded-full')) as HTMLElement
    expect(badge.className).toContain('bg-muted')
  })
})
