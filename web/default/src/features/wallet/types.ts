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
// ============================================================================
// Wallet Type Definitions
// ============================================================================

/**
 * Generic API response
 */
export interface ApiResponse<T = unknown> {
  success?: boolean
  message?: string
  data?: T
}

/**
 * Standard API response types
 */
export type TopupInfoResponse = ApiResponse<TopupInfo>
export type RedemptionResponse = ApiResponse<number>
export type AmountResponse = ApiResponse<string>
export type PaymentResponse = ApiResponse<Record<string, unknown>> & {
  url?: string
}
export type AffiliateCodeResponse = ApiResponse<string>
export type AffiliateTransferResponse = ApiResponse
export type WechatPaymentResponse = ApiResponse<{ code_url: string; trade_no: string }>

/**
 * Wechat payment request
 */
export interface WechatPaymentRequest {
  /** Topup amount */
  amount: number
  /** Payment method identifier */
  payment_method: 'wxpay'
  /** Optional return URL */
  return_url?: string
  /** Optional payment method index used by the gateway */
  pay_method_index?: number
}

/**
 * Alipay payment request
 */
export interface AlipayPaymentRequest {
  /** Topup amount */
  amount: number
  /** Payment method identifier */
  payment_method: 'alipay'
  /** Optional return URL */
  return_url?: string
}

/**
 * Payment method configuration
 */
export interface PaymentMethod {
  /** Display name of payment method */
  name: string
  /** Payment method type identifier */
  type: string
  /** Optional color for UI display */
  color?: string
  /** Minimum topup amount for this payment method */
  min_topup?: number
  /** Optional icon URL provided by backend (preferred over built-in icons) */
  icon?: string
}

/**
 * Topup configuration information
 */
export interface TopupInfo {
  /** Available payment methods */
  pay_methods: PaymentMethod[]
  /** Minimum topup amount for online topup */
  min_topup: number
  /** Preset amount options */
  amount_options: number[]
  /** Extra credited quota (additive bonus) by preset amount, e.g. {100: 20} */
  bonus_credit?: Record<number, number>
  /** Optional topup link for purchasing codes */
  topup_link?: string
  /** Whether redemption code usage is enabled */
  enable_redemption?: boolean
  /** Whether compliance confirmation has been completed */
  payment_compliance_confirmed?: boolean
  /** Current compliance terms version */
  payment_compliance_terms_version?: string
  /** Whether Alipay topup is enabled */
  enable_alipay_topup?: boolean
  /** Minimum topup amount for Alipay */
  alipay_min_topup?: number
  /** Whether WeChat topup is enabled */
  enable_wechat_topup?: boolean
  /** Minimum topup amount for WeChat */
  wechat_min_topup?: number
  /** Whether online topup is enabled */
  enable_online_topup?: boolean
  /** Recharge gift config (bonus tiers + threshold gift) */
  recharge_gift?: RechargeGiftInfo
}

/**
 * Recharge gift tier: recharge >= amount (in the same units as preset values)
 * grants an extra bonus_rate of equivalent quota.
 */
export interface RechargeGiftTierInfo {
  /** Tier threshold amount (same units as preset values) */
  amount: number
  /** Bonus rate, 0.2 = extra 20% equivalent quota */
  bonus_rate: number
}

/**
 * Threshold gift (e.g. bonus gift) granted when a single recharge meets the threshold.
 */
export interface RechargeGiftGiftInfo {
  /** Whether the threshold gift is enabled */
  enabled: boolean
  /** Threshold amount (same units as preset values) */
  threshold: number
  /** Gift name, e.g. "额外配额赠送" */
  name: string
  /** Gift type key, e.g. "gift_type" */
  type: string
}

/**
 * Recharge gift configuration returned by GetTopUpInfo.
 */
export interface RechargeGiftInfo {
  /** Whether recharge gift is enabled */
  enabled: boolean
  /** Bonus tiers by amount */
  tiers: RechargeGiftTierInfo[]
  /** Threshold gift */
  gift: RechargeGiftGiftInfo
}

/**
 * Preset amount option with optional additive bonus
 */
export interface PresetAmount {
  /** Preset amount value */
  value: number
  /** Optional additive bonus credit (0 = no bonus) */
  bonus?: number
}

/**
 * Redemption code request
 */
export interface RedemptionRequest {
  /** Redemption code key */
  key: string
}

/**
 * Payment request parameters
 */
export interface PaymentRequest {
  /** Topup amount */
  amount: number
  /** Payment method identifier */
  payment_method: string
}

/**
 * Amount calculation request
 */
export interface AmountRequest {
  /** Topup amount to calculate */
  amount: number
}

/**
 * Affiliate quota transfer request
 */
export interface AffiliateTransferRequest {
  /** Quota amount to transfer */
  quota: number
}

/**
 * User wallet data
 */
export interface UserWalletData {
  /** User ID */
  id: number
  /** Username */
  username: string
  /** Current quota balance */
  quota: number
  /** Total used quota */
  used_quota: number
  /** Total request count */
  request_count: number
  /** Affiliate quota (pending rewards) */
  aff_quota: number
  /** Total affiliate quota earned (historical) */
  aff_history_quota: number
  /** Number of successful affiliate invites */
  aff_count: number
  /** Cumulative referred-user actual payment amount (for referral tier calculation) */
  aff_recharge_total: number
  /** User group */
  group: string
}

/**
 * Topup record status
 */
export type TopupStatus = 'success' | 'pending' | 'expired'

/**
 * Topup billing record
 */
export interface TopupRecord {
  /** Record ID */
  id: number
  /** User ID */
  user_id: number
  /** Topup amount (quota) */
  amount: number
  /** Payment amount (actual money paid) */
  money: number
  /** Trade/order number */
  trade_no: string
  /** Payment method type */
  payment_method: string
  /** Creation timestamp */
  create_time: number
  /** Completion timestamp */
  complete_time?: number
  /** Payment status */
  status: TopupStatus
}

/**
 * Billing history response
 */
export interface BillingHistoryResponse {
  items: TopupRecord[]
  total: number
}

/**
 * Complete order request (admin only)
 */
export interface CompleteOrderRequest {
  trade_no: string
}

/**
 * Creem product configuration used for user recharge options.
 */
export interface CreemProduct {
  /** Display name of the product */
  name: string
  /** Creem product identifier */
  productId: string
  /** Price in the product currency */
  price: number
  /** Granted quota amount */
  quota: number
  /** ISO currency code, e.g. "USD" */
  currency: 'USD' | 'EUR'
}
