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
along with this program see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact hello@fasttoken.example.com
*/
import { useTranslation } from 'react-i18next'
import { Gift } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { formatTimestampToDate } from '@/lib/format'
import { useGifts, type UserGiftStatus } from '../hooks/use-gifts'

const STATUS_STYLES: Record<UserGiftStatus, string> = {
  active: 'bg-green-100 text-green-700',
  used: 'bg-muted text-muted-foreground',
  expired: 'bg-amber-100 text-amber-700',
}

export function MyGiftsCard() {
  const { t } = useTranslation()
  const { gifts, loading } = useGifts()

  return (
    <Card>
      <CardHeader className='pb-3'>
        <CardTitle className='flex items-center gap-2 text-base'>
          <Gift className='h-4 w-4' />
          {t('My Gifts')}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className='space-y-2'>
            {Array.from({ length: 2 }).map((_, i) => (
              <Skeleton key={i} className='h-16 w-full' />
            ))}
          </div>
        ) : gifts.length === 0 ? (
          <p className='py-4 text-center text-sm text-muted-foreground'>
            {t('No gifts yet')}
          </p>
        ) : (
          <div className='space-y-2'>
            {gifts.map((g) => (
              <div
                key={g.id}
                className='flex items-start justify-between gap-3 rounded-lg border border-border bg-stone-50 p-3'
              >
                <div className='min-w-0'>
                  <div className='font-medium'>{g.gift_name}</div>
                  {g.description && (
                    <div className='mt-0.5 truncate text-xs text-muted-foreground'>
                      {g.description}
                    </div>
                  )}
                  <div className='mt-1 text-xs text-muted-foreground'>
                    {formatTimestampToDate(g.create_time)}
                  </div>
                </div>
                <Badge className={cn('border-0 shrink-0', STATUS_STYLES[g.status])}>
                  {t(`${g.status} Gifts`)}
                </Badge>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
