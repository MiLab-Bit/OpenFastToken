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
import { useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Building2, Loader2, Send, WalletCards } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { formatQuota } from '@/lib/format'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  getTenantWallet,
  grantTenantQuota,
  requestEnterpriseTopup,
} from '../api'
import type { TenantWalletData } from '../types'

// ============================================================================
// Enterprise Wallet Card
//
// Shown on the Wallet page when the current user belongs to an enterprise
// (user.enterprise_id > 0). Two tiers:
//   - member:  enterprise quota granted to this member (my_quota / my_used_quota)
//   - admin:   additionally the enterprise main wallet (balance / total_granted /
//              total_recycled) + grant & self-recharge actions
// Data comes from GET /api/user/tenant/wallet — the backend resolves the
// tenant from the session, never from client input.
// ============================================================================

interface EnterpriseWalletCardProps {
  enterpriseId?: number
}

export function EnterpriseWalletCard(props: EnterpriseWalletCardProps) {
  const { t } = useTranslation()
  const [grantOpen, setGrantOpen] = useState(false)
  const [topupOpen, setTopupOpen] = useState(false)
  const [grantUserId, setGrantUserId] = useState('')
  const [grantQuota, setGrantQuota] = useState('')
  const [granting, setGranting] = useState(false)
  const [topupAmount, setTopupAmount] = useState('')
  const [topupMethod, setTopupMethod] = useState<'wechat' | 'alipay'>('wechat')
  const [topupLoading, setTopupLoading] = useState(false)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['tenant-wallet'],
    queryFn: getTenantWallet,
    enabled: Boolean(props.enterpriseId && props.enterpriseId > 0),
  })

  const wallet = useMemo<TenantWalletData | undefined>(
    () => data?.data,
    [data]
  )

  if (!props.enterpriseId || props.enterpriseId <= 0) {
    return null
  }

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['tenant-wallet'] })

  const handleGrant = async () => {
    const userId = Number(grantUserId)
    const quota = Number(grantQuota)
    if (!userId || !quota || quota <= 0) return
    setGranting(true)
    try {
      const res = await grantTenantQuota(userId, quota)
      if (res.success) {
        setGrantOpen(false)
        setGrantUserId('')
        setGrantQuota('')
        await refresh()
      }
    } finally {
      setGranting(false)
    }
  }

  const handleTopup = async () => {
    const amount = Number(topupAmount)
    if (!amount || amount <= 0) return
    setTopupLoading(true)
    try {
      const res = await requestEnterpriseTopup(amount, topupMethod)
      if (res.success && res.data) {
        setTopupOpen(false)
        setTopupAmount('')
        if (topupMethod === 'wechat' && res.data.code_url) {
          const payUrl =
            '/pay/' +
            (res.data.trade_no || '') +
            '?amount=' +
            amount +
            '&codeUrl=' +
            encodeURIComponent(res.data.code_url) +
            '&topupAmount=' +
            amount +
            '&enterprise=1'
          window.open(payUrl, '_blank')
        } else if (res.data.pay_link) {
          window.open(res.data.pay_link, '_blank')
        }
        await refresh()
      }
    } finally {
      setTopupLoading(false)
    }
  }

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='flex flex-col gap-4 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5'>
        <div className='flex items-center gap-3'>
          <div className='bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-md'>
            <Building2 className='size-4.5' />
          </div>
          <div>
            <div className='text-foreground text-sm font-semibold'>
              {t('Enterprise Wallet')}
            </div>
            <div className='text-muted-foreground text-xs'>
              {wallet?.is_admin
                ? t('Managed by you — grant quota to members or recharge the main wallet.')
                : t('Quota granted by your enterprise administrator.')}
            </div>
          </div>
        </div>

        <Link
          to='/tenant'
          className='text-primary hover:text-primary/80 text-sm font-medium'
        >
          {t('Open Tenant Console')} →
        </Link>
      </div>

      {isLoading ? (
        <div className='grid grid-cols-2 gap-px border-t bg-border/60 sm:grid-cols-4'>
          {Array.from({ length: wallet?.is_admin ? 4 : 2 }).map((_, i) => (
            <div key={i} className='bg-card px-4 py-3'>
              <Skeleton className='h-3 w-16' />
              <Skeleton className='mt-2 h-6 w-24' />
            </div>
          ))}
        </div>
      ) : (
        <div className='grid grid-cols-2 gap-px border-t bg-border/60 sm:grid-cols-4'>
          <div className='bg-card px-4 py-3'>
            <div className='text-muted-foreground text-xs'>
              {t('My Enterprise Balance')}
            </div>
            <div className='text-foreground mt-1 font-mono text-lg font-bold tabular-nums'>
              {formatQuota(wallet?.my_quota ?? 0)}
            </div>
            <div className='text-muted-foreground/60 mt-0.5 text-xs'>
              {t('Used')}: {formatQuota(wallet?.my_used_quota ?? 0)}
            </div>
          </div>

          {wallet?.is_admin ? (
            <>
              <div className='bg-card px-4 py-3'>
                <div className='text-muted-foreground text-xs'>
                  {t('Enterprise Main Balance')}
                </div>
                <div className='text-foreground mt-1 font-mono text-lg font-bold tabular-nums'>
                  {formatQuota(wallet?.wallet?.balance ?? 0)}
                </div>
                <div className='text-muted-foreground/60 mt-0.5 text-xs'>
                  {t('Total granted')}: {formatQuota(wallet?.wallet?.total_granted ?? 0)}
                </div>
              </div>

              <div className='bg-card flex flex-col justify-center gap-2 px-4 py-3'>
                <Button
                  size='sm'
                  variant='outline'
                  onClick={() => setGrantOpen(true)}
                >
                  <Send className='size-3.5' />
                  {t('Grant Quota')}
                </Button>
                <Button size='sm' variant='outline' onClick={() => setTopupOpen(true)}>
                  <WalletCards className='size-3.5' />
                  {t('Recharge')}
                </Button>
              </div>
            </>
          ) : (
            <div className='bg-card flex items-center px-4 py-3'>
              <span className='text-muted-foreground text-xs'>
                {t('Balance is deducted from enterprise quota first, then personal wallet if insufficient.')}
              </span>
            </div>
          )}
        </div>
      )}

      {/* Grant dialog */}
      <Dialog open={grantOpen} onOpenChange={setGrantOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Grant Quota to Member')}</DialogTitle>
            <DialogDescription>
              {t('Quota is deducted from the enterprise main balance and added to the member.')}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-3'>
            <div className='space-y-1.5'>
              <Label>{t('User ID')}</Label>
              <Input
                type='number'
                value={grantUserId}
                onChange={(e) => setGrantUserId(e.target.value)}
                placeholder={t('Member user id')}
              />
            </div>
            <div className='space-y-1.5'>
              <Label>{t('Quota')}</Label>
              <Input
                type='number'
                value={grantQuota}
                onChange={(e) => setGrantQuota(e.target.value)}
                placeholder={t('Amount to grant')}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setGrantOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button onClick={handleGrant} disabled={granting}>
              {granting && <Loader2 className='size-4 animate-spin' />}
              {t('Grant')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Recharge dialog */}
      <Dialog open={topupOpen} onOpenChange={setTopupOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Recharge Enterprise Wallet')}</DialogTitle>
            <DialogDescription>
              {t('Funds are added to the enterprise main wallet, which you can then grant to members.')}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-3'>
            <div className='space-y-1.5'>
              <Label>{t('Amount')}</Label>
              <Input
                type='number'
                value={topupAmount}
                onChange={(e) => setTopupAmount(e.target.value)}
                placeholder={t('Recharge amount')}
              />
            </div>
            <div className='flex gap-2'>
              <Button
                type='button'
                size='sm'
                variant={topupMethod === 'wechat' ? 'default' : 'outline'}
                onClick={() => setTopupMethod('wechat')}
              >
                {t('WeChat')}
              </Button>
              <Button
                type='button'
                size='sm'
                variant={topupMethod === 'alipay' ? 'default' : 'outline'}
                onClick={() => setTopupMethod('alipay')}
              >
                {t('Alipay')}
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setTopupOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button onClick={handleTopup} disabled={topupLoading}>
              {topupLoading && <Loader2 className='size-4 animate-spin' />}
              {t('Pay')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
