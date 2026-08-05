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
import { api } from '@/lib/api'
import type {
  RedemptionRequest,
  AmountRequest,
  AffiliateTransferRequest,
  ApiResponse,
  TopupInfoResponse,
  RedemptionResponse,
  AmountResponse,
  PaymentResponse,
  AffiliateCodeResponse,
  AffiliateTransferResponse,
  BillingHistoryResponse,
  CompleteOrderRequest,
  WechatPaymentRequest,
  WechatPaymentResponse,
  AlipayPaymentRequest,
} from './types'

// ============================================================================
// Wallet API Functions
// ============================================================================

/**
 * Check if API response is successful
 */
export function isApiSuccess(response: ApiResponse): boolean {
  return response.success === true || response.message === 'success'
}

/**
 * Get topup configuration info
 */
export async function getTopupInfo(): Promise<TopupInfoResponse> {
  const res = await api.get('/api/user/topup/info')
  return res.data
}

/**
 * Redeem a topup code
 */
export async function redeemTopupCode(
  request: RedemptionRequest
): Promise<RedemptionResponse> {
  const res = await api.post('/api/user/topup', request)
  return res.data
}

/**
 * Calculate payment amount for Alipay
 */
export async function calculateAlipayAmount(
  request: AmountRequest
): Promise<AmountResponse> {
  const res = await api.post('/api/user/alipay/amount', request, {
    skipBusinessError: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Request Alipay payment
 */
export async function requestAlipayPayment(
  request: AlipayPaymentRequest
): Promise<PaymentResponse> {
  const res = await api.post('/api/user/alipay/pay', request, {
    skipBusinessError: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Calculate payment amount for WeChat payment
 */
export async function calculateWechatAmount(
  request: AmountRequest
): Promise<AmountResponse> {
  const res = await api.post('/api/user/wechat/amount', request, {
    skipBusinessError: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Request WeChat payment
 */
export async function requestWechatPayment(
  request: WechatPaymentRequest
): Promise<WechatPaymentResponse> {
  const res = await api.post('/api/user/wechat/pay', request, {
    skipBusinessError: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Creem payment request payload
 */
export interface CreemPaymentRequest {
  /** Creem product identifier */
  product_id: string
  /** Payment method identifier */
  payment_method: 'creem'
}

/**
 * Creem payment response payload
 */
export type CreemPaymentResponse = ApiResponse<{ checkout_url: string }>

/**
 * Request a Creem-hosted checkout session for a product.
 */
export async function requestCreemPayment(
  request: CreemPaymentRequest
): Promise<CreemPaymentResponse> {
  const res = await api.post('/api/user/creem/pay', request, {
    skipBusinessError: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Waffo Pancake payment request payload
 */
export interface WaffoPancakePaymentRequest {
  /** Topup amount (quota) */
  amount: number
  /** Payment method identifier */
  payment_method?: 'waffo_pancake'
}

/**
 * Request a Waffo Pancake hosted-checkout session.
 */
export async function requestWaffoPancakePayment(
  request: WaffoPancakePaymentRequest
): Promise<ApiResponse> {
  const res = await api.post('/api/user/waffo_pancake/pay', request, {
    skipBusinessError: true,
  } as Record<string, unknown>)
  return res.data
}

/**
 * Get affiliate code
 */
export async function getAffiliateCode(): Promise<AffiliateCodeResponse> {
  const res = await api.get('/api/user/aff')
  return res.data
}

/**
 * Transfer affiliate quota to balance
 */
export async function transferAffiliateQuota(
  request: AffiliateTransferRequest
): Promise<AffiliateTransferResponse> {
  const res = await api.post('/api/user/aff_transfer', request)
  return res.data
}

/**
 * Get billing history for current user
 */
export async function getUserBillingHistory(
  page: number,
  pageSize: number,
  keyword?: string
): Promise<ApiResponse<BillingHistoryResponse>> {
  const params = new URLSearchParams({
    p: page.toString(),
    page_size: pageSize.toString(),
  })
  if (keyword) {
    params.append('keyword', keyword)
  }
  const res = await api.get(`/api/user/topup/self?${params.toString()}`)
  return res.data
}

// ============================================================================
// Enterprise (Tenant) Wallet API
//
// Backend resolves the tenant from the authenticated session — the client
// never sends an enterprise id. Safe against cross-tenant enumeration.
// ============================================================================

/**
 * Fetch the current user's enterprise wallet view (member quota + admin main wallet).
 */
export async function getTenantWallet(): Promise<TenantWalletResponse> {
  const res = await api.get('/api/user/tenant/wallet')
  return res.data
}

/**
 * Enterprise admin grants quota from the main wallet to a member.
 */
export async function grantTenantQuota(
  userId: number,
  quota: number
): Promise<ApiResponse> {
  const res = await api.post('/api/user/tenant/wallet/grant', {
    user_id: userId,
    quota,
  })
  return res.data
}

/**
 * Enterprise admin self-recharges the main wallet (WeChat native / Alipay).
 */
export async function requestEnterpriseTopup(
  amount: number,
  paymentMethod: 'wechat' | 'alipay'
): Promise<EnterpriseTopupResponse> {
  const res = await api.post(
    '/api/user/tenant/wallet/topup',
    {
      amount,
      payment_method: paymentMethod === 'wechat' ? 'wxpay' : 'alipay',
    },
    {
      skipBusinessError: true,
    } as Record<string, unknown>
  )
  return res.data
}

/**
 * Get billing history for all users (admin only)
 */
export async function getAllBillingHistory(
  page: number,
  pageSize: number,
  keyword?: string
): Promise<ApiResponse<BillingHistoryResponse>> {
  const params = new URLSearchParams({
    p: page.toString(),
    page_size: pageSize.toString(),
  })
  if (keyword) {
    params.append('keyword', keyword)
  }
  const res = await api.get(`/api/user/topup?${params.toString()}`)
  return res.data
}

/**
 * Complete a pending order (admin only)
 */
export async function completeOrder(
  request: CompleteOrderRequest
): Promise<ApiResponse> {
  const res = await api.post('/api/user/topup/complete', request)
  return res.data
}

// ============================================================================
// Usage Stats API
// ============================================================================

/** Response from /api/log/self/stat */
export interface LogSelfStatData {
  quota: number
  rpm: number
  tpm: number
}

export type LogSelfStatResponse = ApiResponse<LogSelfStatData>

/** Parameters for log queries */
export interface LogQueryParams {
  start_timestamp?: number
  end_timestamp?: number
  type?: number
  token_name?: string
  model_name?: string
  page?: number
  page_size?: number
}

/**
 * Get usage statistics for the current user
 */
export async function getLogsSelfStat(
  params: LogQueryParams
): Promise<LogSelfStatResponse> {
  const searchParams = new URLSearchParams()
  if (params.start_timestamp !== undefined)
    searchParams.set('start_timestamp', String(params.start_timestamp))
  if (params.end_timestamp !== undefined)
    searchParams.set('end_timestamp', String(params.end_timestamp))
  if (params.type !== undefined)
    searchParams.set('type', String(params.type))
  if (params.token_name)
    searchParams.set('token_name', params.token_name)
  if (params.model_name)
    searchParams.set('model_name', params.model_name)
  const qs = searchParams.toString()
  const res = await api.get(`/api/log/self/stat${qs ? `?${qs}` : ''}`)
  return res.data
}

/**
 * Get usage logs for the current user (paginated)
 */
export async function getUserLogs(
  params: LogQueryParams
): Promise<ApiResponse<{ items: unknown[]; total: number }>> {
  const searchParams = new URLSearchParams()
  if (params.start_timestamp !== undefined)
    searchParams.set('start_timestamp', String(params.start_timestamp))
  if (params.end_timestamp !== undefined)
    searchParams.set('end_timestamp', String(params.end_timestamp))
  if (params.type !== undefined)
    searchParams.set('type', String(params.type))
  if (params.token_name)
    searchParams.set('token_name', params.token_name)
  if (params.model_name)
    searchParams.set('model_name', params.model_name)
  if (params.page !== undefined)
    searchParams.set('p', String(params.page))
  if (params.page_size !== undefined)
    searchParams.set('page_size', String(params.page_size))
  const qs = searchParams.toString()
  const res = await api.get(`/api/log/self${qs ? `?${qs}` : ''}`)
  return res.data
}
