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
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ChevronDown,
  ChevronUp,
  Copy,
  ExternalLink,
  Clock,
  Zap,
  AlertTriangle,
  CheckCircle2,
  RefreshCw,
} from 'lucide-react'
import { api } from '@/lib/api'
import { cn } from '@/lib/utils'
import { formatTimestampToDate, formatLogQuota, formatTokens } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import { tryPrettyJson } from '@/lib/utils'

// ============================================================================
// Types
// ============================================================================

interface LogEntry {
  id: number
  user_id: number
  created_at: number
  type: number
  content: string
  username: string
  token_name: string
  model_name: string
  quota: number
  prompt_tokens: number
  completion_tokens: number
  use_time: number
  is_stream: boolean
  channel: number
  channel_name: string
  token_id: number
  group: string
  ip: string
  request_id: string
  upstream_request_id: string
  other: string
}

interface LogSearchResponse {
  success: boolean
  message: string
  data?: {
    items: LogEntry[]
    total: number
    page: number
    page_size: number
  }
}

// ============================================================================
// Helpers
// ============================================================================

const LOG_TYPE_CONSUME = 2
const LOG_TYPE_ERROR = 5

function isErrorLog(log: LogEntry): boolean {
  return log.type === LOG_TYPE_ERROR
}

function isConsumeLog(log: LogEntry): boolean {
  return log.type === LOG_TYPE_CONSUME
}

// ============================================================================
// Component
// ============================================================================

