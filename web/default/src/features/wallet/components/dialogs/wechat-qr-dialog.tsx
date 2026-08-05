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
import { useEffect, useRef, useState, useCallback } from 'react'
import { Loader2, Smartphone, Copy, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { QRCodeSVG } from 'qrcode.react'
import { toast } from 'sonner'
import { formatLocalCurrencyAmount } from '@/lib/currency'
import { api } from '@/lib/api'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { SiWechat } from 'react-icons/si'
import { PAYMENT_ICON_COLORS, PAYMENT_TYPES } from '../../constants'

interface WechatQrDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  codeUrl: string
  tradeNo: string
  amount: number
  topupAmount: number
  usdExchangeRate?: number
  onPaymentComplete?: () => void
}

export function WechatQrDialog({
  open,
  onOpenChange,
  codeUrl,
  tradeNo,
  amount,
  topupAmount,
  usdExchangeRate = 1,
  onPaymentComplete,
}: WechatQrDialogProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const [pollStatus, setPollStatus] = useState<'pending' | 'success' | 'closed' | 'error'>('pending')
  const [_pollCount, setPollCount] = useState(0)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const pollAbortedRef = useRef(false)
  const copyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const MAX_POLL_COUNT = 200 // ~10 minutes at 3s intervals

  // Query order status from the backend
  const queryOrderStatus = useCallback(async () => {
    if (pollAbortedRef.current) return
    try {
      const res = await api.get('/api/user/wechat/query', {
        params: { trade_no: tradeNo },
        skipBusinessError: true,
      } as Record<string, unknown>)
      const status = res.data?.data?.status
      if (status) {
        if (status === 'success') {
          setPollStatus('success')
        } else if (status === 'closed' || status === 'not_found') {
          setPollStatus('closed')
        } else {
          // 'pending' or 'unknown' — keep polling
          setPollStatus('pending')
        }
      }
    } catch {
      setPollStatus('error')
    }
  }, [tradeNo])

  // Periodic polling every 3 seconds
  useEffect(() => {
    if (open && tradeNo) {
      setPollCount(0)
      setPollStatus('pending')
      pollAbortedRef.current = false

      // Immediate first query
      queryOrderStatus()

      pollRef.current = setInterval(() => {
        setPollCount((c) => {
          if (c >= MAX_POLL_COUNT) {
            if (pollRef.current) {
              clearInterval(pollRef.current)
              pollRef.current = null
            }
            setPollStatus('closed')
            return c
          }
          queryOrderStatus()
          return c + 1
        })
      }, 3000)
    }
    return () => {
      pollAbortedRef.current = true
      if (pollRef.current) {
        clearInterval(pollRef.current)
        pollRef.current = null
      }
    }
  }, [open, tradeNo, queryOrderStatus])

  // Handle pollStatus changes for side effects
  useEffect(() => {
    if (pollStatus === 'success') {
      // Stop polling
      if (pollRef.current) {
        clearInterval(pollRef.current)
        pollRef.current = null
      }
      toast.success(t('Payment successful'))
      onPaymentComplete?.()
      // Delay closing dialog to show success state
      const timer = setTimeout(() => {
        onOpenChange(false)
      }, 2000)
      return () => clearTimeout(timer)
    } else if (pollStatus === 'closed') {
      if (pollRef.current) {
        clearInterval(pollRef.current)
        pollRef.current = null
      }
      toast.error(t('Order has been closed'))
    }
  }, [pollStatus, onPaymentComplete, onOpenChange, t])

  // Cleanup timers on unmount
  useEffect(() => {
    return () => {
      if (copyTimerRef.current) {
        clearTimeout(copyTimerRef.current)
        copyTimerRef.current = null
      }
    }
  }, [])

  const handleCopyUrl = async () => {
    try {
      await navigator.clipboard.writeText(codeUrl)
      setCopied(true)
      toast.success(t('Payment link copied'))
      // Clear any existing timer
      if (copyTimerRef.current) {
        clearTimeout(copyTimerRef.current)
      }
      copyTimerRef.current = setTimeout(() => {
        setCopied(false)
        copyTimerRef.current = null
      }, 2000)
    } catch {
      toast.error(t('Failed to copy'))
    }
  }

  const handleDone = () => {
    // Manual check triggered by user
    queryOrderStatus()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-sm:w-[calc(100vw-1.5rem)] sm:max-w-sm'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2 text-xl font-semibold'>
            <SiWechat
              className='h-6 w-6'
              style={{ color: PAYMENT_ICON_COLORS[PAYMENT_TYPES.WECHAT] }}
            />
            {t('WeChat Scan to Pay')}
          </DialogTitle>
          <DialogDescription>
            {t('Scan the QR code with WeChat to complete payment')}
          </DialogDescription>
        </DialogHeader>

        <div className='flex flex-col items-center space-y-4 py-4'>
          {/* 金额展示 */}
          <div className='text-center'>
            <p className='text-muted-foreground text-sm'>
              {t('Topup Amount')}:{' '}
              {formatLocalCurrencyAmount(topupAmount * usdExchangeRate, {
                digitsLarge: 2,
                digitsSmall: 2,
                abbreviate: false,
              })}
            </p>
            <p className='text-2xl font-bold text-green-600'>
              ¥{(amount || 0).toFixed(2)}
            </p>
          </div>

          {/* 二维码 */}
          <div className='rounded-xl border bg-white p-4 shadow-sm'>
            {codeUrl ? (
              <QRCodeSVG
                value={codeUrl}
                size={200}
                level='M'
                includeMargin
                imageSettings={{
                  src: '',
                  height: 0,
                  width: 0,
                  excavate: false,
                }}
              />
            ) : (
              <div className='flex h-[200px] w-[200px] items-center justify-center'>
                <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
              </div>
            )}
          </div>

          {/* 提示 */}
          <div className='flex items-center gap-2 text-sm text-muted-foreground'>
            <Smartphone className='h-4 w-4' />
            <span>{t('Open WeChat app and scan the QR code')}</span>
          </div>

          <Separator />

          {/* 备用操作 */}
          <div className='w-full space-y-2 text-xs text-muted-foreground'>
            <p className='text-center'>
              {t('If scan doesn\'t work, copy the payment link:')}
            </p>
            <div className='flex items-center gap-2'>
              <code className='flex-1 truncate rounded bg-muted px-2 py-1 text-[11px]'>
                {codeUrl}
              </code>
              <Button
                variant='outline'
                size='sm'
                className='h-7 gap-1 px-2 text-xs'
                onClick={handleCopyUrl}
              >
                {copied ? (
                  <Check className='h-3 w-3 text-green-500' />
                ) : (
                  <Copy className='h-3 w-3' />
                )}
                {copied ? t('Copied') : t('Copy')}
              </Button>
            </div>
          </div>

          {/* 轮询状态 */}
          <div className='text-center'>
            {pollStatus === 'pending' && (
              <p className='text-muted-foreground text-xs'>
                {t('Awaiting payment...')} {t('Order')}: {tradeNo.slice(-8)}
              </p>
            )}
            {pollStatus === 'success' && (
              <p className='text-green-600 text-sm font-medium'>
                {t('Payment successful!')}
              </p>
            )}
            {pollStatus === 'closed' && (
              <p className='text-destructive text-xs'>
                {t('Order has been closed')}
              </p>
            )}
            {pollStatus === 'error' && (
              <p className='text-destructive text-xs'>
                {t('Failed to query order status')}
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={handleDone} className='w-full'>
            {t('I\'ve Completed Payment')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
