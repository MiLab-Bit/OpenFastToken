import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { Search, ChevronLeft, ChevronRight, X } from 'lucide-react'
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
import { formatTimestampToDate } from '@/lib/format'
import { useRechargeAdmin } from './hooks/use-recharge-admin'

const STATUS_STYLES: Record<string, string> = {
  success: 'bg-green-100 text-green-700',
  completed: 'bg-green-100 text-green-700',
  expired: 'bg-amber-100 text-amber-700',
  refunded: 'bg-red-100 text-red-700',
}

const STATUS_LABELS: Record<string, string> = {
  success: '成功',
  completed: '已完成',
  expired: '已过期',
  refunded: '已退款',
}

const PM_LABELS: Record<string, string> = {
  alipay: '支付宝',
  wxpay: '微信支付',
  admin: '管理员',
  manual: '手动',
}

function StatusBadge({ status }: { status: string }) {
  const { t } = useTranslation()
  const style = STATUS_STYLES[status] ?? 'bg-muted text-muted-foreground'
  const label = STATUS_LABELS[status] ? t(STATUS_LABELS[status]) : status
  return <Badge className={cn('border-0', style)}>{label}</Badge>
}

export function RechargeManagement() {
  const { t } = useTranslation()
  const { topups, total, page, totalPages, loading, setPage, applyFilter } =
    useRechargeAdmin()

  const [draftUserId, setDraftUserId] = useState('')
  const [draftKeyword, setDraftKeyword] = useState('')

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    applyFilter(draftUserId.trim(), draftKeyword.trim())
  }

  const onClear = () => {
    setDraftUserId('')
    setDraftKeyword('')
    applyFilter('', '')
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('充值管理')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4'>
          <Card>
            <CardContent className='py-4'>
              <form
                onSubmit={onSubmit}
                className='mb-3 flex flex-wrap items-end gap-3'
              >
                <div className='flex flex-col gap-1'>
                  <label className='text-sm text-muted-foreground'>{t('用户ID')}</label>
                  <input
                    type='number'
                    value={draftUserId}
                    onChange={(e) => setDraftUserId(e.target.value)}
                    placeholder={t('例如 24')}
                    className='h-9 w-36 rounded-md border border-input bg-background px-3 text-sm'
                  />
                </div>
                <div className='flex flex-col gap-1'>
                  <label className='text-sm text-muted-foreground'>
                    {t('订单号关键字')}
                  </label>
                  <input
                    value={draftKeyword}
                    onChange={(e) => setDraftKeyword(e.target.value)}
                    placeholder={t('留空查全部')}
                    className='h-9 w-56 rounded-md border border-input bg-background px-3 text-sm'
                  />
                </div>
                <Button type='submit' disabled={loading}>
                  <Search className='mr-1 h-4 w-4' />
                  {t('筛选')}
                </Button>
                <Button
                  type='button'
                  variant='ghost'
                  onClick={onClear}
                  disabled={loading}
                >
                  <X className='mr-1 h-4 w-4' />
                  {t('清除')}
                </Button>
              </form>

              <div className='overflow-x-auto'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>{t('用户ID')}</TableHead>
                      <TableHead>{t('金额(¥)')}</TableHead>
                      <TableHead>{t('配额')}</TableHead>
                      <TableHead>{t('支付方式')}</TableHead>
                      <TableHead>{t('订单号')}</TableHead>
                      <TableHead>{t('状态')}</TableHead>
                      <TableHead>{t('创建时间')}</TableHead>
                      <TableHead>{t('完成时间')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {loading ? (
                      Array.from({ length: 8 }).map((_, i) => (
                        <TableRow key={i}>
                          <TableCell colSpan={9}>
                            <Skeleton className='h-8 w-full' />
                          </TableCell>
                        </TableRow>
                      ))
                    ) : topups.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={9}
                          className='py-8 text-center text-muted-foreground'
                        >
                          {t('暂无充值记录')}
                        </TableCell>
                      </TableRow>
                    ) : (
                      topups.map((r) => (
                        <TableRow key={r.id}>
                          <TableCell>{r.id}</TableCell>
                          <TableCell>{r.user_id}</TableCell>
                          <TableCell>
                            {r.money != null ? `¥${r.money}` : '-'}
                          </TableCell>
                          <TableCell>
                            {r.amount != null ? r.amount : '-'}
                          </TableCell>
                          <TableCell>
                            {PM_LABELS[r.payment_method] ? t(PM_LABELS[r.payment_method]) : r.payment_method}
                          </TableCell>
                          <TableCell className='font-mono text-xs'>
                            {r.trade_no || '-'}
                          </TableCell>
                          <TableCell>
                            <StatusBadge status={r.status} />
                          </TableCell>
                          <TableCell className='whitespace-nowrap text-xs'>
                            {formatTimestampToDate(r.create_time)}
                          </TableCell>
                          <TableCell className='whitespace-nowrap text-xs'>
                            {formatTimestampToDate(r.complete_time)}
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>

              <div className='mt-4 flex items-center justify-between'>
                <span className='text-sm text-muted-foreground'>
                  {t('共 {{count}} 条', { count: total })}
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