export function RecentRequestsCard() {
  const { t } = useTranslation()
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<number | null>(null)

  const fetchLogs = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get('/api/log/search', {
        params: {
          p: 1,
          page_size: 20,
        },
      })
      const data = res.data as LogSearchResponse
      if (data.success && data.data?.items) {
        setLogs(data.data.items)
      } else {
        setError(data.message || t('Failed to load request logs'))
      }
    } catch {
      setError(t('Failed to load request logs'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    fetchLogs()
  }, [fetchLogs])

  const toggleExpand = (id: number) => {
    setExpandedId((prev) => (prev === id ? null : id))
  }

  const handleCopyContent = (log: LogEntry) => {
    const payload = JSON.stringify(
      {
        model_name: log.model_name,
        content: log.content,
        prompt_tokens: log.prompt_tokens,
        completion_tokens: log.completion_tokens,
        is_stream: log.is_stream,
        request_id: log.request_id,
      },
      null,
      2
    )
    navigator.clipboard.writeText(payload).then(
      () => toast.success(t('Request payload copied to clipboard')),
      () => toast.error(t('Failed to copy'))
    )
  }

  // ==========================================================================
  // Loading State
  // ==========================================================================

  if (loading) {
    return (
      <Card className='bg-card border-border'>
        <CardHeader className='pb-2'>
          <CardTitle className='font-serif text-base'>
            {t('Recent API Requests')}
          </CardTitle>
        </CardHeader>
        <CardContent className='space-y-3'>
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className='flex items-center gap-3'>
              <Skeleton className='h-4 w-32' />
              <Skeleton className='h-4 w-24' />
              <Skeleton className='ml-auto h-4 w-20' />
            </div>
          ))}
        </CardContent>
      </Card>
    )
  }

  // ==========================================================================
  // Error / Empty States
  // ==========================================================================

  if (error) {
    return (
      <Card className='bg-card border-border'>
        <CardHeader className='pb-2'>
          <CardTitle className='font-serif text-base'>
            {t('Recent API Requests')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='flex flex-col items-center gap-3 py-6'>
            <AlertTriangle className='text-stone-muted size-8' />
            <p className='text-stone-muted text-sm'>{error}</p>
            <Button variant='outline' size='sm' onClick={fetchLogs}>
              <RefreshCw className='mr-1.5 size-3.5' />
              {t('Retry')}
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (logs.length === 0) {
    return (
      <Card className='bg-card border-border'>
        <CardHeader className='pb-2'>
          <CardTitle className='font-serif text-base'>
            {t('Recent API Requests')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className='text-stone-muted py-4 text-center text-sm'>
            {t('No API requests yet. Start using the API to see your request logs here.')}
          </p>
        </CardContent>
      </Card>
    )
  }

  // ==========================================================================
  // Main View
  // ==========================================================================

  return (
    <Card className='bg-card border-border'>
      <CardHeader className='flex flex-row items-center justify-between pb-2'>
        <CardTitle className='font-serif text-base'>
          {t('Recent API Requests')}
        </CardTitle>
        <Button variant='ghost' size='sm' onClick={fetchLogs} disabled={loading}>
          <RefreshCw className={cn('mr-1.5 size-3.5', loading && 'animate-spin')} />
          {t('Refresh')}
        </Button>
      </CardHeader>
      <CardContent className='p-0'>
        <div className='overflow-x-auto'>
          <table className='w-full'>
            <thead>
              <tr className='border-stone-100 border-b'>
                <th className='text-stone-muted px-4 py-2 text-left text-xs font-medium uppercase tracking-wider'>
                  {t('Time')}
                </th>
                <th className='text-stone-muted px-4 py-2 text-left text-xs font-medium uppercase tracking-wider'>
                  {t('Model')}
                </th>
                <th className='text-stone-muted px-4 py-2 text-left text-xs font-medium uppercase tracking-wider'>
                  {t('Type')}
                </th>
                <th className='text-stone-muted px-4 py-2 text-right text-xs font-medium uppercase tracking-wider'>
                  {t('Tokens')}
                </th>
                <th className='text-stone-muted px-4 py-2 text-right text-xs font-medium uppercase tracking-wider'>
                  {t('Cost')}
                </th>
                <th className='text-stone-muted w-10 px-2 py-2' />
              </tr>
            </thead>
            <tbody className='divide-stone-100 divide-y'>
              {logs.map((log) => {
                const expanded = expandedId === log.id
                return (
                  <tr key={log.id} className='group'>
                    <td
                      className={cn(
                        'cursor-pointer px-4 py-2.5',
                        expanded && 'bg-stone-50'
                      )}
                      onClick={() => toggleExpand(log.id)}
                    >
                      <div className='flex items-center gap-1.5'>
                        <Clock className='text-stone-muted/50 size-3 shrink-0' />
                        <span className='text-xs'>
                          {formatTimestampToDate(log.created_at)}
                        </span>
                        {expanded ? (
                          <ChevronUp className='text-stone-muted/50 ml-0.5 size-3' />
                        ) : (
                          <ChevronDown className='text-stone-muted/50 ml-0.5 size-3' />
                        )}
                      </div>
                    </td>
                    <td
                      className={cn(
                        'cursor-pointer px-4 py-2.5',
                        expanded && 'bg-stone-50'
                      )}
                      onClick={() => toggleExpand(log.id)}
                    >
                      <span className='text-xs font-medium'>
                        {log.model_name || '-'}
                      </span>
                    </td>
                    <td
                      className={cn(
                        'cursor-pointer px-4 py-2.5',
                        expanded && 'bg-stone-50'
                      )}
                      onClick={() => toggleExpand(log.id)}
                    >
                      <span className='inline-flex items-center gap-1'>
                        {isErrorLog(log) ? (
                          <AlertTriangle className='size-3 text-red-500' />
                        ) : isConsumeLog(log) ? (
                          <CheckCircle2 className='size-3 text-emerald-500' />
                        ) : (
                          <Zap className='text-stone-muted size-3' />
                        )}
                        <span
                          className={cn(
                            'text-xs',
                            isErrorLog(log) ? 'text-red-600' : 'text-stone-600'
                          )}
                        >
                          {isErrorLog(log) ? t('Error') : t('Success')}
                        </span>
                      </span>
                    </td>
                    <td
                      className={cn(
                        'cursor-pointer px-4 py-2.5 text-right',
                        expanded && 'bg-stone-50'
                      )}
                      onClick={() => toggleExpand(log.id)}
                    >
                      <span className='text-stone-muted text-xs tabular-nums'>
                        {formatTokens(
                          log.prompt_tokens + log.completion_tokens
                        )}
                      </span>
                    </td>
                    <td
                      className={cn(
                        'cursor-pointer px-4 py-2.5 text-right',
                        expanded && 'bg-stone-50'
                      )}
                      onClick={() => toggleExpand(log.id)}
                    >
                      <span className='text-stone-muted text-xs tabular-nums font-mono'>
                        {formatLogQuota(log.quota)}
                      </span>
                    </td>
                    <td
                      className={cn(
                        'px-2 py-2.5',
                        expanded && 'bg-stone-50'
                      )}
                    >
                      {isErrorLog(log) && (
                        <Button
                          variant='ghost'
                          size='icon'
                          className='size-7 opacity-0 group-hover:opacity-100 transition-opacity'
                          title={t('Copy request payload for replay')}
                          onClick={() => handleCopyContent(log)}
                        >
                          <Copy className='size-3.5' />
                        </Button>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        {/* Expanded detail row */}
        {expandedId != null && (() => {
          const log = logs.find((l) => l.id === expandedId)
          if (!log) return null
          return (
            <div className='bg-stone-50/80 border-stone-100 border-t px-4 py-3 space-y-3'>
              {/* Key info grid */}
              <div className='grid grid-cols-2 gap-x-6 gap-y-2 sm:grid-cols-4'>
                <div>
                  <div className='text-stone-muted text-[10px] uppercase tracking-wider'>
                    {t('Request ID')}
                  </div>
                  <div className='font-mono text-xs truncate'>
                    {log.request_id || '-'}
                  </div>
                </div>
                <div>
                  <div className='text-stone-muted text-[10px] uppercase tracking-wider'>
                    {t('Token')}
                  </div>
                  <div className='text-xs truncate'>
                    {log.token_name || '-'}
                  </div>
                </div>
                <div>
                  <div className='text-stone-muted text-[10px] uppercase tracking-wider'>
                    {t('Channel')}
                  </div>
                  <div className='text-xs truncate'>
                    {log.channel_name || '-'}
                  </div>
                </div>
                <div>
                  <div className='text-stone-muted text-[10px] uppercase tracking-wider'>
                    {t('Duration')}
                  </div>
                  <div className='text-xs tabular-nums'>
                    {log.use_time > 0 ? `${log.use_time}s` : '-'}
                  </div>
                </div>
              </div>

              {/* Content preview */}
              <div>
                <div className='text-stone-muted mb-1 text-[10px] uppercase tracking-wider'>
                  {t('Request Content')}
                </div>
                <pre className='bg-stone-100/80 max-h-40 overflow-auto rounded border border-border p-2 font-mono text-[11px] leading-relaxed whitespace-pre-wrap break-all'>
                  {tryPrettyJson(log.content) || '-'}
                </pre>
              </div>

              {/* Other metadata */}
              {log.other && log.other !== '{}' && (
                <div>
                  <div className='text-stone-muted mb-1 text-[10px] uppercase tracking-wider'>
                    {t('Additional Info')}
                  </div>
                  <pre className='bg-stone-100/80 max-h-40 overflow-auto rounded border border-border p-2 font-mono text-[11px] leading-relaxed whitespace-pre-wrap break-all'>
                    {tryPrettyJson(log.other)}
                  </pre>
                </div>
              )}

              {/* Action buttons */}
              <div className='flex items-center gap-2 pt-1'>
                {isErrorLog(log) && (
                  <Button
                    variant='outline'
                    size='sm'
                    onClick={() => handleCopyContent(log)}
                  >
                    <Copy className='mr-1.5 size-3.5' />
                    {t('Copy for Replay')}
                  </Button>
                )}
                {log.request_id && (
                  <Button
                    variant='ghost'
                    size='sm'
                    render={
                      <a
                        href={`/playground?request_id=${encodeURIComponent(log.request_id)}`}
                        target='_blank'
                        rel='noopener noreferrer'
                      >
                        <ExternalLink className='mr-1.5 size-3.5' />
                        {t('Open in Playground')}
                      </a>
                    }
                  />
                )}
              </div>
            </div>
          )
        })()}
      </CardContent>
    </Card>
  )
}
