/*
Copyright (C) 2023-2026 OpenFastToken
*/
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { MEMBERSHIP_LEVEL_CONFIGS } from '../constants'
import type { MembershipInfo, MembershipLevel } from '../types'

interface MembershipUpgradeCardProps {
  membershipInfo: MembershipInfo | null
  loading: boolean
}

export function MembershipUpgradeCard({
  membershipInfo,
}: MembershipUpgradeCardProps) {
  const { i18n } = useTranslation()
  const isZh = i18n.language === 'zh' || i18n.language?.startsWith('zh')

  const levels: MembershipLevel[] = ['silver', 'gold', 'platinum']

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {isZh ? '会员等级说明' : 'Membership Levels'}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='grid gap-3 sm:grid-cols-3'>
          {levels.map((level) => {
            const config = MEMBERSHIP_LEVEL_CONFIGS[level]
            const isCurrent = membershipInfo?.membership_level === level

            return (
              <div
                key={level}
                className={`relative rounded-lg border-2 p-4 transition-all ${
                  isCurrent
                    ? `${config.bgColor} shadow-md`
                    : 'border-border bg-card'
                }`}
              >
                {isCurrent && (
                  <Badge
                    className='absolute -right-2 -top-2 bg-green-500 text-white'
                  >
                    {isZh ? '当前' : 'Current'}
                  </Badge>
                )}
                <div className='text-center'>
                  <span className='text-3xl'>{config.icon}</span>
                  <h3 className={`mt-2 text-lg font-semibold ${config.color}`}>
                    {isZh ? config.labelZh : config.label}
                  </h3>
                  <p className='mt-1 text-2xl font-bold'>{config.discountLabel}</p>
                  <p className='mt-2 text-xs text-muted-foreground'>
                    {isZh ? config.description : config.description}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}