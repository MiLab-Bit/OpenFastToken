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
import { DEFAULT_DISCOUNT_RATE } from '../constants'

// ============================================================================
// Wallet-specific Formatting Functions
// ============================================================================

/**
 * Format currency amount that is already in local currency.
 * This is used for payment amounts that have been calculated via priceRatio.
 */
export function formatCurrency(amount: number | string): string {
  const numeric =
    typeof amount === 'number' ? amount : Number.parseFloat(String(amount))
  if (!Number.isFinite(numeric)) return '-'

  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: Math.abs(numeric) >= 1 ? 2 : 4,
  }).format(numeric)
}

/**
 * Get discount label for display (e.g., "20% OFF" or "+20%")
 * discount < 1: shows "XX% OFF" (discount)
 * discount > 1: shows "+XX%" (bonus)
 */
export function getDiscountLabel(discount: number): string {
  if (discount <= 0 || discount === DEFAULT_DISCOUNT_RATE) {
    return ''
  }
  if (discount < 1) {
    const off = Math.round((1 - discount) * 100)
    return `${off}% OFF`
  }
  // discount > 1 means bonus (e.g., 1.2 = +20%)
  const bonus = Math.round((discount - 1) * 100)
  return `+${bonus}%`
}

/**
 * Calculate pricing details for a preset amount
 *
 * - bonus > 0: 加法赠送模式 → 按面额实付 (actualPrice = 面额)，额外赠送额度累加到到账额 (bonusCredit = 面额 + 赠送)
 * - bonus == 0: 无赠送 → 按面额实付
 */
export function calculatePresetPricing(
  presetValue: number,
  priceRatio: number,
  bonus: number,
  usdExchangeRate: number = 1
) {
  const originalPrice = presetValue * priceRatio
  const hasBonus = bonus > 0
  // 加法赠送：用户按面额实付，赠送额度直接累加进到账额
  const actualPrice = originalPrice
  const bonusCredit = hasBonus
    ? Math.round((originalPrice + bonus) * 100) / 100
    : originalPrice
  const displayValue = presetValue * usdExchangeRate

  return {
    displayValue,
    originalPrice,
    actualPrice,
    savedAmount: 0,
    hasDiscount: false,
    isBonus: hasBonus,
    bonusCredit,
  }
}

/**
 * Format a Creem product price using the product's currency.
 */
export function formatCreemPrice(price: number, currency: string): string {
  const numeric = Number(price)
  if (!Number.isFinite(numeric)) return '-'
  const safeCurrency = currency || 'USD'
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: safeCurrency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(numeric)
  } catch {
    return `${safeCurrency} ${numeric.toFixed(2)}`
  }
}

/**
 * Format a large quota value into a compact, human-readable string
 * (e.g. 1500 -> "1.5K", 2_000_000 -> "2.0M").
 */
export function formatQuotaShort(quota: number): string {
  const numeric = Number(quota)
  if (!Number.isFinite(numeric)) return '-'
  if (numeric >= 1_000_000) return `${(numeric / 1_000_000).toFixed(1)}M`
  if (numeric >= 1_000) return `${(numeric / 1_000).toFixed(1)}K`
  return String(numeric)
}
