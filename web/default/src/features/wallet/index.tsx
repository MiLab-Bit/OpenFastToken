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
import { useState, useEffect, useCallback, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { api, getSelf } from '@/lib/api'
import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'
import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AffiliateRewardsCard } from './components/affiliate-rewards-card'
import { BillingHistoryDialog } from './components/dialogs/billing-history-dialog'
import { PaymentConfirmDialog } from './components/dialogs/payment-confirm-dialog'
import { TransferDialog } from './components/dialogs/transfer-dialog'
import { WechatQrDialog } from './components/dialogs/wechat-qr-dialog'
import { RechargeFormCard } from './components/recharge-form-card'
import { SubscriptionPlansCard } from './components/subscription-plans-card'
import { WalletStatsCard } from './components/wallet-stats-card'
import { EnterpriseWalletCard } from './components/enterprise-wallet-card'
import { UsageStatsCard } from './components/usage-stats-card'
import { RecentRequestsCard } from './components/recent-requests-card'
import { MyGiftsCard } from './components/my-gifts-card'
import {
  useTopupInfo,
  usePayment,
  useAffiliate,
  useRedemption,
} from './hooks'
import {
  getDefaultPaymentType,
  getMinTopupAmount,
} from './lib'
import type {
  UserWalletData,
  PaymentMethod,
  PresetAmount,
} from './types'

interface WalletProps {
  initialShowHistory?: boolean
  pendingEvent?: string
}

