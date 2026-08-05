/*
Copyright (C) 2023-2026 FastToken
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
  listInvitationCodes,
  createInvitationCodes,
  deleteInvitationCode,
  getInvitationCodeStats,
  listEnterprises,
  approveEnterprise,
  rejectEnterprise,
} from '../api'
import type {
  InvitationCode,
  InvitationCodeStats,
  Enterprise,
  MembershipLevel,
} from '../types'

export function EnterpriseAdmin() {
  const { i18n } = useTranslation()
  const isZh = i18n.language === 'zh' || i18n.language?.startsWith('zh')

  // Invitation codes state
  const [codes, setCodes] = useState<InvitationCode[]>([])
  const [, setCodeTotal] = useState(0)
  const [codePage, _setCodePage] = useState(1)
  const [codeStats, setCodeStats] = useState<InvitationCodeStats | null>(null)
  const [, setCodeLoading] = useState(false)

  // Create code dialog state
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [newCodeType, setNewCodeType] = useState<MembershipLevel>('gold')
  const [newCodeCount, setNewCodeCount] = useState(1)
  const [newCodeExpireDays, setNewCodeExpireDays] = useState(365)
  const [newCodeRemark, setNewCodeRemark] = useState('')
  const [creating, setCreating] = useState(false)

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

  // Fetch invitation codes
  const fetchCodes = useCallback(async () => {
    try {
      setCodeLoading(true)
      const res = await listInvitationCodes({
        page: codePage,
        page_size: 20,
      })
      if (res.success && res.data) {
        setCodes(res.data.codes)
        setCodeTotal(res.data.total)
      }
    } catch (error) {
      console.error('Failed to fetch invitation codes:', error)
    } finally {
      setCodeLoading(false)
    }
  }, [codePage])

  // Fetch code stats
  const fetchCodeStats = useCallback(async () => {
    try {
      const res = await getInvitationCodeStats()
      if (res.success && res.data) {
        setCodeStats(res.data)
      }
    } catch (error) {
      console.error('Failed to fetch invitation code stats:', error)
    }
  }, [])

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
    fetchCodes()
    fetchCodeStats()
  }, [fetchCodes, fetchCodeStats])

  useEffect(() => {
    fetchEnterprises()
  }, [fetchEnterprises])

  // Create invitation codes
  const handleCreateCodes = async () => {
    try {
      setCreating(true)
      if (typeof window !== 'undefined' && window.console) {
        window.console.log('[createCodes] submitting', { type: newCodeType, count: newCodeCount, expires_in: newCodeExpireDays, remark: newCodeRemark })
      }
      const res = await createInvitationCodes({
        type: newCodeType,
        count: newCodeCount,
        expires_in: newCodeExpireDays,
        remark: newCodeRemark,
      })
      if (typeof window !== 'undefined' && window.console) {
        window.console.log('[createCodes] response', res)
      }
      if (res.success) {
        toast.success(
          isZh
            ? `成功创建 ${newCodeCount} 个邀请码`
            : `Created ${newCodeCount} invitation codes`
        )
        setCreateDialogOpen(false)
        fetchCodes()
        fetchCodeStats()
      } else {
        const msg = res.message || (isZh ? '创建失败' : 'Failed to create')
        if (typeof window !== 'undefined' && window.console) {
          window.console.error('[createCodes] business error', res)
        }
        toast.error(msg)
      }
    } catch (error) {
      if (typeof window !== 'undefined' && window.console) {
        window.console.error('[createCodes] exception', error)
      }
      const msg = isZh ? '创建失败' : 'Failed to create'
      toast.error(msg)
      if (typeof window !== 'undefined' && window.alert) {
        window.alert((isZh ? '创建邀请码异常：' : 'Create invitation code error: ') + (error && error.message ? error.message : String(error)))
      }
    } finally {
      setCreating(false)
    }
  }

  // Delete invitation code
  const handleDeleteCode = async (id: number) => {
    try {
      const res = await deleteInvitationCode(id)
      if (res.success) {
        toast.success(isZh ? '已删除' : 'Deleted')
        fetchCodes()
        fetchCodeStats()
      } else {
        toast.error(res.message || (isZh ? '删除失败' : 'Failed to delete'))
      }
    } catch {
      toast.error(isZh ? '删除失败' : 'Failed to delete')
    }
  }

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
        {isZh ? '企业管理与邀请码' : 'Enterprise & Invitation Codes'}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
          {/* Stats Cards */}
          {codeStats && (
            <div className='grid gap-3 sm:grid-cols-4'>
              <Card>
                <CardContent className='p-4 text-center'>
                  <p className='text-2xl font-bold'>{codeStats.total}</p>
                  <p className='text-sm text-muted-foreground'>
                    {isZh ? '总邀请码' : 'Total Codes'}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className='p-4 text-center'>
                  <p className='text-2xl font-bold text-green-600'>
                    {codeStats.unused}
                  </p>
                  <p className='text-sm text-muted-foreground'>
                    {isZh ? '未使用' : 'Unused'}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className='p-4 text-center'>
                  <p className='text-2xl font-bold text-blue-600'>
                    {codeStats.used}
                  </p>
                  <p className='text-sm text-muted-foreground'>
                    {isZh ? '已使用' : 'Used'}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className='p-4 text-center'>
                  <p className='text-2xl font-bold text-red-600'>
                    {codeStats.expired}
                  </p>
                  <p className='text-sm text-muted-foreground'>
                    {isZh ? '已过期' : 'Expired'}
                  </p>
                </CardContent>
              </Card>
            </div>
          )}

          {/* Invitation Codes Section */}
          <Card>
            <CardHeader className='flex flex-row items-center justify-between'>
              <CardTitle>{isZh ? '邀请码管理' : 'Invitation Codes'}</CardTitle>
              <Button onClick={() => setCreateDialogOpen(true)}>
                + {isZh ? '创建邀请码' : 'Create Codes'}
              </Button>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>{isZh ? '邀请码' : 'Code'}</TableHead>
                    <TableHead>{isZh ? '等级' : 'Level'}</TableHead>
                    <TableHead>{isZh ? '状态' : 'Status'}</TableHead>
                    <TableHead>{isZh ? '备注' : 'Remark'}</TableHead>
                    <TableHead>{isZh ? '操作' : 'Actions'}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {codes.map((code) => {
                    const levelConfig =
                      MEMBERSHIP_LEVEL_CONFIGS[code.type] ||
                      MEMBERSHIP_LEVEL_CONFIGS.silver
                    return (
                      <TableRow key={code.id}>
                        <TableCell>{code.id}</TableCell>
                        <TableCell className='font-mono'>{code.code}</TableCell>
                        <TableCell>
                          <span className={levelConfig.color}>
                            {levelConfig.icon}{' '}
                            {isZh ? levelConfig.labelZh : levelConfig.label}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant={code.used_by > 0 ? 'secondary' : 'default'}
                          >
                            {code.used_by > 0
                              ? isZh
                                ? '已使用'
                                : 'Used'
                              : isZh
                                ? '未使用'
                                : 'Unused'}
                          </Badge>
                        </TableCell>
                        <TableCell className='max-w-[200px] truncate'>
                          {code.remark || '-'}
                        </TableCell>
                        <TableCell>
                          {code.used_by === 0 && (
                            <Button
                              variant='destructive'
                              size='sm'
                              onClick={() => handleDeleteCode(code.id)}
                            >
                              {isZh ? '删除' : 'Delete'}
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

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

      {/* Create Invitation Code Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {isZh ? '创建邀请码' : 'Create Invitation Codes'}
            </DialogTitle>
          </DialogHeader>
          <div className='space-y-4'>
            <div>
              <label htmlFor="membership-level" className='text-sm font-medium'>
                {isZh ? '会员等级' : 'Membership Level'}
              </label>
              <Select
                value={newCodeType}
                onValueChange={(v) => setNewCodeType(v as MembershipLevel)}
              >
                <SelectTrigger id="membership-level">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {(['gold', 'platinum'] as MembershipLevel[]).map((level) => {
                    const config = MEMBERSHIP_LEVEL_CONFIGS[level]
                    return (
                      <SelectItem key={level} value={level}>
                        {config.icon} {isZh ? config.labelZh : config.label} ({config.discountLabel})
                      </SelectItem>
                    )
                  })}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label htmlFor="code-count" className='text-sm font-medium'>
                {isZh ? '数量' : 'Count'}
              </label>
              <Input
                id="code-count"
                type='number'
                min={1}
                max={100}
                value={newCodeCount}
                onChange={(e) => setNewCodeCount(Number(e.target.value))}
              />
            </div>
            <div>
              <label htmlFor="code-validity" className='text-sm font-medium'>
                {isZh ? '有效期（天）' : 'Validity (days)'}
              </label>
              <Input
                id="code-validity"
                type='number'
                min={1}
                value={newCodeExpireDays}
                onChange={(e) => setNewCodeExpireDays(Number(e.target.value))}
              />
            </div>
            <div>
              <label htmlFor="code-remark" className='text-sm font-medium'>
                {isZh ? '备注' : 'Remark'}
              </label>
              <Input
                id="code-remark"
                value={newCodeRemark}
                onChange={(e) => setNewCodeRemark(e.target.value)}
                placeholder={isZh ? '可选备注' : 'Optional remark'}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => setCreateDialogOpen(false)}
            >
              {isZh ? '取消' : 'Cancel'}
            </Button>
            <Button onClick={handleCreateCodes} disabled={creating}>
              {creating
                ? isZh
                  ? '创建中...'
                  : 'Creating...'
                : isZh
                  ? '创建'
                  : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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