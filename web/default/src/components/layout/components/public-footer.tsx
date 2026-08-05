/*
Copyright (C) 2023-2026 FastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact hello@fasttoken.example.com
*/
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Skeleton } from '@/components/ui/skeleton'

type PublicFooterProps = {
  className?: string
}

export function PublicFooter({ className }: PublicFooterProps) {
  const { t } = useTranslation()
  const { systemName, logo: systemLogo, loading } = useSystemConfig()

  const displayLogo = systemLogo || '/logo.svg'
  const displayName = systemName || 'FastToken'
  const currentYear = new Date().getFullYear()

  return (
    <footer
      className={`bg-stone-bg border-t border-stone-card text-stone-muted mt-auto ${className ?? ''}`}
    >
      <div className='container mx-auto flex flex-col items-center justify-center gap-3 px-4 py-8 text-center'>
        {/* Brand: Logo + Name */}
        <Link to='/' className='group inline-flex items-center gap-2.5'>
          <img
            src={displayLogo}
            alt={displayName}
            className='size-7 rounded-lg object-contain'
          />
          <span className='text-sm font-serif font-bold tracking-tight'>
            {loading ? (
              <Skeleton className='inline-block h-4 w-24' />
            ) : (
              displayName
            )}
          </span>
        </Link>

        {/* Tagline */}
        <p className='text-muted-foreground/60 max-w-[280px] text-xs leading-relaxed'>
          {t('Powerful API Management Platform')}
        </p>

        {/* Copyright + Attribution */}
        <p className='font-serif text-xs tracking-wide'>
          &copy; {currentYear}{' '}
          <span className='font-serif font-bold'>{displayName}</span>
          .{' '}
          {t('All rights reserved.')}{' '}
          <a
            href='https://fasttoken.example.com/docs'
            target='_blank'
            rel='noopener noreferrer'
            className='text-muted-foreground hover:text-accent-brand transition-colors'
          >
            {' '}{'FastToken'}
          </a>
        </p>

        {/* Legal Links */}
        <nav className='flex flex-wrap items-center justify-center gap-4 text-xs'>
          <Link
            to='/privacy-policy'
            className='hover:text-stone-text transition-colors duration-500'
          >
            {t('Privacy Policy')}
          </Link>
          <Link
            to='/user-agreement'
            className='hover:text-stone-text transition-colors duration-500'
          >
            {t('User Agreement')}
          </Link>
        </nav>
      </div>
    </footer>
  )
}