export function Wallet(props: WalletProps) {
  const { t } = useTranslation()
  const [user, setUser] = useState<UserWalletData | null>(null)
  const [userLoading, setUserLoading] = useState(true)
  const [topupAmount, setTopupAmount] = useState(0)
  const [selectedPreset, setSelectedPreset] = useState<number | null>(null)
  const [selectedPaymentMethod, setSelectedPaymentMethod] =
    useState<PaymentMethod>()
  const [paymentLoading, setPaymentLoading] = useState<string | null>(null)
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false)
  const [transferDialogOpen, setTransferDialogOpen] = useState(false)
  const [billingDialogOpen, setBillingDialogOpen] = useState(false)
  const [redemptionCode, setRedemptionCode] = useState('')
  const [showSubscriptionPanel, setShowSubscriptionPanel] = useState(true)

  // WeChat QR dialog state
  const [wechatQrOpen, setWechatQrOpen] = useState(false)
  const [wechatCodeUrl] = useState('')
  const [wechatTradeNo] = useState('')

  // Event ticket dialog state (满额送福利)
  const [pendingAlipayVerify, setPendingAlipayVerify] = useState<{ tradeNo: string } | null>(null)

  const { status } = useStatus()
  const { currency } = useSystemConfig()
  const { topupInfo, presetAmounts, loading: topupLoading } = useTopupInfo()

  const eventGift = topupInfo?.recharge_gift?.gift
  const eventGiftThreshold =
    eventGift?.enabled && eventGift?.threshold ? eventGift.threshold : 0

  const effectiveUsdExchangeRate = useMemo(() => {
    return currency?.quotaDisplayType === 'USD'
      ? 1
      : currency?.usdExchangeRate || 1
  }, [currency?.quotaDisplayType, currency?.usdExchangeRate])

  const {
    amount: paymentAmount,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
  } = usePayment()

  const {
    affiliateLink,
    loading: affiliateLoading,
    transferQuota,
    transferring,
  } = useAffiliate()
  const { redeeming, redeemCode } = useRedemption()

  const fetchUser = useCallback(async () => {
    try {
      setUserLoading(true)
      const response = await getSelf()
      if (response.success && response.data) {
        setUser(response.data as UserWalletData)
      }
    } catch (error) {
       
      console.error('Failed to fetch user data:', error)
    } finally {
      setUserLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchUser()
  }, [fetchUser])

  useEffect(() => {
    if (props.initialShowHistory) {
      setBillingDialogOpen(true)
      // Only replace state if there are query params to clean up
      // Use setTimeout to defer and avoid browser warning about history manipulation
      if (window.location.search) {
        setTimeout(() => {
          window.history.replaceState({}, '', window.location.pathname)
        }, 0)
      }
    }

    // Handle Alipay return: defer ticket dialog until payment is verified
    if (props.pendingEvent) {
      try {
        const { tradeNo } = JSON.parse(props.pendingEvent)
        setPendingAlipayVerify({ tradeNo })
      } catch { /* ignore */ }
    }
  }, [props.initialShowHistory])

  useEffect(() => {
    if (topupInfo && topupAmount === 0) {
      const minTopup = getMinTopupAmount(topupInfo)
      setTopupAmount(minTopup)
      const defaultPaymentType = getDefaultPaymentType(topupInfo)
      calculatePaymentAmount(minTopup, defaultPaymentType)
    }
  }, [topupInfo, topupAmount, calculatePaymentAmount])

  const getCurrentPaymentType = useCallback(() => {
    return selectedPaymentMethod?.type || getDefaultPaymentType(topupInfo)
  }, [selectedPaymentMethod, topupInfo])

  const handleSelectPreset = (preset: PresetAmount) => {
    setTopupAmount(preset.value)
    setSelectedPreset(preset.value)
    calculatePaymentAmount(preset.value, getCurrentPaymentType())
  }

  const handleTopupAmountChange = (amount: number) => {
    setTopupAmount(amount)
    setSelectedPreset(null)
    calculatePaymentAmount(amount, getCurrentPaymentType())
  }

  const handlePaymentMethodSelect = async (method: PaymentMethod) => {
    setSelectedPaymentMethod(method)
    setPaymentLoading(method.type)

    try {
      const minTopup = getMinTopupAmount(topupInfo)
      if (topupAmount < minTopup) return

      await calculatePaymentAmount(topupAmount, method.type)
      setConfirmDialogOpen(true)
    } finally {
      setPaymentLoading(null)
    }
  }

  // Unified payment confirm — routes to Alipay redirect or WeChat QR dialog
  const handlePaymentConfirm = async () => {
    if (!selectedPaymentMethod) return

    const result = await processPayment(topupAmount, selectedPaymentMethod.type)

    if (result.success) {
      setConfirmDialogOpen(false)

      if (result.type === 'wechat' && result.codeUrl) {
        // Redirect to payment page (Alipay-style, opens in new tab)
        const payUrl = '/pay/' + (result.tradeNo || '') + '?amount=' + (paymentAmount || 0) + '&codeUrl=' + encodeURIComponent(result.codeUrl) + '&topupAmount=' + topupAmount + '&usdExchangeRate=' + effectiveUsdExchangeRate
        window.open(payUrl, '_blank')
      }
      // For Alipay, the form submit has already redirected the browser
    }
  }

  // WeChat QR payment completed
  const handleWechatPaymentComplete = async () => {
    setWechatQrOpen(false)
    await fetchUser()
    // Check if this recharge met the gift threshold
    if (eventGiftThreshold > 0 && topupAmount >= eventGiftThreshold) {
      setEventTradeNo(wechatTradeNo)
    }
  }

  // Verify Alipay payment success before showing event ticket dialog
  useEffect(() => {
    if (!pendingAlipayVerify) return
    const verify = async () => {
      try {
        // Query the order status — works for both WeChat and Alipay topup records
        const res = await api.get('/api/user/wechat/query', {
          params: { trade_no: pendingAlipayVerify.tradeNo },
          skipBusinessError: true,
        } as Record<string, unknown>)
        const status = res.data?.data?.status
        if (status === 'success') {
          setEventTradeNo(pendingAlipayVerify.tradeNo)
        }
      } catch {
        // Payment not verified — silently skip ticket dialog
      } finally {
        setPendingAlipayVerify(null)
      }
    }
    verify()
  }, [pendingAlipayVerify])

  const handleRedeem = async () => {
    if (!redemptionCode) return
    const success = await redeemCode(redemptionCode)
    if (success) {
      setRedemptionCode('')
      await fetchUser()
    }
  }

  const handleTransfer = async (amount: number) => {
    const success = await transferQuota(amount)
    if (success) await fetchUser()
    return success
  }

  const getBonusCredit = useCallback(() => {
    return topupInfo?.bonus_credit?.[topupAmount] || 0
  }, [topupInfo, topupAmount])

  const handleSubscriptionAvailabilityChange = useCallback(
    (available: boolean) => {
      setShowSubscriptionPanel(available)
    },
    []
  )

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Wallet')}</SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
            <WalletStatsCard user={user} loading={userLoading} />

            {/* Enterprise wallet card — shown for enterprise members */}
            <EnterpriseWalletCard enterpriseId={user?.enterprise_id} />

            {/* Low balance warning banner */}
            {user && user.quota > 0 && user.quota < 100000 && (
              <Alert variant='default' className='bg-amber-50 border-amber-200'>
                <AlertDescription className='text-amber-800'>
                  {t('Your balance is only {{quota}} tokens remaining. We recommend topping up soon to avoid service interruption.', {
                    quota: (user.quota).toLocaleString(),
                  })}
                </AlertDescription>
              </Alert>
            )}

            <UsageStatsCard />

            <div
              className={
                showSubscriptionPanel
                  ? 'grid gap-4 xl:grid-cols-[minmax(0,1.05fr)_minmax(360px,0.95fr)] xl:items-start'
                  : 'grid gap-4'
              }
            >
              <div id='wallet-add-funds' className='scroll-mt-4'>
                <RechargeFormCard
                  topupInfo={topupInfo}
                  presetAmounts={presetAmounts}
                  selectedPreset={selectedPreset}
                  onSelectPreset={handleSelectPreset}
                  topupAmount={topupAmount}
                  onTopupAmountChange={handleTopupAmountChange}
                  paymentAmount={paymentAmount}
                  calculating={calculating}
                  selectedPaymentMethod={selectedPaymentMethod}
                  onPaymentMethodSelect={handlePaymentMethodSelect}
                  paymentLoading={paymentLoading}
                  redemptionCode={redemptionCode}
                  onRedemptionCodeChange={setRedemptionCode}
                  onRedeem={handleRedeem}
                  redeeming={redeeming}
                  topupLink={topupInfo?.topup_link}
                  loading={topupLoading}
                  priceRatio={(status?.price as number) || 1}
                  usdExchangeRate={effectiveUsdExchangeRate}
                  onOpenBilling={() => setBillingDialogOpen(true)}
                />
              </div>

              <SubscriptionPlansCard
                topupInfo={topupInfo}
                onAvailabilityChange={handleSubscriptionAvailabilityChange}
              />
            </div>

            <AffiliateRewardsCard
              user={user}
              affiliateLink={affiliateLink}
              onTransfer={() => setTransferDialogOpen(true)}
              complianceConfirmed={
                topupInfo?.payment_compliance_confirmed !== false
              }
              loading={affiliateLoading}
            />

            <MyGiftsCard />

            <RecentRequestsCard />
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <PaymentConfirmDialog
        open={confirmDialogOpen}
        onOpenChange={setConfirmDialogOpen}
        onConfirm={handlePaymentConfirm}
        topupAmount={topupAmount}
        paymentAmount={paymentAmount}
        paymentMethod={selectedPaymentMethod}
        calculating={calculating}
        processing={processing}
        bonusCredit={getBonusCredit()}
        usdExchangeRate={effectiveUsdExchangeRate}
      />

      {/* WeChat QR Code Payment Dialog */}
      <WechatQrDialog
        open={wechatQrOpen}
        onOpenChange={setWechatQrOpen}
        codeUrl={wechatCodeUrl}
        tradeNo={wechatTradeNo}
        amount={paymentAmount}
        topupAmount={topupAmount}
        usdExchangeRate={effectiveUsdExchangeRate}
        onPaymentComplete={handleWechatPaymentComplete}
      />

      <TransferDialog
        open={transferDialogOpen}
        onOpenChange={setTransferDialogOpen}
        onConfirm={handleTransfer}
        availableQuota={user?.aff_quota ?? 0}
        transferring={transferring}
      />

      <BillingHistoryDialog
        open={billingDialogOpen}
        onOpenChange={setBillingDialogOpen}
      />
    </>
  )
}
