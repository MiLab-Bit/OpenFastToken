/*
Copyright (C) 2023-2026 OpenFastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without any the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@example.com
*/
import { useTranslation } from 'react-i18next'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { SectionPageLayout } from '@/components/layout'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { ExportCsvButton } from '@/components/export-csv-button'
import { ROLE } from '@/lib/roles'
import {
  useGiftsAdmin,
  type GiftStatus,
  type GiftStatusFilter,
} from './hooks/use-gifts-admin'

const STATUS_STYLES: Record<GiftStatus, string> = {
  active: 'bg-green-100 text-green-700',
  used: 'bg-muted text-muted-foreground',
  expired: 'bg-amber-100 text-amber-700',
}

const STATUS_LABEL_KEY: Record<GiftStatus, string> = {
  active: 'Active Gifts',
  used: 'Used Gifts',
  expired: 'Expired Gifts',
}

const STATUS_FILTERS: GiftStatusFilter[] = ['all', 'active', 'used', 'expired']

function StatCard({
  label,
  value,
}: {
  label: string
  value: number
}) {
  return (
    <Card className='flex-1'>
      <CardContent className='py-4'>
        <div className='text-sm text-muted-foreground'>{label}</div>
        <div className='mt-1 text-2xl font-semibold'>{value}</div>
      </CardContent>
    </Card>
  )
}

export function GiftManagement() {
  const { t } = useTranslation()
  const {
    gifts,
    total,
    page,
    totalPages,
    statusFilter,
    loading,
    updatingId,
    setPage,
    setStatusFilter,
    updateStatus,
  } = useGiftsAdmin()

  const activeCount = gifts.filter((g) => g.status === 'active').length
  const usedCount = gifts.filter((g) => g.status === 'used').length

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Gift Management')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4'>
          <div className='flex flex-wrap gap-3'>
            <StatCard label={t('Total Gifts')} value={total} />
            <StatCard label={t('Active Gifts')} value={activeCount} />
            <StatCard label={t('Used Gifts')} value={usedCount} />
          </div>

          <Card>
            <CardContent className='py-4'>
              <div className='mb-3 flex items-center justify-between gap-2'>
                <h3 className='text-base font-medium'>{t('Gift Management')}</h3>
                <div className='flex items-center gap-2'>
                  <select
                  id='gift-status-filter'
                  name='statusFilter'
                  aria-label={t('Status')}
                  className='h-9 rounded-md border border-input bg-background px-3 text-sm'
                  value={statusFilter}
                  onChange={(e) =>
                    setStatusFilter(e.target.value as GiftStatusFilter)
                  }
                >
                  {STATUS_FILTERS.map((s) => (
                    <option key={s} value={s}>
                      {s === 'all'
                        ? t('All')
                        : t(STATUS_LABEL_KEY[s as GiftStatus])}
                    </option>
                  ))}
                </select>
                <ExportCsvButton url='/api/admin/gifts/export' requireRole={ROLE.SUPER_ADMIN} params={statusFilter === 'all' ? undefined : { status: statusFilter }} />
                </div>
              </div>

              <div className='overflow-x-auto'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>{t('User ID')}</TableHead>
                      <TableHead>{t('Username')}</TableHead>
                      <TableHead>{t('Gift Type')}</TableHead>
                      <TableHead>{t('Gift Name')}</TableHead>
                      <TableHead>{t('Order Number')}</TableHead>
                      <TableHead>{t('Status')}</TableHead>
                      <TableHead>{t('Description')}</TableHead>
                      <TableHead>{t('Actions')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {loading ? (
                      Array.from({ length: 5 }).map((_, i) => (
                        <TableRow key={i}>
                          <TableCell colSpan={9}>
                            <Skeleton className='h-8 w-full' />
                          </TableCell>
                        </TableRow>
                      ))
                    ) : gifts.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={8}
                          className='py-8 text-center text-muted-foreground'
                        >
                          {t('No gifts yet')}
                        </TableCell>
                      </TableRow>
                    ) : (
                      gifts.map((g) => (
                        <TableRow key={g.id}>
                          <TableCell>{g.id}</TableCell>
                          <TableCell className='max-w-[180px] truncate'>
                            {g.user_id}
                          </TableCell>
                          <TableCell>{g.username}</TableCell>
                          <TableCell>{g.gift_type}</TableCell>
                          <TableCell>{g.gift_name}</TableCell>
                          <TableCell className='font-mono text-xs'>
                            {g.trade_no || '-'}
                          </TableCell>
                          <TableCell>
                            <Badge
                              className={cn(
                                'border-0',
                                STATUS_STYLES[g.status]
                              )}
                            >
                              {t(STATUS_LABEL_KEY[g.status])}
                            </Badge>
                          </TableCell>
                          <TableCell className='max-w-[220px] truncate'>
                            {g.description || '-'}
                          </TableCell>
                          <TableCell>
                            <div className='flex items-center gap-1'>
                              {g.status === 'active' ? (
                                <>
                                  <Button
                                    size='sm'
                                    variant='outline'
                                    disabled={updatingId === g.id}
                                    onClick={() => updateStatus(g.id, 'used')}
                                  >
                                    {t('Redeem Gift')}
                                  </Button>
                                  <Button
                                    size='sm'
                                    variant='ghost'
                                    disabled={updatingId === g.id}
                                    onClick={() => updateStatus(g.id, 'expired')}
                                  >
                                    {t('Mark Expired')}
                                  </Button>
                                </>
                              ) : (
                                <span className='text-xs text-muted-foreground'>
                                  {'—'}
                                </span>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>

              <div className='mt-4 flex items-center justify-between'>
                <span className='text-sm text-muted-foreground'>
                  {t('Total')}: {total}
                </span>
                <div className='flex items-center gap-2'>
                  <Button
                    variant='outline'
                    size='sm'
                    disabled={page <= 1 || loading}
                    onClick={() => setPage(Math.max(1, page - 1))}
                  >
                    <ChevronLeft className='h-4 w-4' />
                  </Button>
                  <span className='text-sm'>
                    {page} / {totalPages}
                  </span>
                  <Button
                    variant='outline'
                    size='sm'
                    disabled={page >= totalPages || loading}
                    onClick={() => setPage(Math.min(totalPages, page + 1))}
                  >
                    <ChevronRight className='h-4 w-4' />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
