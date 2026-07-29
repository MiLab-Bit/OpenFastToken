/*
Copyright (C) 2023-2026 OpenFastToken

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

For commercial licensing, please contact support@example.com
*/
import type { TopNavLink } from '../types'
import { PublicHeader, type PublicHeaderProps } from './public-header'
import { PublicFooter } from './public-footer'

type PublicLayoutProps = {
  children: React.ReactNode
  navContent?: React.ReactNode
  headerProps?: Omit<PublicHeaderProps, 'navContent'>
  navLinks?: TopNavLink[]
  showThemeSwitch?: boolean
  showAuthButtons?: boolean
  showNotifications?: boolean
  logo?: React.ReactNode
  siteName?: string
  showMainContainer?: boolean
}

export function PublicLayout(props: PublicLayoutProps) {
  return (
    <div className='bg-stone-bg text-stone-text relative min-h-svh overflow-x-clip flex flex-col'>
      {/* Portfolio-style subtle background — dot grid is on body */}
      <div aria-hidden className='pointer-events-none fixed inset-0 -z-10'>
        <div
          className='absolute inset-0 opacity-[0.03]'
          style={{
            background:
              'radial-gradient(ellipse 80% 60% at 10% 0%, oklch(0.45 0.12 25 / 60%) 0%, transparent 60%)',
          }}
        />
      </div>

      <PublicHeader
        navContent={props.navContent}
        navLinks={props.navLinks}
        showThemeSwitch={props.showThemeSwitch}
        showAuthButtons={props.showAuthButtons}
        showNotifications={props.showNotifications}
        logo={props.logo}
        siteName={props.siteName}
        {...props.headerProps}
      />

      {/* Content with consistent top padding to clear the fixed header */}
      <main className='flex-1 pt-20'>{props.children}</main>

      <PublicFooter />
    </div>
  )
}