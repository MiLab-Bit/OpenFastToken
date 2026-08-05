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

For commercial licensing, please contact hello@fasttoken.example.com
*/
import { useMemo } from 'react'
import { useSystemConfig } from '@/hooks/use-system-config'
import { useTopNavLinks } from '@/hooks/use-top-nav-links'
import { SidebarTrigger } from '@/components/ui/sidebar'
import { Skeleton } from '@/components/ui/skeleton'
import { defaultTopNavLinks } from '../config/top-nav.config'
import { type TopNavLink } from '../types'
import { HeaderLogo } from './header-logo'
import { PublicHeader } from './public-header'

type AppHeaderProps = {
  navLinks?: TopNavLink[]
  showNotifications?: boolean
  showProfileDropdown?: boolean
}

export function AppHeader({
  navLinks = defaultTopNavLinks,
  showNotifications = true,
  showProfileDropdown = true,
}: AppHeaderProps) {
  const dynamicLinks = useTopNavLinks()
  const links = dynamicLinks.length > 0 ? dynamicLinks : navLinks
  const { systemName, logo: systemLogo, loading, logoLoaded } = useSystemConfig()

  const normalizedLinks = useMemo(
    () =>
      links.map((link) => ({
        isActive: false,
        disabled: false,
        external: false,
        requiresAuth: false,
        ...link,
      })),
    [links]
  )

  // Custom logo with SidebarTrigger + full branding (replaces both logo and siteName)
  const fullBranding = (
    <div className='flex items-center gap-2.5'>
      <SidebarTrigger variant='ghost' className='size-8' />
      <div className='flex size-7 shrink-0 items-center justify-center transition-all duration-300'>
        {loading ? (
          <Skeleton className='size-full rounded-lg' />
        ) : (
          <HeaderLogo
            src={systemLogo}
            loading={loading}
            logoLoaded={logoLoaded}
            className='size-full rounded-lg object-contain'
          />
        )}
      </div>
      <span className='text-sm font-serif font-bold tracking-tight'>
        {loading ? <Skeleton className='h-4 w-16' /> : (systemName || 'FastToken')}
      </span>
    </div>
  )

  return (
    <>
      {/* Spacer so content starts below the fixed header */}
      <div className='h-16 shrink-0' />

      <PublicHeader
        navLinks={normalizedLinks}
        logo={fullBranding}
        siteName=' '
        showThemeSwitch={true}
        showLanguageSwitcher={true}
        showNotifications={showNotifications}
        showAuthButtons={showProfileDropdown}
      />
    </>
  )
}
