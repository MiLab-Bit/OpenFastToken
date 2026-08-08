/*
Copyright (C) 2023-2026 智企惠

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

For commercial licensing, please contact abovetigers@qq.com
*/
import { useState } from 'react'
import { Copy, Check, ChevronLeft, ChevronRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota, formatTimestamp } from '@/lib/format'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { StatusBadge } from '@/components/status-badge'
import { useEnterpriseWalletTxns } from '../hooks/use-enterprise-wallet-txns'
import { getTxnTypeConfig } from '../lib/billing'
import type { EnterpriseWalletTxnType } from '../types'

interface EnterpriseWalletTxnsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Income types increase the main wallet; expense types decrease it.
const INCOME_TYPES: EnterpriseWalletTxnType[] = ['recharge', 'refund', 'recycle']

export function EnterpriseWalletTxnsDialog({
  open,
  onOpenChange,
}: EnterpriseWalletTxnsDialogProps) {
  const { t } = useTranslation()
  const {
    items,
    total,
    page,
    pageSize,
    loading,
    error,
    handlePageChange,
    handlePageSizeChange,
    retry,
  } = useEnterpriseWalletTxns({ enabled: open })

  const [copied, setCopied] = useState<string | null>(null)
  const { copyToClipboard } = useCopyToClipboard({ notify: false })

  const totalPages = Math.ceil(total / pageSize)

  const renderAmount = (txn: { type: EnterpriseWalletTxnType; amount: number }) => {
    const isIncome = INCOME_TYPES.includes(txn.type)
    const sign = isIncome ? '+' : '−' // U+2212 MINUS SIGN
    const color = isIncome ? 'text-success' : 'text-destructive'
    return (
      <span className={`font-semibold tabular-nums ${color}`}>
        {sign}
        {formatQuota(Math.abs(txn.amount))}
      </span>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='flex max-h-[calc(100dvh-2rem)] flex-col max-sm:h-dvh max-sm:w-screen max-sm:max-w-none max-sm:rounded-none max-sm:p-4 sm:max-w-3xl'>
        <DialogHeader>
          <DialogTitle>{t('Fund Flow')}</DialogTitle>
          <DialogDescription>
            {t('View enterprise wallet recharge, grant, recycle, consume and refund records')}
          </DialogDescription>
        </DialogHeader>

        <div className='min-h-0 flex-1 space-y-3 sm:space-y-4'>
          {/* Toolbar: page size only (no keyword search in Phase 1) */}
          <div className='flex items-center justify-end gap-2'>
            <Select
              items={[
                { value: '10', label: t('10 / page') },
                { value: '20', label: t('20 / page') },
                { value: '50', label: t('50 / page') },
                { value: '100', label: t('100 / page') },
              ]}
              value={pageSize.toString()}
              onValueChange={(value) =>
                value !== null && handlePageSizeChange(parseInt(value))
              }
            >
              <SelectTrigger className='h-9 w-[92px] sm:w-32'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent alignItemWithTrigger={false}>
                <SelectGroup>
                  <SelectItem value='10'>{t('10 / page')}</SelectItem>
                  <SelectItem value='20'>{t('20 / page')}</SelectItem>
                  <SelectItem value='50'>{t('50 / page')}</SelectItem>
                  <SelectItem value='100'>{t('100 / page')}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <ScrollArea className='h-[calc(100dvh-15rem)] pr-3 sm:h-[460px] sm:pr-4'>
            {loading ? (
              <div className='space-y-3'>
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className='rounded-lg border p-3 sm:p-4'>
                    <div className='flex items-start justify-between'>
                      <Skeleton className='h-4 w-48' />
                      <Skeleton className='h-5 w-16' />
                    </div>
                    <div className='mt-3 grid grid-cols-3 gap-3'>
                      <Skeleton className='h-3 w-full' />
                      <Skeleton className='h-3 w-full' />
                      <Skeleton className='h-3 w-full' />
                    </div>
                  </div>
                ))}
              </div>
            ) : error ? (
              <div className='text-muted-foreground flex h-[320px] flex-col items-center justify-center text-center'>
                <p className='text-sm font-medium'>{t('Failed to load fund flow')}</p>
                <Button variant='outline' size='sm' className='mt-3' onClick={retry}>
                  {t('Retry')}
                </Button>
              </div>
            ) : items.length === 0 ? (
              <div className='text-muted-foreground flex h-[320px] flex-col items-center justify-center text-center'>
                <p className='text-sm font-medium'>{t('No fund flow records')}</p>
                <p className='mt-1 text-xs'>{t('Your fund flow records will appear here')}</p>
              </div>
            ) : (
              <div className='space-y-3'>
                {items.map((txn) => {
                  const cfg = getTxnTypeConfig(txn.type)
                  return (
                    <div
                      key={txn.id}
                      className='hover:bg-muted/50 rounded-lg border p-3 transition-colors sm:p-4'
                    >
                      {/* Header: trade no + copy, time, type badge */}
                      <div className='flex items-start justify-between gap-2'>
                        <div className='flex-1 space-y-1'>
                          <div className='flex min-w-0 items-center gap-2'>
                            <code className='text-foreground truncate font-mono text-sm'>
                              {txn.trade_no}
                            </code>
                            <Button
                              variant='ghost'
                              size='sm'
                              className='h-5 w-5 p-0'
                              onClick={() => {
                                copyToClipboard(txn.trade_no)
                                setCopied(txn.trade_no)
                                setTimeout(() => setCopied(null), 1500)
                              }}
                            >
                              {copied === txn.trade_no ? (
                                <Check className='h-3 w-3' />
                              ) : (
                                <Copy className='h-3 w-3' />
                              )}
                            </Button>
                          </div>
                          <div className='text-muted-foreground text-xs'>
                            {formatTimestamp(txn.created_at)}
                          </div>
                        </div>
                        <StatusBadge
                          label={t(cfg.label)}
                          variant={cfg.variant}
                          copyable={false}
                        />
                      </div>

                      {/* Details grid */}
                      <div className='mt-3 grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4'>
                        <div className='space-y-1'>
                          <span className='text-muted-foreground text-xs'>
                            {t('Amount')}
                          </span>
                          <div>{renderAmount(txn)}</div>
                        </div>
                        <div className='space-y-1'>
                          <span className='text-muted-foreground text-xs'>
                            {t('Balance After')}
                          </span>
                          <div className='text-foreground text-sm font-medium font-mono tabular-nums'>
                            {formatQuota(txn.balance_after)}
                          </div>
                        </div>
                        <div className='space-y-1'>
                          <span className='text-muted-foreground text-xs'>
                            {t('Operator')}
                          </span>
                          <div className='text-sm font-medium'>
                            {txn.operator_id > 0 ? (
                              <StatusBadge
                                label={`${t('User ID')}: ${txn.operator_id}`}
                                variant='neutral'
                                size='sm'
                                copyText={String(txn.operator_id)}
                              />
                            ) : (
                              t('System')
                            )}
                            {txn.user_id > 0 && (
                              <span className='text-muted-foreground ml-1 text-xs'>
                                {t('Related User')}: {txn.user_id}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </ScrollArea>

          {/* Pagination */}
          {!loading && !error && items.length > 0 && (
            <div className='flex flex-col items-center gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between'>
              <div className='text-muted-foreground text-xs sm:text-sm'>
                {t('Showing')} {(page - 1) * pageSize + 1}-
                {Math.min(page * pageSize, total)} {t('of')} {total}
              </div>
              <div className='flex items-center gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => handlePageChange(page - 1)}
                  disabled={page <= 1}
                  className='h-8 w-8 p-0'
                >
                  <ChevronLeft className='h-4 w-4' />
                </Button>
                <div className='text-muted-foreground flex items-center gap-1 text-sm'>
                  <span className='font-medium'>{page}</span>
                  <span>/</span>
                  <span>{totalPages}</span>
                </div>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => handlePageChange(page + 1)}
                  disabled={page >= totalPages}
                  className='h-8 w-8 p-0'
                >
                  <ChevronRight className='h-4 w-4' />
                </Button>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
