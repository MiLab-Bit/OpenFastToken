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
import { useState, useEffect } from 'react'
import { Gift, ExternalLink, Loader2, Receipt, WalletCards, Smartphone, PartyPopper } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { SiAlipay, SiWechat } from 'react-icons/si'
import { formatNumber } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { TitledCard } from '@/components/ui/titled-card'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  formatCurrency,
  getMinTopupAmount,
  calculatePresetPricing,
} from '../lib'
import { PAYMENT_TYPES, PAYMENT_ICON_COLORS } from '../constants'
import type {
  PaymentMethod,
  PresetAmount,
  RechargeGiftInfo,
  TopupInfo,
} from '../types'

interface RechargeFormCardProps {
  topupInfo: TopupInfo | null
  presetAmounts: PresetAmount[]
  selectedPreset: number | null
  onSelectPreset: (preset: PresetAmount) => void
  topupAmount: number
  onTopupAmountChange: (amount: number) => void
  paymentAmount: number
  calculating: boolean
  selectedPaymentMethod?: PaymentMethod
  onPaymentMethodSelect: (method: PaymentMethod) => void
  paymentLoading: string | null
  redemptionCode: string
  onRedemptionCodeChange: (code: string) => void
  onRedeem: () => void
  redeeming: boolean
  topupLink?: string
  loading?: boolean
  priceRatio?: number
  usdExchangeRate?: number
  onOpenBilling?: () => void
}

