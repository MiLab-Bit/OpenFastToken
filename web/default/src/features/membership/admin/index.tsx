/*
Copyright (C) 2023-2026 智企惠
*/
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { toast } from 'sonner'
import { ExportCsvButton } from '@/components/export-csv-button'
import { ROLE } from '@/lib/roles'
import {
  MEMBERSHIP_LEVEL_CONFIGS,
  ENTERPRISE_STATUS_MAP,
} from '../constants'
import {
  listEnterprises,
  approveEnterprise,
  rejectEnterprise,
} from '../api'
import type {
  Enterprise,
} from '../types'

export function EnterpriseAdmin() {
  const { i18n } = useTranslation()
  const isZh = i18n.language === 'zh' || i18n.language?.startsWith('zh')

  // Enterprises state
  const [enterprises, setEnterprises] = useState<Enterprise[]>([])
  const [, setEntTotal] = useState(0)
  const [entPage, setEntPage] = useState(1)
  const [entStatusFilter, setEntStatusFilter] = useState('')
  const [, setEntLoading] = useState(false)

  // Reject dialog state
  const [rejectDialogOpen, setRejectDialogOpen] = useState(false)
  const [rejectingEnterpriseId, setRejectingEnterpriseId] = useState(0)
  const [rejectReason, setRejectReason] = useState('')


  // Fetch enterprises
  const fetchEnterprises = useCallback(async () => {
    try {
      setEntLoading(true)
      const res = await listEnterprises({
        status: entStatusFilter,
        page: entPage,
        page_size: 20,
      })
      if (res.success && res.data) {
        setEnterprises(res.data.enterprises)
        setEntTotal(res.data.total)
      }
    } catch (error) {
      console.error('Failed to fetch enterprises:', error)
    } finally {
      setEntLoading(false)
    }
  }, [entPage, entStatusFilter])


  useEffect(() => {
    fetchEnterprises()
  }, [fetchEnterprises])

  // Approve enterprise
  const handleApprove = async (id: number) => {
    try {
      const res = await approveEnterprise(id)
      if (res.success) {
        toast.success(isZh ? '已审核通过' : 'Approved')
        fetchEnterprises()
      } else {
        toast.error(res.message || (isZh ? '审核失败' : 'Failed to approve'))
      }
    } catch {
      toast.error(isZh ? '审核失败' : 'Failed to approve')
    }
  }

  // Reject enterprise
  const handleReject = async () => {
    try {
      const res = await rejectEnterprise(rejectingEnterpriseId, rejectReason)
      if (res.success) {
        toast.success(isZh ? '已拒绝' : 'Rejected')
        setRejectDialogOpen(false)
        fetchEnterprises()
      } else {
        toast.error(res.message || (isZh ? '拒绝失败' : 'Failed to reject'))
      }
    } catch {
      toast.error(isZh ? '拒绝失败' : 'Failed to reject')
    }
  }

  const openRejectDialog = (id: number) => {
    setRejectingEnterpriseId(id)
    setRejectReason('')
    setRejectDialogOpen(true)
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {isZh ? '企业审核管理' : 'Enterprise Management'}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
                              {/* Enterprises Section */}
          <Card>
            <CardHeader className='flex flex-row items-center justify-between'>
              <CardTitle>{isZh ? '企业审核' : 'Enterprise Review'}</CardTitle>
              <div className='flex items-center gap-2'>
              <Select
                value={entStatusFilter}
                onValueChange={(v) => {
                  setEntStatusFilter(v === 'all' ? '' : (v ?? ''))
                  setEntPage(1)
                }}
              >
                <SelectTrigger className='w-[150px]'>
                  <SelectValue placeholder={isZh ? '全部状态' : 'All'} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='all'>
                    {isZh ? '全部' : 'All'}
                  </SelectItem>
                  <SelectItem value='pending'>
                    {isZh ? '待审核' : 'Pending'}
                  </SelectItem>
                  <SelectItem value='approved'>
                    {isZh ? '已通过' : 'Approved'}
                  </SelectItem>
                  <SelectItem value='rejected'>
                    {isZh ? '已拒绝' : 'Rejected'}
                  </SelectItem>
                </SelectContent>
              </Select>
              <ExportCsvButton url='/api/enterprise/export' requireRole={ROLE.SUPER_ADMIN} params={entStatusFilter ? { status: entStatusFilter } : undefined} />
            </div>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>{isZh ? '企业名称' : 'Name'}</TableHead>
                    <TableHead>{isZh ? '统一信用代码' : 'Credit Code'}</TableHead>
                    <TableHead>{isZh ? '联系人' : 'Contact'}</TableHead>
                    <TableHead>{isZh ? '状态' : 'Status'}</TableHead>
                    <TableHead>{isZh ? '会员等级' : 'Level'}</TableHead>
                    <TableHead>{isZh ? '操作' : 'Actions'}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {enterprises.map((ent) => {
                    const statusInfo =
                      ENTERPRISE_STATUS_MAP[ent.status] ||
                      ENTERPRISE_STATUS_MAP.pending
                    const levelConfig =
                      MEMBERSHIP_LEVEL_CONFIGS[ent.membership_level] ||
                      MEMBERSHIP_LEVEL_CONFIGS.gold
                    return (
                      <TableRow key={ent.id}>
                        <TableCell>{ent.id}</TableCell>
                        <TableCell className='font-medium'>{ent.name}</TableCell>
                        <TableCell className='font-mono text-sm'>
                          {ent.credit_code}
                        </TableCell>
                        <TableCell>{ent.contact_name || '-'}</TableCell>
                        <TableCell>
                          <Badge className={statusInfo.color}>
                            {isZh ? statusInfo.labelZh : statusInfo.label}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <span className={levelConfig.color}>
                            {isZh ? levelConfig.labelZh : levelConfig.label}
                          </span>
                        </TableCell>
                        <TableCell>
                          {ent.status === 'pending' && (
                            <div className='flex gap-1'>
                              <Button
                                size='sm'
                                variant='default'
                                onClick={() => handleApprove(ent.id)}
                              >
                                {isZh ? '通过' : 'Approve'}
                              </Button>
                              <Button
                                size='sm'
                                variant='destructive'
                                onClick={() => openRejectDialog(ent.id)}
                              >
                                {isZh ? '拒绝' : 'Reject'}
                              </Button>
                            </div>
                          )}
                          {ent.status === 'rejected' && ent.reject_reason && (
                            <span className='text-xs text-muted-foreground'>
                              {isZh ? '原因：' : 'Reason: '}
                              {ent.reject_reason}
                            </span>
                          )}
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </div>
      </SectionPageLayout.Content>

            {/* Reject Enterprise Dialog */}
      <Dialog open={rejectDialogOpen} onOpenChange={setRejectDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {isZh ? '拒绝企业认证' : 'Reject Enterprise'}
            </DialogTitle>
          </DialogHeader>
          <div className='space-y-4'>
            <div>
              <label htmlFor="reject-reason" className='text-sm font-medium'>
                {isZh ? '拒绝原因' : 'Reject Reason'}
              </label>
              <Input
                id="reject-reason"
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                placeholder={
                  isZh ? '请输入拒绝原因' : 'Enter the reason for rejection'
                }
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => setRejectDialogOpen(false)}
            >
              {isZh ? '取消' : 'Cancel'}
            </Button>
            <Button variant='destructive' onClick={handleReject}>
              {isZh ? '确认拒绝' : 'Confirm Reject'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SectionPageLayout>
  )
}