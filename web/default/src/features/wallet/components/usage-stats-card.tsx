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
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { BarChart3, Download, Gauge, Zap } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { formatQuota } from '@/lib/format'
import { getLogsSelfStat, getUserLogs, type LogSelfStatData } from '../api'

type Period = 'today' | 'week' | 'month'

interface UsageStatsCardProps {
  className?: string
}

function getPeriodTimestamps(period: Period): { start: number; end: number } {
  const now = new Date()
  const end = Math.floor(now.getTime() / 1000)

  const start = new Date(now)

  if (period === 'today') {
    start.setHours(0, 0, 0, 0)
  } else if (period === 'week') {
    // Start of current week (Monday)
    const day = start.getDay() || 7 // Sunday = 7
    start.setHours(0, 0, 0, 0)
    start.setDate(start.getDate() - day + 1)
  } else {
    // Start of current month
    start.setDate(1)
    start.setHours(0, 0, 0, 0)
  }

  return {
    start: Math.floor(start.getTime() / 1000),
    end,
  }
}

const periods: { key: Period; label: string }[] = [
  { key: 'today', label: 'Today' },
  { key: 'week', label: 'This Week' },
  { key: 'month', label: 'This Month' },
]

export function UsageStatsCard({ className }: UsageStatsCardProps) {
  const { t } = useTranslation()
  const [period, setPeriod] = useState<Period>('month')
  const [stat, setStat] = useState<LogSelfStatData | null>(null)
  const [loading, setLoading] = useState(true)
  const [exporting, setExporting] = useState(false)

  const fetchStats = useCallback(async (p: Period) => {
    setLoading(true)
    try {
      const { start, end } = getPeriodTimestamps(p)
      const res = await getLogsSelfStat({
        start_timestamp: start,
        end_timestamp: end,
      })
      if (res.success && res.data) {
        setStat(res.data)
      } else {
        setStat(null)
      }
    } catch {
      setStat(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStats(period)
  }, [period, fetchStats])

  const handlePeriodChange = (p: Period) => {
    setPeriod(p)
  }

  const handleExportCSV = async () => {
    setExporting(true)
    try {
      const { start, end } = getPeriodTimestamps(period)
      const res = await getUserLogs({
        start_timestamp: start,
        end_timestamp: end,
        page: 1,
        page_size: 10000,
      })

      if (!res.success || !res.data?.items) {
        return
      }

      const logs = res.data.items as Record<string, unknown>[]
      if (logs.length === 0) return

      // Build CSV header and rows
      const headers = [
        'ID',
        'Username',
        'Model',
        'Token Name',
        'Type',
        'Quota',
        'Created At',
      ]
      const csvRows = [headers.join(',')]

      for (const log of logs) {
        const row = [
          String(log.id ?? ''),
          String(log.username ?? ''),
          String(log.model_name ?? ''),
          String(log.token_name ?? ''),
          String(log.type ?? ''),
          String(log.quota ?? ''),
          String(log.created_at ?? ''),
        ].map((v) => `"${v.replace(/"/g, '""')}"`)
        csvRows.push(row.join(','))
      }

      const csvContent = csvRows.join('\n')
      const blob = new Blob(['\uFEFF' + csvContent], {
        type: 'text/csv;charset=utf-8',
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `usage-logs-${period}-${new Date().toISOString().slice(0, 10)}.csv`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch {
      // Silently fail export errors
    } finally {
      setExporting(false)
    }
  }

  if (loading) {
    return (
      <div className={`overflow-hidden rounded-lg border border-stone-card bg-card shadow-sm ${className ?? ''}`}>
        <div className='grid grid-cols-3 divide-x divide-stone-card/60'>
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className='px-3 py-3 sm:px-5 sm:py-4'>
              <Skeleton className='h-3.5 w-20' />
              <Skeleton className='mt-2 h-7 w-28' />
              <Skeleton className='mt-1.5 h-3.5 w-24' />
            </div>
          ))}
        </div>
      </div>
    )
  }

  const stats = [
    {
      label: t('Total Tokens'),
      value: formatQuota(stat?.quota ?? 0),
      icon: BarChart3,
    },
    {
      label: t('RPM Peak'),
      value: String(stat?.rpm ?? 0),
      icon: Gauge,
    },
    {
      label: t('TPM Peak'),
      value: formatQuota(stat?.tpm ?? 0),
      icon: Zap,
    },
  ]

  return (
    <div className={`overflow-hidden rounded-lg border border-stone-card bg-card shadow-sm ${className ?? ''}`}>
      {/* Stats row */}
      <div className='grid grid-cols-3 divide-x divide-stone-card/60'>
        {stats.map((item) => (
          <div key={item.label} className='px-3 py-3 sm:px-5 sm:py-4'>
            <div className='flex items-center gap-2'>
              <item.icon className='text-stone-muted size-3.5 shrink-0' />
              <div className='text-stone-muted truncate text-xs font-medium tracking-wider uppercase'>
                {item.label}
              </div>
            </div>
            <div className='text-stone-text mt-1.5 font-mono text-base font-bold tracking-tight break-all tabular-nums sm:mt-2 sm:text-2xl'>
              {item.value}
            </div>
          </div>
        ))}
      </div>

      {/* Footer: period selector + export */}
      <div className='flex items-center justify-between border-t border-stone-card/60 px-3 py-2 sm:px-5 sm:py-3'>
        <div className='flex items-center gap-1'>
          {periods.map((p) => (
            <button
              key={p.key}
              type='button'
              onClick={() => handlePeriodChange(p.key)}
              className={`rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
                period === p.key
                  ? 'bg-stone-text text-white'
                  : 'text-stone-muted hover:bg-stone-card hover:text-stone-text'
              }`}
            >
              {t(p.label)}
            </button>
          ))}
        </div>
        <Button
          variant='outline'
          size='sm'
          onClick={handleExportCSV}
          disabled={exporting}
          className='h-8 gap-1.5 border-stone-card bg-card text-stone-muted hover:bg-stone-card hover:text-stone-text text-xs'
        >
          <Download className='size-3.5' />
          {exporting ? t('Exporting...') : t('CSV')}
        </Button>
      </div>
    </div>
  )
}
