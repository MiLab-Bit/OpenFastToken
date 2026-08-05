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
import { Building2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { StatusBadge, type StatusVariant } from '@/components/status-badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { type TenantInfo } from '../types'

// ============================================================================
// Tenant Info Card — enterprise name / membership level / status / discount
// ============================================================================

const MEMBERSHIP_LEVEL_VARIANTS: Record<string, StatusVariant> = {
  silver: 'neutral',
  gold: 'warning',
  platinum: 'purple',
}

const TENANT_STATUS_VARIANTS: Record<string, StatusVariant> = {
  approved: 'success',
  pending: 'warning',
  rejected: 'danger',
}

type TenantInfoCardProps = {
  info?: TenantInfo
  isLoading?: boolean
}

export function TenantInfoCard({ info, isLoading }: TenantInfoCardProps) {
  const { t } = useTranslation()

  // Literal translation keys keep the i18n extraction script able to find them.
  const levelLabels: Record<string, string> = {
    silver: t('Silver'),
    gold: t('Gold'),
    platinum: t('Platinum'),
  }
  const statusLabels: Record<string, string> = {
    approved: t('Approved'),
    pending: t('Pending'),
    rejected: t('Rejected'),
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className='h-6 w-40' />
        </CardHeader>
        <CardContent className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className='h-12 w-full' />
          ))}
        </CardContent>
      </Card>
    )
  }

  if (!info || !info.joined) {
    return (
      <Card>
        <CardContent className='text-muted-foreground py-6 text-center text-sm'>
          {t('You do not belong to any enterprise yet.')}
        </CardContent>
      </Card>
    )
  }

  const level = info.membership_level ?? ''
  const status = info.status ?? ''
  // Backend returns a multiplier (0.7 => 7 折). Render it on the 10-point scale
  // used elsewhere in the console.
  const discountLabel =
    typeof info.discount_rate === 'number'
      ? (info.discount_rate * 10).toFixed(1)
      : '-'

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <Building2 className='text-muted-foreground size-4 shrink-0' />
          <span className='truncate'>{info.name || t('Enterprise')}</span>
        </CardTitle>
      </CardHeader>
      <CardContent className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
        <TenantInfoField label={t('Membership Level')}>
          {level ? (
            <StatusBadge
              label={levelLabels[level] ?? level}
              variant={MEMBERSHIP_LEVEL_VARIANTS[level] ?? 'neutral'}
              copyable={false}
            />
          ) : (
            <span className='text-sm'>-</span>
          )}
        </TenantInfoField>

        <TenantInfoField label={t('Status')}>
          {status ? (
            <StatusBadge
              label={statusLabels[status] ?? status}
              variant={TENANT_STATUS_VARIANTS[status] ?? 'neutral'}
              copyable={false}
            />
          ) : (
            <span className='text-sm'>-</span>
          )}
        </TenantInfoField>

        <TenantInfoField label={t('Discount')}>
          <span className='text-sm font-medium tabular-nums'>
            {discountLabel}
          </span>
        </TenantInfoField>

        <TenantInfoField label={t('Members')}>
          <span className='text-sm font-medium tabular-nums'>
            {info.total_members ?? 0}
          </span>
        </TenantInfoField>
      </CardContent>
    </Card>
  )
}

function TenantInfoField({
  label,
  children,
}: {
  label: string
  children: ReactNode
}) {
  return (
    <div className='flex flex-col gap-1.5'>
      <span className='text-muted-foreground text-xs'>{label}</span>
      <div className='flex min-h-6 items-center'>{children}</div>
    </div>
  )
}
