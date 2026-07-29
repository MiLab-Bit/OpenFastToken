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
import { Share2, TrendingUp } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { CopyButton } from '@/components/copy-button'
import type { UserWalletData } from '../types'

// ---- Referral tier config (must match backend referral_rebate_setting defaults) ----
interface ReferralTier {
  name: string
  nameZh: string
  threshold: number
  rate: number
  color: string
}

const REFERRAL_TIERS: ReferralTier[] = [
  { name: 'Starter', nameZh: '入门推广', threshold: 0, rate: 0.01, color: 'bg-muted text-muted-foreground' },
  { name: 'Active', nameZh: '活跃推广', threshold: 500, rate: 0.03, color: 'bg-blue-100 text-blue-700' },
  { name: 'Pro', nameZh: '资深推广', threshold: 2000, rate: 0.05, color: 'bg-purple-100 text-purple-700' },
  { name: 'Partner', nameZh: '金牌合伙人', threshold: 10000, rate: 0.08, color: 'bg-amber-100 text-amber-700' },
]

/** Get current tier based on cumulative referred payment total */
function getCurrentTier(total: number): ReferralTier {
  let match = REFERRAL_TIERS[0]
  for (const t of REFERRAL_TIERS) {
    if (total >= t.threshold) match = t
    else break
  }
  return match
}

interface AffiliateRewardsCardProps {
  user: UserWalletData | null
  affiliateLink: string
  onTransfer: () => void
  complianceConfirmed?: boolean
  loading?: boolean
}

export function AffiliateRewardsCard({
  user,
  affiliateLink,
  onTransfer,
  complianceConfirmed = true,
  loading,
}: AffiliateRewardsCardProps) {
  const { t, i18n } = useTranslation()
  if (loading) {
    return (
      <div className='overflow-hidden rounded-lg border border-stone-card bg-card shadow-sm py-0'>
        <div className='grid gap-4 p-3 sm:p-4 lg:grid-cols-[minmax(220px,1fr)_minmax(220px,0.72fr)_minmax(320px,1.15fr)] lg:items-center'>
          <div>
            <Skeleton className='h-5 w-32' />
            <Skeleton className='mt-2 h-4 w-48' />
          </div>
          <Skeleton className='h-14 rounded-lg' />
          <Skeleton className='h-10 rounded-lg' />
        </div>
      </div>
    )
  }

  const hasRewards = (user?.aff_quota ?? 0) > 0
  const currentTier = getCurrentTier(user?.aff_recharge_total ?? 0)
  const isZh = i18n.language === 'zh' || i18n.language?.startsWith('zh')

  return (
    <div className='overflow-hidden rounded-lg border border-stone-card bg-card shadow-sm py-0'>
      <div className='grid gap-3 p-3 sm:gap-4 sm:p-4 lg:grid-cols-[minmax(200px,1fr)_minmax(180px,0.65fr)_minmax(280px,1fr)] lg:items-center'>
        <div className='flex min-w-0 items-center gap-2.5'>
          <div className='flex size-8 shrink-0 items-center justify-center rounded-lg border border-stone-card bg-stone-bg'>
            <Share2 className='text-accent-brand size-4' />
          </div>
          <div className='min-w-0'>
            <h3 className='truncate font-serif text-lg font-bold text-stone-text'>
              {t('Referral Program')}
            </h3>
            <Badge variant='secondary' className={`${currentTier.color} shrink-0 gap-0.5 border-0 px-2 py-0.5 text-[11px] font-semibold`}>
              <TrendingUp className='size-3' />
              {isZh ? currentTier.nameZh : currentTier.name}
              {' '}
              {(currentTier.rate * 100).toFixed(0)}%
            </Badge>
            <p className='text-stone-muted line-clamp-1 text-xs'>
              {t(
                'Earn rewards when your referrals add funds. Transfer accumulated rewards to your balance anytime.'
              )}
            </p>
          </div>
        </div>

        <div className='grid grid-cols-3 gap-1.5 text-center'>
          {[
            [t('Pending'), formatQuota(user?.aff_quota ?? 0)],
            [t('Total Earned'), formatQuota(user?.aff_history_quota ?? 0)],
            [t('Invites'), String(user?.aff_count ?? 0)],
          ].map(([label, value]) => (
            <div key={label}>
              <div className='text-stone-muted truncate text-[10px] font-medium tracking-wider uppercase'>
                {label}
              </div>
              <div className='mt-0.5 truncate text-sm font-semibold tabular-nums text-stone-text'>
                {value}
              </div>
            </div>
          ))}
        </div>

        <div className='flex items-center gap-2'>
          <Input
            value={affiliateLink}
            readOnly
            className='h-9 min-w-0 flex-1 border-stone-card bg-stone-bg font-mono text-xs text-stone-text'
          />
          <CopyButton
            value={affiliateLink}
            variant='outline'
            className='border-stone-card bg-card text-stone-muted hover:text-accent-brand hover:border-accent-brand size-9 shrink-0'
            iconClassName='size-4'
            tooltip={t('Copy referral link')}
            aria-label={t('Copy referral link')}
          />
          {hasRewards && (
            <Button
              onClick={onTransfer}
              disabled={!complianceConfirmed}
              className='h-9 shrink-0 bg-primary text-primary-foreground hover:bg-primary/90 px-3'
              size='sm'
            >
              {t('Transfer to Balance')}
            </Button>
          )}
        </div>
        {!complianceConfirmed ? (
          <p className='text-stone-muted text-xs lg:col-span-3'>
            {t(
              'Referral reward transfer is disabled until the administrator confirms compliance terms.'
            )}
          </p>
        ) : null}
      </div>
    </div>
  )
}
