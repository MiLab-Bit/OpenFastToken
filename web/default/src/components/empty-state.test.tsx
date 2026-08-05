/*
Copyright (C) 2023-2026 FastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import { describe, it, expect } from 'vitest'
import { screen } from '@testing-library/react'
import { renderWithProviders } from '@/test/utils'
import { EmptyState } from './empty-state'

describe('EmptyState', () => {
  it('renders the default state without crashing', () => {
    const { container } = renderWithProviders(<EmptyState />)
    // Default Database icon is rendered as an svg.
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('renders a custom title and description', () => {
    renderWithProviders(
      <EmptyState title='No records' description='Try creating one.' />
    )
    expect(screen.getByText('No records')).toBeInTheDocument()
    expect(screen.getByText('Try creating one.')).toBeInTheDocument()
  })

  it('renders custom action content', () => {
    renderWithProviders(
      <EmptyState title='Empty' action={<button type='button'>Reload</button>} />
    )
    expect(
      screen.getByRole('button', { name: 'Reload' })
    ).toBeInTheDocument()
  })
})
