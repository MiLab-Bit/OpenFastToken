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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Skeleton } from '@/components/ui/skeleton'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='relative grid h-svh max-w-none overflow-hidden'>
      {/* Subtle tech gradient — matches PublicLayout aesthetic */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10'
      >
        <div
          className='absolute inset-0 opacity-40 dark:opacity-[0.06]'
          style={{
            background: [
              'radial-gradient(ellipse 80% 60% at 10% 0%, oklch(0.75 0.05 250 / 40%) 0%, transparent 60%)',
              'radial-gradient(ellipse 60% 50% at 90% 5%, oklch(0.75 0.04 200 / 30%) 0%, transparent 55%)',
              'radial-gradient(ellipse 50% 40% at 50% 100%, oklch(0.80 0.04 280 / 25%) 0%, transparent 50%)',
            ].join(', '),
          }}
        />
        <div
          className='absolute inset-0'
          style={{
            backgroundImage:
              'radial-gradient(var(--border) 1px, transparent 1px)',
            backgroundSize: '3rem 3rem',
            maskImage:
              'radial-gradient(ellipse 80% 80% at 50% 50%, black 30%, transparent 100%)',
            WebkitMaskImage:
              'radial-gradient(ellipse 80% 80% at 50% 50%, black 30%, transparent 100%)',
            opacity: '0.35',
          }}
        />
      </div>

      {/* Logo + brand name */}
      <Link
        to='/'
        className='absolute top-4 left-4 z-10 flex items-center gap-2 transition-opacity hover:opacity-80 sm:top-8 sm:left-8'
      >
        <div className='relative h-8 w-8'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-full' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='h-8 w-8 rounded-full object-cover'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-6 w-24' />
        ) : (
          <h1 className='text-xl font-medium'>{systemName}</h1>
        )}
      </Link>

      {/* Auth form — centered, white card surface */}
      <div className='container flex items-center pt-16 sm:pt-0'>
        <div className='mx-auto flex w-full flex-col justify-center space-y-2 px-4 py-8 sm:w-[480px] sm:p-8'>
          {children}
        </div>
      </div>
    </div>
  )
}