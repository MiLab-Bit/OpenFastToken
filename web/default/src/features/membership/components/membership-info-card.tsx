/*
Copyright (C) 2023-2026 OpenFastToken
*/
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  MEMBERSHIP_LEVEL_CONFIGS,
  formatExpireTime,
  formatDiscountRate,
} from '../constants'
import type { MembershipInfo } from '../types'

interface MembershipInfoCardProps {
  membershipInfo: MembershipInfo | null
  loading: boolean
  onRefresh: () => void
}

export function MembershipInfoCard({
  membershipInfo,
  loading,
  onRefresh,
}: MembershipInfoCardProps) {
  const { i18n } = useTranslation()

  if (loading || !membershipInfo) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className='h-6 w-48' />
        </CardHeader>
        <CardContent className='space-y-4'>
          <Skeleton className='h-20 w-full' />
          <Skeleton className='h-4 w-32' />
        </CardContent>
      </Card>
    )
  }

  const levelConfig =
    MEMBERSHIP_LEVEL_CONFIGS[membershipInfo.membership_level] ||
    MEMBERSHIP_LEVEL_CONFIGS.silver
  const isZh = i18n.language === 'zh' || i18n.language?.startsWith('zh')

  return (
    <Card className={`border-2 ${levelConfig.bgColor}`}>
      <CardHeader className='flex flex-row items-center justify-between'>
        <CardTitle className='flex items-center gap-2'>
          <span className='text-2xl'>{levelConfig.icon}</span>
          <span className={levelConfig.color}>
            {isZh ? levelConfig.labelZh : levelConfig.label}
          </span>
        </CardTitle>
        <div className='flex items-center gap-2'>
          <Badge
            variant={membershipInfo.is_active ? 'default' : 'secondary'}
            className={membershipInfo.is_active ? 'bg-green-500' : 'bg-gray-400'}
          >
            {membershipInfo.is_active
              ? isZh
                ? '有效'
                : 'Active'
              : isZh
                ? '已过期'
                : 'Expired'}
          </Badge>
          <Button variant='ghost' size='sm' onClick={onRefresh}>
            ↻
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className='grid gap-4 sm:grid-cols-3'>
          <div className='space-y-1'>
            <p className='text-sm text-muted-foreground'>
              {isZh ? '折扣率' : 'Discount Rate'}
            </p>
            <p className={`text-2xl font-bold ${levelConfig.color}`}>
              {formatDiscountRate(membershipInfo.discount_rate)}
            </p>
          </div>
          <div className='space-y-1'>
            <p className='text-sm text-muted-foreground'>
              {isZh ? '有效期至' : 'Valid Until'}
            </p>
            <p className='text-lg font-medium'>
              {formatExpireTime(membershipInfo.expire_time)}
            </p>
          </div>
          <div className='space-y-1'>
            <p className='text-sm text-muted-foreground'>
              {isZh ? '会员等级' : 'Level'}
            </p>
            <p className={`text-lg font-medium ${levelConfig.color}`}>
              {isZh ? levelConfig.labelZh : levelConfig.label}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}