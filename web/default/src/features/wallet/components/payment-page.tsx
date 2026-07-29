/*
Copyright (C) 2023-2026 OpenFastToken
*/

import { useEffect, useRef, useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Smartphone, Copy, Check, CircleCheck, CircleX, Clock } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { toast } from 'sonner'
import { api } from '@/lib/api'
import { SiWechat } from 'react-icons/si'
import { PAYMENT_ICON_COLORS, PAYMENT_TYPES } from '../constants'

interface PaymentPageProps {
  tradeNo: string
  codeUrl: string
  amount: number
  topupAmount?: number
  usdExchangeRate?: number
}

type PollStatus = 'pending' | 'success' | 'closed' | 'error'

const MAX_POLL_SECONDS = 600

export function PaymentPage({
  tradeNo,
  codeUrl,
  amount,
}: PaymentPageProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const [pollStatus, setPollStatus] = useState<PollStatus>('pending')
  const [elapsed, setElapsed] = useState(0)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const copyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const abortedRef = useRef(false)

  const queryOrder = useCallback(async () => {
    if (abortedRef.current) return
    try {
      const res = await api.get('/api/user/wechat/query', {
        params: { trade_no: tradeNo },
        skipBusinessError: true,
      } as Record<string, unknown>)
      const status = res.data?.data?.status
      if (status === 'success') setPollStatus('success')
      else if (status === 'closed' || status === 'not_found') setPollStatus('closed')
    } catch {
      // keep pending
    }
  }, [tradeNo])

  useEffect(() => {
    abortedRef.current = false
    setPollStatus('pending')
    setElapsed(0)
    queryOrder()
    pollRef.current = setInterval(queryOrder, 3000)
    timerRef.current = setInterval(() => {
      setElapsed((e) => {
        if (e >= MAX_POLL_SECONDS) {
          if (pollRef.current) clearInterval(pollRef.current)
          if (timerRef.current) clearInterval(timerRef.current)
          setPollStatus('closed')
          return e
        }
        return e + 1
      })
    }, 1000)
    return () => {
      abortedRef.current = true
      if (pollRef.current) clearInterval(pollRef.current)
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [queryOrder])

  useEffect(() => {
    if (pollStatus === 'success') {
      if (pollRef.current) clearInterval(pollRef.current)
      if (timerRef.current) clearInterval(timerRef.current)
      const t = setTimeout(() => { window.close() }, 3000)
      return () => clearTimeout(t)
    }
  }, [pollStatus])

  useEffect(() => {
    return () => {
      if (copyTimerRef.current) clearTimeout(copyTimerRef.current)
    }
  }, [])

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(codeUrl)
      setCopied(true)
      toast.success(t('支付链接已复制'))
      if (copyTimerRef.current) clearTimeout(copyTimerRef.current)
      copyTimerRef.current = setTimeout(() => { setCopied(false); copyTimerRef.current = null }, 2000)
    } catch {
      toast.error(t('复制失败'))
    }
  }

  const formatTime = (totalSeconds: number) => {
    const m = Math.floor(totalSeconds / 60)
    const sec = totalSeconds % 60
    return String(m).padStart(2, '0') + ':' + String(sec).padStart(2, '0')
  }

  const remaining = MAX_POLL_SECONDS - elapsed
  const displayAmount = (amount || 0).toFixed(2)

  return (
    <div className='min-h-screen bg-gradient-to-b from-green-50 to-white flex items-center justify-center p-4'>
      <div className='w-full max-w-md'>
        <div className='text-center mb-6'>
          <div className='inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 mb-4'>
            <SiWechat className='h-8 w-8' style={{ color: PAYMENT_ICON_COLORS[PAYMENT_TYPES.WECHAT] }} />
          </div>
          <h1 className='text-2xl font-bold text-gray-900'>{t('微信扫码支付')}</h1>
          <p className='text-sm text-gray-500 mt-1'>FastToken</p>
        </div>
        <div className='bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden'>
          <div className='text-center pt-6 pb-4 border-b border-gray-50'>
            <p className='text-sm text-gray-400'>{t('充值金额')}</p>
            <p className='text-3xl font-bold text-gray-900 mt-1'>
              {'¥' + displayAmount}
            </p>
          </div>
          <div className='flex justify-center py-6'>
            <div className='rounded-xl border-2 border-gray-100 bg-white p-4 shadow-inner'>
              {codeUrl ? (
                <QRCodeSVG value={codeUrl} size={200} level='M' includeMargin imageSettings={{ src: '', height: 0, width: 0, excavate: false }} />
              ) : (
                <div className='flex h-[200px] w-[200px] items-center justify-center'><Loader2 className='h-8 w-8 animate-spin text-gray-300' /></div>
              )}
            </div>
          </div>
          <div className='px-6 pb-4'>
            {pollStatus === 'pending' && (
              <div className='text-center space-y-2'>
                <div className='flex items-center justify-center gap-1.5 text-sm text-gray-500'><Clock className='h-3.5 w-3.5' /><span>{t('等待支付 · 剩余 ') + formatTime(remaining)}</span></div>
                <p className='text-xs text-gray-400'>{t('订单号：') + tradeNo.slice(-8)}</p>
              </div>
            )}
            {pollStatus === 'success' && (
              <div className='text-center space-y-2 py-2'>
                <CircleCheck className='h-10 w-10 text-green-500 mx-auto' />
                <p className='text-green-600 font-semibold'>{t('支付成功！')}</p>
                <p className='text-xs text-gray-400'>{t('页面将自动关闭…')}</p>
              </div>
            )}
            {(pollStatus === 'closed' || pollStatus === 'error') && (
              <div className='text-center space-y-2 py-2'>
                <CircleX className='h-10 w-10 text-red-400 mx-auto' />
                <p className='text-red-500 font-medium'>{pollStatus === 'closed' ? t('订单已关闭') : t('查询失败')}</p>
                <p className='text-xs text-gray-400'>{t('请返回钱包页面重新发起支付')}</p>
              </div>
            )}
          </div>
          <div className='border-t border-gray-50 px-6 py-4'>
            <div className='flex items-center gap-2 text-xs text-gray-400 mb-3'><Smartphone className='h-3.5 w-3.5 flex-shrink-0' /><span>{t('请使用微信扫描二维码完成支付')}</span></div>
            <div className='flex items-center gap-2'>
              <code className='flex-1 truncate rounded-md bg-gray-50 px-2.5 py-1.5 text-[11px] text-gray-500 border border-gray-100'>{codeUrl}</code>
              <button onClick={handleCopy} className='flex-shrink-0 inline-flex items-center gap-1 px-2.5 py-1.5 text-xs rounded-md border border-gray-200 text-gray-600 hover:bg-gray-50 transition-colors'>
                {copied ? <Check className='h-3 w-3 text-green-500' /> : <Copy className='h-3 w-3' />}
                {copied ? t('已复制') : t('复制')}
              </button>
            </div>
          </div>
        </div>
        <p className='text-center text-xs text-gray-300 mt-4'>{t('安全支付由微信支付提供')}</p>
      </div>
    </div>
  )
}
