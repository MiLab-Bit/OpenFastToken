/*
Copyright (C) 2023-2026 智企惠

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

For commercial licensing, please contact abovetigers@qq.com
*/
import { formatTimestampToDate } from '@/lib/format'
import type { StatusBadgeProps } from '@/components/status-badge'
import type { TopupStatus, EnterpriseWalletTxnType } from '../types'

// ============================================================================
// Billing Utility Functions
// ============================================================================

interface StatusConfig {
  variant: StatusBadgeProps['variant']
  label: string
}

/**
 * Status badge configuration
 */
export const STATUS_CONFIG: Record<TopupStatus, StatusConfig> = {
  success: {
    variant: 'success',
    label: 'Success',
  },
  pending: {
    variant: 'warning',
    label: 'Pending',
  },
  expired: {
    variant: 'danger',
    label: 'Expired',
  },
}

/**
 * Get status badge configuration
 */
export function getStatusConfig(status: TopupStatus): StatusConfig {
  return STATUS_CONFIG[status] || STATUS_CONFIG.pending
}

/**
 * Enterprise wallet transaction type badge configuration.
 * Variant mapping follows the project's semantic token set (no `secondary`
 * variant exists — `grant` uses `neutral`). Income types (recharge / refund /
 * recycle) read as gains, expense types (grant / consume) as deductions.
 */
export const TXN_TYPE_CONFIG: Record<EnterpriseWalletTxnType, StatusConfig> = {
  recharge: { variant: 'success', label: 'Enterprise Recharge' },
  refund: { variant: 'success', label: 'Refund' },
  recycle: { variant: 'info', label: 'Recycle from Member' },
  grant: { variant: 'neutral', label: 'Grant to Member' },
  consume: { variant: 'warning', label: 'Member Consume' },
}

/**
 * Get the badge variant + i18n label key for an enterprise wallet txn type.
 */
export function getTxnTypeConfig(type: EnterpriseWalletTxnType): StatusConfig {
  return TXN_TYPE_CONFIG[type] || TXN_TYPE_CONFIG.recharge
}

/**
 * Payment method display names
 */
export const PAYMENT_METHOD_NAMES: Record<string, string> = {
  alipay: 'Alipay',
  wxpay: 'WeChat Pay',
}

/**
 * Get payment method display name
 */
export function getPaymentMethodName(
  method: string,
  t?: (key: string) => string
): string {
  const name = PAYMENT_METHOD_NAMES[method] || method
  return t ? t(name) : name
}

/**
 * Format timestamp to readable date string
 */
export function formatTimestamp(timestamp: number): string {
  return formatTimestampToDate(timestamp)
}
