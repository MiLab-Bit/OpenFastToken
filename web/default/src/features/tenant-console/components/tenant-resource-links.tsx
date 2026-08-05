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
import { type ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { ChevronRight, FileText, Key, Wallet } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent } from '@/components/ui/card'

// ============================================================================
// Tenant Resource Links
//
// Shortcuts into the existing, already tenant-scoped resource pages. No new
// endpoints are involved: keys / logs / wallet are filtered by the backend
// using the caller's tenant, so these are plain navigation entries.
// ============================================================================

export function TenantResourceLinks() {
  const { t } = useTranslation()

  return (
    <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
      <ResourceLinkCard
        icon={<Key className='size-4 shrink-0' />}
        title={t('API Keys')}
        description={t('Manage the API keys available to your enterprise.')}
      >
        <Link
          to='/keys'
          className='absolute inset-0'
          aria-label={t('API Keys')}
        />
      </ResourceLinkCard>

      <ResourceLinkCard
        icon={<FileText className='size-4 shrink-0' />}
        title={t('Usage Logs')}
        description={t('Review consumption records shared within your tenant.')}
      >
        <Link
          to='/usage-logs/$section'
          params={{ section: 'common' }}
          className='absolute inset-0'
          aria-label={t('Usage Logs')}
        />
      </ResourceLinkCard>

      <ResourceLinkCard
        icon={<Wallet className='size-4 shrink-0' />}
        title={t('Wallet')}
        description={t('Check balance and top-up history.')}
      >
        <Link
          to='/wallet'
          className='absolute inset-0'
          aria-label={t('Wallet')}
        />
      </ResourceLinkCard>
    </div>
  )
}

type ResourceLinkCardProps = {
  icon: ReactNode
  title: string
  description: string
  children: ReactNode
}

function ResourceLinkCard({
  icon,
  title,
  description,
  children,
}: ResourceLinkCardProps) {
  return (
    <Card size='sm' className='hover:bg-accent/40 relative transition-colors'>
      <CardContent className='flex items-start gap-3'>
        <div className='text-muted-foreground mt-0.5'>{icon}</div>
        <div className='min-w-0 flex-1'>
          <div className='truncate text-sm font-medium'>{title}</div>
          <p className='text-muted-foreground mt-1 text-xs'>{description}</p>
        </div>
        <ChevronRight className='text-muted-foreground mt-0.5 size-4 shrink-0' />
      </CardContent>
      {children}
    </Card>
  )
}
