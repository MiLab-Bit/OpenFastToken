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
import { useState, useCallback } from 'react'
import i18next from 'i18next'
import { toast } from 'sonner'
import {
  calculateAlipayAmount,
  calculateWechatAmount,
  requestAlipayPayment,
  requestWechatPayment,
  isApiSuccess,
} from '../api'
import {
  isAlipayPayment,
  isWechatPayment,
  submitPaymentForm,
} from '../lib'
import type { WechatPaymentResponse } from '../types'

// ============================================================================
// Payment Hook
// ============================================================================

interface PaymentResult {
  /** Whether payment was initiated successfully */
  success: boolean
  /** Payment type for routing */
  type: 'alipay' | 'wechat' | null
  /** WeChat: code_url for QR rendering */
  codeUrl?: string
  /** WeChat: trade number for tracking */
  tradeNo?: string
}

export function usePayment() {
  const [amount, setAmount] = useState<number>(0)
  const [calculating, setCalculating] = useState(false)
  const [processing, setProcessing] = useState(false)

  // Calculate payment amount
  const calculatePaymentAmount = useCallback(
    async (topupAmount: number, paymentType: string) => {
      try {
        setCalculating(true)

        const isAlipay = isAlipayPayment(paymentType)
        const isWechat = isWechatPayment(paymentType)
        const response = isAlipay
          ? await calculateAlipayAmount({ amount: topupAmount })
          : isWechat
            ? await calculateWechatAmount({ amount: topupAmount })
            : null

        if (response && isApiSuccess(response) && response.data) {
          // Backend returns different formats:
          // - Alipay: { data: "1.00" } (string number directly)
          // - WeChat: { data: { money: "1.00" } } (object with money field)
          const rawData: unknown = response.data
          const calculatedAmount = typeof rawData === 'object' && rawData !== null && 'money' in rawData
            ? (Number((rawData as Record<string, unknown>).money) || 0)
            : (Number(rawData) || 0)
          setAmount(calculatedAmount)
          return calculatedAmount
        }

        setAmount(0)
        return 0
      } catch {
        setAmount(0)
        return 0
      } finally {
        setCalculating(false)
      }
    },
    []
  )

  // Process payment — returns result with type routing info
  const processPayment = useCallback(
    async (topupAmount: number, paymentType: string): Promise<PaymentResult> => {
      const failResult: PaymentResult = { success: false, type: null }

      try {
        setProcessing(true)

        const isAlipay = isAlipayPayment(paymentType)
        const amount = Math.floor(topupAmount)

        if (isAlipay) {
          const response = await requestAlipayPayment({
            amount,
            payment_method: paymentType as 'alipay',
          })

          if (!isApiSuccess(response)) {
            toast.error(response.message || i18next.t('Payment request failed'))
            return failResult
          }

          // Backend returns { message: "success", data: { pay_link: "..." } }
          const responseData = response.data as Record<string, unknown>
          const url = (responseData?.pay_link || response.url || '') as string
          if (url) {
            submitPaymentForm(url, {})
            toast.success(i18next.t('Redirecting to Alipay...'))
            return { success: true, type: 'alipay' }
          }
          toast.error(i18next.t('No payment URL returned'))
          return failResult
        }

        // WeChat Pay Native — return code_url for QR dialog
        const response: WechatPaymentResponse = await requestWechatPayment({
          amount,
          payment_method: 'wxpay',
        })

        if (!isApiSuccess(response)) {
          toast.error(response.message || i18next.t('Payment request failed'))
          return failResult
        }

        if (response.data?.code_url) {
          return {
            success: true,
            type: 'wechat',
            codeUrl: response.data.code_url,
            tradeNo: response.data.trade_no,
          }
        }

        return failResult
      } catch {
        toast.error(i18next.t('Payment request failed'))
        return failResult
      } finally {
        setProcessing(false)
      }
    },
    []
  )

  return {
    amount,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
    setAmount,
  }
}