export function RechargeFormCard({
  topupInfo,
  presetAmounts,
  selectedPreset,
  onSelectPreset,
  topupAmount,
  onTopupAmountChange,
  paymentAmount,
  calculating,
  selectedPaymentMethod,
  onPaymentMethodSelect,
  paymentLoading,
  redemptionCode,
  onRedemptionCodeChange,
  onRedeem,
  redeeming,
  topupLink,
  loading,
  priceRatio = 1,
  usdExchangeRate = 1,
  onOpenBilling,
}: RechargeFormCardProps) {
  const { t } = useTranslation()
  const [localAmount, setLocalAmount] = useState(topupAmount.toString())

  useEffect(() => {
    setLocalAmount(topupAmount.toString())
  }, [topupAmount])

  const handleAmountChange = (value: string) => {
    setLocalAmount(value)
    const numValue = parseInt(value) || 0
    if (numValue >= 0) {
      onTopupAmountChange(numValue)
    }
  }

  const hasStandardPaymentMethods =
    Array.isArray(topupInfo?.pay_methods) && topupInfo.pay_methods.length > 0
  const hasAnyTopup = hasStandardPaymentMethods
  const minTopup = getMinTopupAmount(topupInfo)
  const redemptionEnabled = topupInfo?.enable_redemption !== false

  const gift: RechargeGiftInfo['gift'] | undefined =
    topupInfo?.recharge_gift?.gift
  const giftThreshold = gift?.enabled ? (gift?.threshold || 0) : 0

  if (loading) {
    return (
      <Card className='gap-0 overflow-hidden py-0'>
        <CardHeader className='border-b p-3 !pb-3 sm:p-5 sm:!pb-5'>
          <Skeleton className='h-6 w-32' />
          <Skeleton className='mt-2 h-4 w-48' />
        </CardHeader>
        <CardContent className='space-y-4 p-3 sm:space-y-6 sm:p-5'>
          <div className='space-y-4 sm:space-y-6'>
            {/* Preset Amounts Skeleton */}
            <div className='space-y-3'>
              <Skeleton className='h-3 w-16' />
              <div className='grid grid-cols-2 gap-3 sm:grid-cols-4'>
                {Array.from({ length: 8 }).map((_, i) => (
                  <Skeleton key={i} className='h-[72px] rounded-lg' />
                ))}
              </div>
            </div>

            {/* Custom Amount Input Skeleton */}
            <div className='space-y-3'>
              <Skeleton className='h-3 w-28' />
              <Skeleton className='h-[42px] w-full' />
            </div>

            {/* Payment Methods Skeleton */}
            <div className='space-y-3'>
              <Skeleton className='h-3 w-32' />
              <div className='flex flex-wrap gap-3'>
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className='h-10 w-24 rounded-lg' />
                ))}
              </div>
            </div>
          </div>

          {/* Redemption Code Section Skeleton */}
          <div className='space-y-3 border-t pt-8'>
            <Skeleton className='h-3 w-24' />
            <div className='flex gap-2'>
              <Skeleton className='h-10 flex-1' />
              <Skeleton className='h-10 w-20' />
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <TitledCard
      title={t('Add Funds')}
      description={t('Choose an amount and payment method')}
      icon={<WalletCards className='h-4 w-4' />}
      action={
        onOpenBilling ? (
          <Button
            variant='outline'
            size='sm'
            onClick={onOpenBilling}
            className='w-full gap-2 sm:w-auto'
          >
            <Receipt className='h-4 w-4' />
            {t('Order History')}
          </Button>
        ) : null
      }
      contentClassName='space-y-4 sm:space-y-6'
    >
      {/* Online Topup Section */}
      {hasAnyTopup ? (
        <div className='space-y-4 sm:space-y-6'>
          {presetAmounts.length > 0 && (
            <div className='space-y-2.5 sm:space-y-3'>
              <span className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
                {t('Amount')}
              </span>
              <div className='grid grid-cols-2 gap-1.5 sm:gap-3 md:grid-cols-4'>
                {presetAmounts.map((preset, index) => {
                  const bonus =
                    preset.bonus ||
                    topupInfo?.bonus_credit?.[preset.value] ||
                    0
                  const {
                    displayValue,
                    actualPrice,
                    isBonus,
                    bonusCredit,
                  } = calculatePresetPricing(
                    preset.value,
                    priceRatio,
                    bonus,
                    usdExchangeRate
                  )
                  return (
                    <Button
                      key={index}
                      variant='outline'
                      className={cn(
                        'hover:border-foreground/50 flex min-h-16 flex-col items-start rounded-xl px-3 py-2.5 text-left whitespace-normal transition-all sm:min-h-[72px] sm:p-4',
                        selectedPreset === preset.value
                          ? 'border-accent-brand bg-accent-brand/10 text-accent-brand'
                          : 'border-muted hover:border-accent-brand/30'
                      )}
                      onClick={() => onSelectPreset(preset)}
                    >
                      <div className='flex w-full items-center justify-between'>
                        <div className='text-base font-semibold sm:text-lg'>
                          {isBonus
                            ? formatNumber(bonusCredit * usdExchangeRate)
                            : formatNumber(displayValue)}
                        </div>
                        {gift?.enabled && giftThreshold > 0 && preset.value >= giftThreshold && (
                          <div className='flex items-center gap-1 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-semibold text-amber-700'>
                            <PartyPopper className='h-3 w-3' />
                            {t('Give away {{gift}}', { gift: gift.name })}
                          </div>
                        )}
                        {preset.value !== giftThreshold && isBonus && (
                          <div className='flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-[10px] font-semibold text-green-700'>
                            +{Math.round((bonus / preset.value) * 100)}%
                          </div>
                        )}
                      </div>
                      <div className='text-muted-foreground mt-1.5 w-full text-xs sm:mt-2'>
                        {isBonus ? (
                          <>
                            Pay {formatCurrency(actualPrice)}
                            <span className='text-green-600 font-medium'>
                              {' '}• {t('Credit')}: {formatCurrency(bonusCredit - actualPrice)}
                            </span>
                          </>
                        ) : (
                          <>{formatCurrency(actualPrice)}</>
                        )}
                      </div>
                    </Button>
                  )
                })}
              </div>
            </div>
          )}

          <div className='space-y-2.5 sm:space-y-3'>
            <Label
              htmlFor='topup-amount'
              className='text-muted-foreground text-xs font-medium tracking-wider uppercase'
            >
              {t('Custom Amount')}
            </Label>
            <div className='grid grid-cols-[minmax(0,1fr)_minmax(110px,0.55fr)] gap-2 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center'>
              <Input
                id='topup-amount'
                type='number'
                value={localAmount}
                onChange={(e) => handleAmountChange(e.target.value)}
                min={minTopup}
                placeholder={`Minimum ${minTopup}`}
                className='h-9 text-base sm:h-10 sm:text-lg'
              />
              <div className='bg-muted/30 flex min-h-9 items-center justify-between gap-2 rounded-md border px-3 lg:min-w-52'>
                <span className='text-muted-foreground truncate text-xs'>
                  {t('Amount to pay:')}
                </span>
                {calculating ? (
                  <Skeleton className='h-5 w-16' />
                ) : (
                  <span className='text-sm font-semibold'>
                    {formatCurrency(paymentAmount)}
                  </span>
                )}
              </div>
            </div>
          </div>

          {gift?.enabled && giftThreshold > 0 && topupAmount >= giftThreshold && (
            <div className='flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-medium text-amber-800'>
              <PartyPopper className='h-4 w-4 shrink-0' />
              {t('This recharge qualifies for: {{gift}}', { gift: gift.name })}
            </div>
          )}

          <div className='space-y-2.5 sm:space-y-3'>
            <span className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Payment Method')}
            </span>
            {hasStandardPaymentMethods ? (
              <div className='grid grid-cols-1 gap-2 sm:grid-cols-2'>
                {topupInfo?.pay_methods?.map((method) => {
                  const minTopup = method.min_topup || 0
                  const disabled = minTopup > topupAmount
                  const isAlipay = method.type === PAYMENT_TYPES.ALIPAY
                  const isWechat = method.type === PAYMENT_TYPES.WECHAT

                  const PaymentIcon = isAlipay ? (
                    <SiAlipay
                      className='h-7 w-7'
                      style={{ color: PAYMENT_ICON_COLORS[PAYMENT_TYPES.ALIPAY] }}
                    />
                  ) : isWechat ? (
                    <SiWechat
                      className='h-7 w-7'
                      style={{ color: PAYMENT_ICON_COLORS[PAYMENT_TYPES.WECHAT] }}
                    />
                  ) : null

                  const subLabel = isWechat
                    ? t('Scan QR with WeChat')
                    : isAlipay
                      ? t('Redirect to Alipay')
                      : ''

                  const button = (
                    <Button
                      key={method.type}
                      variant='outline'
                      onClick={() => onPaymentMethodSelect(method)}
                      disabled={disabled || !!paymentLoading}
                      className={cn(
                        'flex h-auto items-center gap-3 rounded-xl border-2 px-4 py-3 transition-all',
                        selectedPaymentMethod?.type === method.type
                          ? 'border-accent-brand bg-accent-brand/10'
                          : 'border-muted hover:border-accent-brand/30',
                        disabled && 'opacity-50'
                      )}
                    >
                      {paymentLoading === method.type ? (
                        <Loader2 className='h-7 w-7 animate-spin' />
                      ) : (
                        PaymentIcon
                      )}
                      <div className='flex flex-col items-start text-left'>
                        <span className='text-sm font-semibold'>
                          {method.name}
                        </span>
                        {subLabel && (
                          <span className='text-muted-foreground text-[11px]'>
                            {subLabel}
                          </span>
                        )}
                      </div>
                      {isWechat && !disabled && (
                        <Smartphone className='text-muted-foreground ml-auto h-4 w-4' />
                      )}
                    </Button>
                  )

                  return disabled ? (
                    <TooltipProvider key={method.type}>
                      <Tooltip>
                        <TooltipTrigger render={button}></TooltipTrigger>
                        <TooltipContent>
                          {t('Minimum topup amount: {{amount}}', {
                            amount: minTopup,
                          })}
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  ) : (
                    <div key={method.type}>{button}</div>
                  )
                })}
              </div>
            ) : (
              <Alert>
                <AlertDescription>
                  {t(
                    'No payment methods available. Please contact administrator.'
                  )}
                </AlertDescription>
              </Alert>
            )}
          </div>
        </div>
      ) : (
        <Alert>
          <AlertDescription>
            {t(
              'Online topup is not enabled. Please use redemption code or contact administrator.'
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* Redemption Code Section */}
      {redemptionEnabled ? (
        <div className='space-y-2.5 border-t pt-4 sm:space-y-3 sm:pt-6'>
          <div className='flex items-center gap-2'>
            <Gift className='text-muted-foreground h-4 w-4' />
            <Label
              htmlFor='redemption-code'
              className='text-muted-foreground text-xs font-medium tracking-wider uppercase'
            >
              {t('Have a Code?')}
            </Label>
          </div>
          <div className='grid grid-cols-[minmax(0,1fr)_auto] gap-2'>
            <Input
              id='redemption-code'
              value={redemptionCode}
              onChange={(e) => onRedemptionCodeChange(e.target.value)}
              placeholder={t('Enter your redemption code')}
              className='h-9 min-w-0'
            />
            <Button
              onClick={onRedeem}
              disabled={redeeming}
              variant='outline'
              className='h-9 px-4'
            >
              {redeeming && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
              {t('Redeem')}
            </Button>
          </div>
          {topupLink && (
            <p className='text-muted-foreground text-xs'>
              {t('Need a redemption code?')}{' '}
              <a
                href={topupLink}
                target='_blank'
                rel='noopener noreferrer'
                className='inline-flex items-center gap-1 underline-offset-4 hover:underline'
              >
                {t('Get one here')}
                <ExternalLink className='h-3 w-3' />
              </a>
            </p>
          )}
        </div>
      ) : (
        <Alert className='border-t'>
          <AlertDescription>
            {t(
              'Redemption codes are disabled until the administrator confirms compliance terms.'
            )}
          </AlertDescription>
        </Alert>
      )}
    </TitledCard>
  )
}
