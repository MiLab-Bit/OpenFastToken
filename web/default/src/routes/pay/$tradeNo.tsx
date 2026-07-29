/*
Copyright (C) 2023-2026 OpenFastToken
*/

import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { PaymentPage } from '@/features/wallet/components/payment-page'

export const Route = createFileRoute('/pay/$tradeNo')({
  component: PayPageComponent,
})

interface PayPageSearch {
  codeUrl?: string
  amount?: string | number
  topupAmount?: string | number
  usdExchangeRate?: string | number
}

function PayPageComponent() {
  const { t } = useTranslation()
  const { tradeNo } = Route.useParams()
  const search = Route.useSearch() as PayPageSearch

  const codeUrl = (search.codeUrl as string) || ''
  const amount = Number(search.amount) || 0
  const topupAmount = Number(search.topupAmount) || 0
  const usdExchangeRate = Number(search.usdExchangeRate) || 1

  if (!codeUrl) {
    return (
      <div className='min-h-screen flex items-center justify-center bg-gray-50'>
        <div className='text-center'>
          <p className='text-gray-400 text-lg'>{t('支付信息无效')}</p>
          <p className='text-gray-300 text-sm mt-2'>{t('请返回钱包页面重新发起支付')}</p>
        </div>
      </div>
    )
  }

  return (
    <PaymentPage
      tradeNo={tradeNo}
      codeUrl={codeUrl}
      amount={amount}
      topupAmount={topupAmount}
      usdExchangeRate={usdExchangeRate}
    />
  )
}
